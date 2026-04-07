package infrastructure

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type DBTX interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, args interface{}) (sql.Result, error)
}

type InfrastructureManager interface {
	// database
	GetDB() *sqlx.DB
	MigrateDB(isRollingBack bool, steps int, force *int) error
	RefreshDB() error
	CloseDB() error

	// logger
	GetLoggerStack() LoggerStack
}
