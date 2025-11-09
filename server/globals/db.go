package globals

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	db      *pgxpool.Pool
	mongoDb *mongo.Database
)

func Db() *pgxpool.Pool {
	if db != nil {
		return db
	}
	var err error
	conn, err := pgxpool.New(context.Background(), os.Getenv("POSTGRES_URI"))
	if err != nil {
		log.Fatal(err)
	}
	db = conn
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

// creates a lock in postgres to prevent concurrent access to the same token
// set key to something stable, like accountId+userId
func PostgresLock(ctx context.Context, conn *pgx.Conn, key int64, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	// Attempt non-blocking first (fast path)
	var got bool
	err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, key).Scan(&got)
	if err != nil {
		return err
	}
	if got {
		return nil
	}
	// Block (with context timeout) until lock is available
	// pg_advisory_lock is blocking; we rely on the context timeout
	_, err = conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, key)
	if err != nil {
		return err
	}
	return nil
}

func PostgresUnlock(ctx context.Context, conn *pgx.Conn, key int64) {
	_, _ = conn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, key)
}
