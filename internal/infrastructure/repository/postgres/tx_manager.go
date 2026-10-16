package postgres

import (
	"context"

	"go-gin-clean/internal/application/port"

	"gorm.io/gorm"
)

type txContextKey struct{}

var activeTxKey = txContextKey{}

type gormTxManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) port.TransactionManager {
	return &gormTxManager{db: db}
}

func (m *gormTxManager) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	if _, ok := ctx.Value(activeTxKey).(*gorm.DB); ok {
		return fn(ctx)
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, activeTxKey, tx)
		return fn(txCtx)
	})
}

func getTx(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(activeTxKey).(*gorm.DB); ok {
		return tx
	}
	return fallback.WithContext(ctx)
}
