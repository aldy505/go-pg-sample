package main_test

import (
	"context"
	dandelion "dandelion"
	"database/sql"
	"testing"
	"time"
)

func TestAddPost(t *testing.T) {
	t.Cleanup(cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	post := dandelion.Post{
		Title: "Lorem ipsum",
		Body:  "The quick brown fox jumps over the lazy dog",
	}

	_, err := deps.AddPost(ctx, post)
	if err != nil {
		t.Errorf("an error was thrown: %v", err)
	}
}

func TestGetPosts(t *testing.T) {
	t.Cleanup(cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	c, err := deps.DB.Conn(ctx)
	if err != nil {
		t.Errorf("opening connection pool: %v", err)
	}
	defer c.Close()

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		t.Errorf("opening transaction: %v", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO posts
			(id, title, body)
			VALUES
			($1, $2, $3),
			($4, $5, $6)`,
		1, "Sunken ship", "There is a ship called Titanic",
		2, "Fallen angels", "The sky is clear today..",
	)
	if err != nil {
		tx.Rollback()
		t.Errorf("executing query: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		t.Error(err)
	}

	posts, err := deps.GetPosts(ctx)
	if err != nil {
		t.Errorf("an error was thrown: %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("expecting length of post to be 2, instead got: %d", len(posts))
	}
}

func TestGetPostById(t *testing.T) {
	t.Cleanup(cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	c, err := deps.DB.Conn(ctx)
	if err != nil {
		t.Errorf("opening connection pool: %v", err)
	}
	defer c.Close()

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		t.Errorf("opening transaction: %v", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO posts
			(id, title, body)
			VALUES
			($1, $2, $3),
			($4, $5, $6)`,
		1, "Sunken ship", "There is a ship called Titanic",
		2, "Fallen angels", "The sky is clear today..",
	)
	if err != nil {
		tx.Rollback()
		t.Errorf("executing query: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		t.Error(err)
	}

	post, err := deps.GetPostById(ctx, 1)
	if err != nil {
		t.Errorf("an error was thrown: %v", err)
	}

	if post.Title != "Sunken ship" {
		t.Errorf("expecting post.Title to be 'Sunken ship' instead got: %s", post.Title)
	}
}
