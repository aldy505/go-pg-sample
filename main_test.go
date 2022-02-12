package main_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	dandelion "dandelion"

	_ "github.com/lib/pq"

	"github.com/ory/dockertest/v3"
)

var deps *dandelion.Dependency

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	
	resource, err := pool.Run("postgres", "14-alpine", []string{"POSTGRES_PASSWORD=gX886f8Gs88DsQYjqhNZ4", "POSTGRES_USER=postgres", "POSTGRES_DB=dandelion"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	var db *sql.DB

	if err := pool.Retry(func() error {
		dbURL, ok := os.LookupEnv("DATABASE_URL")
		if !ok {
			dbURL = "postgres://postgres:gX886f8Gs88DsQYjqhNZ4@localhost:%s/dandelion?sslmode=disable"
		}
		
		var err error
		db, err = sql.Open("postgres", fmt.Sprintf(dbURL, resource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	deps = &dandelion.Dependency{DB: db}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = deps.Migrate(ctx)
	if err != nil {
		log.Printf("migrating db: %v", err)
	}

	code := m.Run()

	db.Close()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	c, err := deps.DB.Conn(ctx)
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
		tx.Rollback()
		log.Fatalf("commiting transaction: %v", err)
	}
}
