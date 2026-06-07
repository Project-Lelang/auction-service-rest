package infrastructure

import (
	"fmt"

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

	return &infrastructureManager{
		sqlDB:                 sqlDB,
		supabaseStorageClient: supabaseClient,
		midtransClient:        midtransClient,
		biteshipClient:        biteshipClient,
		taskQueueClient:       NewTaskQueueClient(configuration.Redis),
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
	if err := i.MigrateDB(true, 0, nil); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return i.MigrateDB(false, 0, nil)
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
	return i.CloseDB()
}
