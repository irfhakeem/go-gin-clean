package postgres

import (
	"context"

	"gorm.io/gorm"
)

// TableNamer is implemented by every GORM entity so the generic
// BaseRepository can resolve table names at runtime.
type TableNamer interface {
	TableName() string
}

// BaseRepository offers generic GORM helpers keyed on a table name.
// Held (and passed around) as *BaseRepository[T] to avoid copying it.
type BaseRepository[T TableNamer] struct {
	db *gorm.DB
}

func NewBaseRepository[T TableNamer](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{
		db: db,
	}
}

func (r *BaseRepository[T]) getDB(ctx context.Context) *gorm.DB {
	return getTx(ctx, r.db)
}

func (r *BaseRepository[T]) Raw(ctx context.Context, query string) ([]*T, error) {
	var entities []*T
	if err := r.getDB(ctx).Raw(query).Scan(&entities).Error; err != nil {
		return nil, err
	}

	return entities, nil
}

func (r *BaseRepository[T]) FindAll(ctx context.Context, limit, offset int, query any, args ...any) ([]*T, int64, error) {
	var entities []*T
	var count int64

	var entity T
	q := r.getDB(ctx).Model(&entity)

	q = q.Where(query, args...)

	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Limit(limit).Offset(offset).Order("id asc").Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, count, nil
}

func (r *BaseRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	var entity T

	q := r.getDB(ctx)

	if err := q.Where("id = ?::uuid", id).Take(&entity).Error; err != nil {
		return nil, err
	}

	return &entity, nil
}

func (r *BaseRepository[T]) FindFirst(ctx context.Context, query any, args ...any) (*T, error) {
	var entity T

	q := r.getDB(ctx)

	if err := q.Where(query, args...).First(&entity).Error; err != nil {
		return nil, err
	}

	return &entity, nil
}

func (r *BaseRepository[T]) Where(ctx context.Context, query any, args ...any) ([]*T, error) {
	var entities []*T

	q := r.getDB(ctx)

	if err := q.Where(query, args...).Order("id asc").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *BaseRepository[T]) WhereExisting(ctx context.Context, query any, args ...any) (bool, error) {
	var entity T

	q := r.getDB(ctx)

	err := q.Where(query, args...).First(&entity).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	tableName := (*entity).TableName()
	if err := r.getDB(ctx).Table(tableName).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *BaseRepository[T]) BulkCreate(ctx context.Context, entities []*T) ([]*T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	tableName := (*entities[0]).TableName()
	if err := r.getDB(ctx).Table(tableName).CreateInBatches(entities, 100).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T, id string) (*T, error) {
	if err := r.getDB(ctx).Model(entity).Where("id = ?::uuid", id).Updates(entity).Error; err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(new(T), "id = ?::uuid", id).Error
}

func (r *BaseRepository[T]) BulkDelete(ctx context.Context, ids []any) error {
	if len(ids) == 0 {
		return nil
	}
	return r.getDB(ctx).Where("id IN ?", ids).Delete(new(T)).Error
}

func (r *BaseRepository[T]) SoftDelete(ctx context.Context, id string) error {
	return r.getDB(ctx).Model(new(T)).Where("id = ?::uuid", id).Update("is_deleted", true).Error
}

func SanitizeSortColumn(sortBy, sort string, sortColums map[string]string) (string, string) {
	sortColumn, ok := sortColums[sortBy]
	if !ok {
		sortColumn = "created_at"
	}

	sortOrder := "desc"
	if sort == "asc" {
		sortOrder = "asc"
	}

	return sortColumn, sortOrder
}
