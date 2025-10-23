package globals

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

func Open() {
	var err error
	db, err = sql.Open("sqlite3", "./desktop.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
}

func Db() *sql.DB {
	return db
}
