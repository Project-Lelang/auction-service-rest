package infrastructure

import (
	"context"
	"fmt"
	"log"

	"auction-service/global"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	storage_go "github.com/supabase-community/storage-go"
)

type infrastructureManager struct {
	sqlDB                 *sqlx.DB
	supabaseStorageClient *storage_go.Client
	midtransClient        MidtransClient
	biteshipClient        BiteshipClient
	taskQueueClient       TaskQueueClient
	notificationQueue     NotificationQueueClient
	pushClient            PushClient
	loggerStack           LoggerStack
}

func NewInfrastructureManager(configuration global.YamlConfig) InfrastructureManager {
	sqlDB := NewMysqlDB(configuration.Mysql)
	supabaseClient := NewSupabaseStorageClient(configuration.Supabase)

	var midtransClient MidtransClient
	if configuration.Midtrans != nil {
		midtransClient = NewMidtransClient(configuration.Midtrans.ServerKey, configuration.Midtrans.IsSandbox)
	}

	var biteshipClient BiteshipClient
	if configuration.Biteship != nil {
		biteshipClient = NewBiteshipClient(configuration.Biteship.ApiKey)
	}

	pushClient, err := NewFirebasePushClient(context.Background(), configuration.Firebase)
	if err != nil {
		log.Printf("[notification worker] firebase disabled: %v", err)
		pushClient = NewNoopPushClient()
	}

	return &infrastructureManager{
		sqlDB:                 sqlDB,
		supabaseStorageClient: supabaseClient,
		midtransClient:        midtransClient,
		biteshipClient:        biteshipClient,
		taskQueueClient:       NewTaskQueueClient(configuration.Redis),
		notificationQueue:     NewNotificationQueueClient(configuration.Redis, configuration.Notification),
		pushClient:            pushClient,
		loggerStack:           NewLoggerStack(configuration.LogChannel),
	}
}

func (i *infrastructureManager) GetDB() *sqlx.DB {
	return i.sqlDB
}

func (i *infrastructureManager) GetSupabaseStorageClient() *storage_go.Client {
	return i.supabaseStorageClient
}

func (i *infrastructureManager) GetMidtransClient() MidtransClient {
	return i.midtransClient
}

func (i *infrastructureManager) GetBiteshipClient() BiteshipClient {
	return i.biteshipClient
}

func (i *infrastructureManager) GetTaskQueueClient() TaskQueueClient {
	return i.taskQueueClient
}

func (i *infrastructureManager) GetNotificationQueueClient() NotificationQueueClient {
	return i.notificationQueue
}

func (i *infrastructureManager) GetPushClient() PushClient {
	return i.pushClient
}

func (i *infrastructureManager) migrationDSN() string {
	cfg := global.GetMysqlConfig()
	return fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

func (i *infrastructureManager) MigrateDB(isRollingBack bool, steps int, force *int) error {
	m, err := migrate.New("file://./migration", i.migrationDSN())
	if err != nil {
		return err
	}
	defer m.Close()

	if force != nil {
		return m.Force(*force)
	}

	if isRollingBack {
		if steps > 0 {
			err = m.Steps(-1 * steps)
		} else {
			err = m.Down()
		}
	} else {
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
	}

	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (i *infrastructureManager) RefreshDB() error {
	// Fresh migration intentionally discards the entire schema. Do not execute
	// every down migration here: historical down migrations may be unable to
	// represent current data (for example, converting email values back into a
	// short phone column) and can fail before the schema is cleared.
	m, err := migrate.New("file://./migration", i.migrationDSN())
	if err != nil {
		return err
	}
	if err = m.Drop(); err != nil && err != migrate.ErrNoChange {
		_, _ = m.Close()
		return err
	}
	if sourceErr, databaseErr := m.Close(); sourceErr != nil {
		return sourceErr
	} else if databaseErr != nil {
		return databaseErr
	}

	if err := i.MigrateDB(false, 0, nil); err != nil {
		return err
	}

	// A fresh database reuses numeric IDs from the beginning. Every existing
	// Redis task and notification may therefore point at an unrelated new row,
	// and fixed Asynq task IDs would conflict with newly scheduled work.
	if err := i.taskQueueClient.Reset(); err != nil {
		return fmt.Errorf("reset task queue after fresh migration: %w", err)
	}
	if err := i.notificationQueue.Reset(context.Background()); err != nil {
		return fmt.Errorf("reset notification queue after fresh migration: %w", err)
	}

	return nil
}

func (i *infrastructureManager) CloseDB() error {
	if i.sqlDB != nil {
		i.sqlDB.Close()
	}
	return nil
}

func (i *infrastructureManager) GetLoggerStack() LoggerStack {
	return i.loggerStack
}

func (i *infrastructureManager) Close() error {
	if i.taskQueueClient != nil {
		_ = i.taskQueueClient.Close()
	}
	if i.notificationQueue != nil {
		_ = i.notificationQueue.Close()
	}
	return i.CloseDB()
}
