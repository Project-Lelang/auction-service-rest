package infrastructure

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	storage_go "github.com/supabase-community/storage-go"
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

	// supabase storage
	GetSupabaseStorageClient() *storage_go.Client

	// midtrans
	GetMidtransClient() MidtransClient

	// biteship
	GetBiteshipClient() BiteshipClient

	// task queue
	GetTaskQueueClient() TaskQueueClient

	// notifications
	GetNotificationQueueClient() NotificationQueueClient
	GetPushClient() PushClient

	// logger
	GetLoggerStack() LoggerStack

	// lifecycle
	Close() error
}
