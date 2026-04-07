package infrastructure

import (
	"fmt"

	"auction-service/global"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
)

type infrastructureManager struct {
	sqlDB       *sqlx.DB
	pgxPool     *pgxpool.Pool
	loggerStack LoggerStack
}

func NewInfrastructureManager(configuration global.YamlConfig) InfrastructureManager {
	sqlDB, pgxPool := NewPostgresSqlDB(configuration.Postgres)

	return &infrastructureManager{
		sqlDB:       sqlDB,
		pgxPool:     pgxPool,
		loggerStack: NewLoggerStack(configuration.LogChannel),
	}
}

func (i *infrastructureManager) GetDB() *sqlx.DB {
	return i.sqlDB
}

func (i *infrastructureManager) migrationDSN() string {
	pg := global.GetPostgresConfig()
	return fmt.Sprintf("pgx5://%s:%s@%s:%d/%s?sslmode=disable",
		pg.Username, pg.Password, pg.Host, pg.Port, pg.Database)
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
	if i.pgxPool != nil {
		i.pgxPool.Close()
	}
	return nil
}

func (i *infrastructureManager) GetLoggerStack() LoggerStack {
	return i.loggerStack
}

func (i *infrastructureManager) Close() error {
	return i.CloseDB()
}
