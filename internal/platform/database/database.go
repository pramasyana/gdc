// Package database provides the MySQL (GORM) connection and transaction
// manager used by the repository layer.
package database

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/ryanpramasyana/gdc/internal/platform/config"
)

// New opens a MySQL connection and configures the connection pool.
func New(cfg config.DB, env string) (*gorm.DB, error) {
	level := gormlogger.Warn
	if env != "production" {
		level = gormlogger.Info
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(level),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: sql handle: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database: ping: %w", err)
	}
	return db, nil
}

// Ping verifies connectivity to the database.
func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// txKey is the context key under which an active transaction is stored.
type txKey struct{}

// TxManager runs work inside a GORM transaction and injects the transaction
// handle into the context for repositories to pick up.
type TxManager struct {
	db *gorm.DB
}

// NewTxManager builds a TxManager around the given database handle.
func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

// WithinTransaction implements domain.TxManager.
func (m *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// FromContext returns the transaction stored in the context, or the provided
// fallback connection when no transaction is active.
func FromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}
