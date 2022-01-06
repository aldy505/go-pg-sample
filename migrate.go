package main

import (
	"context"
	"database/sql"
)

func (d *Dependency) Migrate(ctx context.Context) error {
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`DROP TABLE IF EXISTS posts`); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`DROP TABLE IF EXISTS comment`); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			body TEXT NOT NULL
		);`); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS comment (
			id SERIAL PRIMARY KEY,
			post_id INTEGER NOT NULL,
			body TEXT NOT NULL
		);`); err != nil {
		tx.Rollback()
		return err
	}

	// create index on comment.post_id
	if _, err := tx.ExecContext(ctx,
		`CREATE INDEX IF NOT EXISTS comment_post_id_idx ON comment (post_id);`); err != nil {
		tx.Rollback()
		return err
	}

	// add foreign key for comment.post_id
	if _, err := tx.ExecContext(ctx,
		`ALTER TABLE comment
		ADD CONSTRAINT fk_comment_post_id
		FOREIGN KEY (post_id)
		REFERENCES posts(id);`); err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
