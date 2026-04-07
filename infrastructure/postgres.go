package infrastructure

import (
	"context"
	"time"

	"auction-service/global"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPostgresSqlDB(postgresConfig global.PostgresConfig) (*sqlx.DB, *pgxpool.Pool) {
	pgxPoolConfig, err := pgxpool.ParseConfig("")
	if err != nil {
		panic(err)
	}

	pgxPoolConfig.MaxConns = 20
	pgxPoolConfig.MinConns = 3
	pgxPoolConfig.MaxConnLifetime = time.Minute * 30
	pgxPoolConfig.MaxConnIdleTime = time.Minute * 10

	pgxConnConfig := pgxPoolConfig.ConnConfig
	postgresConnectionConfig := &pgxConnConfig.Config
	postgresConnectionConfig.Host = postgresConfig.Host
	postgresConnectionConfig.Port = postgresConfig.Port
	postgresConnectionConfig.Database = postgresConfig.Database
	postgresConnectionConfig.User = postgresConfig.Username
	postgresConnectionConfig.Password = postgresConfig.Password
	postgresConnectionConfig.RuntimeParams["timezone"] = "UTC"

	dbPool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig)
	if err != nil {
		panic(err)
	}

	pgxDB := stdlib.OpenDBFromPool(dbPool)
	if err = pgxDB.Ping(); err != nil {
		pgxDB.Close()
		panic(err)
	}

	db := sqlx.NewDb(pgxDB, "pgx")

	return db, dbPool
}
