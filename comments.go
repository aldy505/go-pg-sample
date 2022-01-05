package main

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID     int    `json:"id"`
	PostID int    `json:"post_id,omitempty"`
	Body   string `json:"body"`
}

func (d *Dependency) GetComments(ctx context.Context, postId int) ([]Comment, error) {
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return []Comment{}, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted, ReadOnly: true})
	if err != nil {
		return []Comment{}, err
	}

	rows, err := conn.QueryContext(ctx, "SELECT id, body FROM comment WHERE post_id = $1", postId)
	if err != nil {
		return []Comment{}, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.Body); err != nil {
			return []Comment{}, err
		}
		comments = append(comments, comment)
	}

	err = tx.Commit()
	if err != nil {
		return []Comment{}, err
	}

	return comments, nil
}

func (d *Dependency) AddComment(ctx context.Context, comment Comment) (Comment, error) {
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return Comment{}, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})
	if err != nil {
		return Comment{}, err
	}

	row := tx.QueryRowContext(ctx, `INSERT INTO comment (post_id, body) VALUES ($1, $2) RETURNING id`, comment.PostID, comment.Body)

	var commentId int
	if err := row.Scan(&commentId); err != nil {
		return Comment{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Comment{}, err
	}

	return Comment{ID: commentId, Body: comment.Body}, nil
}

func (d *Dependency) DeleteComment(ctx context.Context, id int) (Comment, error) {
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return Comment{}, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})
	if err != nil {
		return Comment{}, err
	}

	// Check if comment exists
	row := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT id FROM comment WHERE id = $1)", id)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return Comment{}, err
	}

	var commentId int
	if !exists {
		row = tx.QueryRowContext(ctx, `DELETE FROM comment WHERE id = $1 RETURNING id`, id)
		if err := row.Scan(&commentId); err != nil {
			return Comment{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return Comment{}, err
	}

	return Comment{ID: commentId, Body: ""}, nil
}
