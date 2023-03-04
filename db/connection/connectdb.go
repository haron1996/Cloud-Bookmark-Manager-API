package connection

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func ConnectDB() *sql.DB {
	config, err := util.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("pgx", config.DBString)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(30 * time.Second)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db
}
