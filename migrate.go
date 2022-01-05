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

	if _, err := tx.Exec(`
    CREATE TABLE IF NOT EXISTS post (
      id SERIAL PRIMARY KEY AUTOINCREMENT,
      title VARCHAR(255) NOT NULL,
      body TEXT NOT NULL
    );
  `); err != nil {
		return err
	}

	if _, err := tx.Exec(`
    CREATE TABLE IF NOT EXISTS comment (
      id SERIAL PRIMARY KEY AUTOINCREMENT,
      post_id INTEGER NOT NULL,
      body TEXT NOT NULL
    );
  `); err != nil {
		return err
	}

	// create index on comment.post_id
	if _, err := tx.Exec(`
    CREATE INDEX IF NOT EXISTS comment_post_id_idx ON comment (post_id);
  `); err != nil {
		return err
	}

	// add foreign key for comment.post_id
	if _, err := tx.Exec(`
    ALTER TABLE comment
    ADD CONSTRAINT fk_comment_post_id
    FOREIGN KEY (post_id)
    REFERENCES post(id);
  `); err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
