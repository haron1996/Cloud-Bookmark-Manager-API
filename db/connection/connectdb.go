package connection

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func ConnectDB() *sql.DB {
	// config, err := util.LoadConfig(".")
	// if err != nil {
	// 	log.Println(err.Error())
	// 	return nil
	// }

	db, err := sql.Open("pgx", os.Getenv("dbString"))
	if err != nil {
		log.Println(err)
		return nil
	}

	err = db.Ping()
	if err != nil {
		log.Println("db connection closed")
		return nil
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(30 * time.Second)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db
}
