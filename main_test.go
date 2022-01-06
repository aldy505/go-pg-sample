package main_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	dandelion "dandelion"

	_ "github.com/lib/pq"
)

var db *sql.DB

func TestMain(m *testing.M) {
	dbURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		dbURL = "postgres://postgres:gX886f8Gs88DsQYjqhNZ4@localhost:5432/dandelion?sslmode=disable"
	}

	var err error = nil

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("opening postgres connection: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err = (&dandelion.Dependency{DB: db}).Migrate(ctx)
	if err != nil {
		log.Printf("migrating db: %v", err)
	}

	os.Exit(m.Run())
}

func cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	c, err := db.Conn(ctx)
	if err != nil {
		log.Fatalf("acquiring connection pool: %v", err)
	}
	defer c.Close()

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		tx.Rollback()
		log.Fatalf("creating transaction: %v", err)
	}

	_, err = tx.ExecContext(ctx, `TRUNCATE TABLE posts RESTART IDENTITY CASCADE`)
	if err != nil {
		tx.Rollback()
		log.Fatalf("err executing query: %v", err)
	}

	_, err = tx.ExecContext(ctx, `TRUNCATE TABLE comment RESTART IDENTITY CASCADE`)
	if err != nil {
		tx.Rollback()
		log.Fatalf("err executing query: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("commiting transaction: %v", err)
	}
}
