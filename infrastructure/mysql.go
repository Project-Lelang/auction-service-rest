package infrastructure

import (
	"fmt"
	"time"

	"auction-service/global"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func NewMysqlDB(cfg global.MysqlConfig) *sqlx.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC&time_zone=UTC&charset=utf8mb4",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(time.Minute * 30)
	db.SetConnMaxIdleTime(time.Minute * 10)

	if err = db.Ping(); err != nil {
		db.Close()
		panic(err)
	}

	return db
}
