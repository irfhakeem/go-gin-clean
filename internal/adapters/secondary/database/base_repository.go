package database

import (
	"context"
	"go-gin-clean/internal/core/ports"

	"gorm.io/gorm"
)

type BaseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) ports.BaseRepository[T] {
	return &BaseRepository[T]{db: db}
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

	db := r.db.WithContext(ctx)

	if query != nil {
		db = db.Where(query, args...)
	}

	if err := db.Model(new(T)).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("id asc").Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, count, nil
}

func (r *BaseRepository[T]) FindByID(ctx context.Context, id int64) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepository[T]) FindFirst(ctx context.Context, query any, args ...any) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).Where(query, args...).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepository[T]) Where(ctx context.Context, query any, args ...any) ([]*T, error) {
	var entities []*T
	if err := r.db.WithContext(ctx).Where(query, args...).Order("id asc").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *BaseRepository[T]) WhereExisting(ctx context.Context, query any, args ...any) (bool, error) {
	var entity T
	err := r.db.WithContext(ctx).Where(query, args...).First(&entity).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}

	return entity, nil
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) (*T, error) {
	if err := r.db.WithContext(ctx).Updates(&entity).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).First(entity).Error; err != nil {
		return nil, err
	}

	return entity, nil
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(new(T), "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}
