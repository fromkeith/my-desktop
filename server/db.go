package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

func open() {
	var err error
	db, err = sql.Open("sqlite3", "./desktop.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
}
