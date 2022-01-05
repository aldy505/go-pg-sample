package main

import (
	"context"
	"database/sql"
	"sync"
)

type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (d *Dependency) GetPosts(ctx context.Context) ([]Post, error) {
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return []Post{}, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted, ReadOnly: true})
	if err != nil {
		return []Post{}, err
	}

	rows, err := tx.QueryContext(ctx, "SELECT id, title, body FROM posts")
	if err != nil {
		return []Post{}, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Body); err != nil {
			return []Post{}, err
		}
		posts = append(posts, post)
	}

	err = tx.Commit()
	if err != nil {
		return []Post{}, err
	}
	return posts, nil
}

func (d *Dependency) GetPostById(ctx context.Context, id int) (Post, error) {
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return Post{}, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted, ReadOnly: true})
	if err != nil {
		return Post{}, err
	}

	row := tx.QueryRowContext(ctx, "SELECT id, title, body FROM posts WHERE id = $1", id)
	var post Post
	if err := row.Scan(&post.ID, &post.Title, &post.Body); err != nil {
		return Post{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func (d *Dependency) AddPost(ctx context.Context, post Post) (Post, error) {
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return Post{}, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelWriteCommitted, ReadOnly: false})
	if err != nil {
		return Post{}, err
	}

	row := tx.QueryRowContext(ctx, "INSERT INTO posts (title, body) VALUES ($1, $2) RETURNING id", post.Title, post.Body)
	if err := row.Scan(&post.ID); err != nil {
		return Post{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func (d *Dependency) UpdatePost(ctx context.Context, post Post) (Post, error) {
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return Post{}, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelWriteCommitted, ReadOnly: false})
	if err != nil {
		return Post{}, err
	}

	// Checck if the post exists
	row := tx.QueryRowContext(ctx, "SELECT title, body FROM posts WHERE id = $1", post.ID)
	var oldPost Post
	if err := row.Scan(&oldPost.Title, &oldPost.Body); err != nil {
		return Post{}, err
	}

	// Check if there is any change to the current post
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if post.Title == "" {
			post.Title = oldPost.Title
		}
	}()

	go func() {
		defer wg.Done()
		if post.Body == "" {
			post.Body = oldPost.Body
		}
	}()

	wg.Wait()

	// Update the post
	if _, err := tx.ExecContext(ctx, "UPDATE posts SET title = $1, body = $2 WHERE id = $3", post.Title, post.Body, post.ID); err != nil {
		return Post{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func (d *Dependency) DeletePost(ctx context.Context, id int) (Post, error) {
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return Post{}, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelWriteCommitted, ReadOnly: false})
	if err != nil {
		return Post{}, err
	}

	// Check if the post exists
	row := tx.QueryRowContext(ctx, "SELECT id FROM posts WHERE id = $1", id)
	var post Post
	if err := row.Scan(&post.ID); err != nil {
		return Post{}, err
	}

	// Delete the post
	if _, err := tx.ExecContext(ctx, "DELETE FROM posts WHERE id = $1", id); err != nil {
		return Post{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Post{}, err
	}

	return post, nil
}
