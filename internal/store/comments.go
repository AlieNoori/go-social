package store

import (
	"context"
	"database/sql"
	"time"
)

type Comment struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	PostId    int       `json:"post_id"`
	UserId    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user"`
}

type CommentStore struct {
	db *sql.DB
}

func (s *CommentStore) Create(ctx context.Context, commnet *Comment) error {
	query := `
	INSERT INTO comments(post_id,user_id,content) VALUES($1,$2,$3)
	RETURNING id, created_at
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(ctx, query, commnet.PostId,
		commnet.UserId,
		commnet.Content,
	).Scan(
		&commnet.ID,
		&commnet.CreatedAt,
	); err != nil {
		return err
	}

	return nil
}

func (s *CommentStore) GetByPostId(ctx context.Context, postId int) ([]Comment, error) {
	query := `
	SELECT c.id,c.post_id,c.user_id,c.content, c.created_at, u.username,u.id FROM comments as c
INNER JOIN users as u ON u.id = c.user_id
WHERE c.post_id = $1
ORDER BY c.created_at DESC;
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, postId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]Comment, 0)

	for rows.Next() {
		var comment Comment
		comment.User = User{}

		if err := rows.Scan(
			&comment.ID,
			&comment.PostId,
			&comment.UserId,
			&comment.Content,
			&comment.User.UserName,
			&comment.User.ID,
		); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}
