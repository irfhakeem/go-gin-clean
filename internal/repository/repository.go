package repository

import (
	"context"

	"gorm.io/gorm"
)

type TableNamer interface {
	TableName() string
}

type BaseRepository[T TableNamer] struct {
	db *gorm.DB
}

func NewBaseRepository[T TableNamer](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{
		db: db,
	}
}

func (r *BaseRepository[T]) Raw(ctx context.Context, query string) ([]*T, error) {
	var entities []*T
	if err := r.db.WithContext(ctx).Raw(query).Scan(&entities).Error; err != nil {
		return nil, err
	}

	return entities, nil
}

func (r *BaseRepository[T]) FindAll(ctx context.Context, limit, offset int, query any, args ...any) ([]*T, int64, error) {
	var entities []*T
	var count int64

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var entity T
	q := r.db.WithContext(ctx).Model(&entity)

	q = q.Where(query, args...)

	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Limit(limit).Offset(offset).Order("pkid asc").Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, count, nil
}

func (r *BaseRepository[T]) FindByID(ctx context.Context, id any) (*T, error) {
	var entity T

	q := r.db.WithContext(ctx)

	if err := q.Where("pkid = ?", id).Take(&entity).Error; err != nil {
		return nil, err
	}

	return &entity, nil
}

func (r *BaseRepository[T]) FindFirst(ctx context.Context, query any, args ...any) (*T, error) {
	var entity T

	q := r.db.WithContext(ctx)

	if err := q.Where(query, args...).First(&entity).Error; err != nil {
		return nil, err
	}

	return &entity, nil
}

func (r *BaseRepository[T]) Where(ctx context.Context, query any, args ...any) ([]*T, error) {
	var entities []*T

	q := r.db.WithContext(ctx)

	if err := q.Where(query, args...).Order("pkid asc").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *BaseRepository[T]) WhereExisting(ctx context.Context, query any, args ...any) (bool, error) {
	var entity T

	q := r.db.WithContext(ctx)

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
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	tableName := (*entity).TableName()
	if err := tx.Table(tableName).Create(entity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return entity, nil
}

func (r *BaseRepository[T]) BulkCreate(ctx context.Context, entities []*T) ([]*T, error) {
	if len(entities) == 0 {
		return entities, nil
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	tableName := (*entities[0]).TableName()
	if err := tx.Table(tableName).CreateInBatches(entities, 100).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return entities, nil
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T, pkid any) (*T, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(entity).Where("pkid = ?", pkid).Updates(entity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	result, err := r.FindByID(ctx, pkid)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *BaseRepository[T]) Delete(ctx context.Context, pkid any) error {
	if err := r.db.WithContext(ctx).Delete(new(T), "pkid = ?", pkid).Error; err != nil {
		return err
	}

	return nil
}

func (r *BaseRepository[T]) BulkDelete(ctx context.Context, pkids []any) error {
	if len(pkids) == 0 {
		return nil
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("pkid IN ?", pkids).Delete(new(T)).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *BaseRepository[T]) SoftDelete(ctx context.Context, pkid any) error {
	if err := r.db.WithContext(ctx).Model(new(T)).Where("pkid = ?", pkid).Update("is_deleted", true).Error; err != nil {
		return err
	}

	return nil
}
