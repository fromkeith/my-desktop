package globals

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	db      *sql.DB
	mongoDb *mongo.Database
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

func CloseAll() {
	if db != nil {
		db.Close()
		db = nil
	}
	if mongoDb != nil {
		mongoDb.Client().Disconnect(context.Background())
		mongoDb = nil
	}
	if kafkaConn != nil {
		kafkaConn.Close()
		kafkaConn = nil
	}
}

func DocDb() *mongo.Database {
	if mongoDb != nil {
		return mongoDb
	}
	uri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	mongoDb = client.Database(os.Getenv("MONGODB_DB"))
	return mongoDb
}
