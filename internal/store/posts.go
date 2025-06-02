package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type PostStore struct {
	db *sql.DB
}

type Post struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserId    int       `json:"user_id"`
	Tags      []string  `json:"tags"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comments_count"`
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
	INSERT INTO posts (content,title,user_id,tags) 
	VALUES ($1,$2,$3,$4) RETURNING id,created_at, updated_at;

	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserId,
		pq.Array(post.Tags)).
		Scan(
			&post.ID,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetById(ctx context.Context, postID int) (*Post, error) {
	query := `SELECT id,user_id,content,title,version,tags,created_at,updated_at FROM posts 
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	post := &Post{}

	err := s.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.UserId,
		&post.Content,
		&post.Title,
		&post.Version,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return post, nil
}

func (s *PostStore) Delete(ctx context.Context, postID int) error {
	query := `
	DELETE FROM posts 
	WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
	UPDATE posts
    SET title = $3, content = $4, tags = $5, updated_at= NOW(), version= version + 1
	WHERE id = $1 AND version = $2
	RETURNING updated_at,version
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(ctx, query,
		post.ID,
		post.Version,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
	).Scan(
		&post.UpdatedAt,
		&post.Version,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userId int, fq PaginatedFeedQeury) ([]PostWithMetadata, error) {
	query := `
SELECT p.id,p.user_id,p.title,p.content,p.created_at,p.version,p.tags,username,COUNT(c.id) AS comments_count
FROM posts as p
LEFT JOIN comments AS c ON c.post_id = p.id
LEFT JOIN users AS u ON u.id = p.user_id
INNER JOIN followers AS f ON f.follower_id = p.user_id OR p.user_id = $1
WHERE 
	f.user_id = $1 AND 
	(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
	(p.tags @> $5 OR $5 = '{}')
GROUP BY p.id,u.username` + fmt.Sprintf(" ORDER BY p.created_at %s ", strings.ToUpper(fq.Sort)) + `LIMIT $2 OFFSET $3;`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userId, fq.Limit, fq.Offset, pq.Array(fq.Tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []PostWithMetadata
	for rows.Next() {
		var pwd PostWithMetadata
		if err := rows.Scan(
			&pwd.ID,
			&pwd.UserId,
			&pwd.Title,
			&pwd.Content,
			&pwd.CreatedAt,
			&pwd.Version,
			pq.Array(&pwd.Tags),
			&pwd.User.UserName,
			&pwd.CommentsCount,
		); err != nil {
			return nil, err
		}

		feed = append(feed, pwd)
	}

	return feed, nil
}
