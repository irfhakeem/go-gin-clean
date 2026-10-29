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


func (r *BaseRepository[T]) Update(ctx context.Context, entity *T, id string) (*T, error) {
	if err := r.getDB(ctx).Model(entity).Where("id = ?::uuid", id).Updates(entity).Error; err != nil {
		return nil, err
	}
	return r.FindFirst(ctx, "id = ?::uuid", id)
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(new(T), "id = ?::uuid", id).Error
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
