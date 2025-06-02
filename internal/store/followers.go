package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Follower struct {
	UserId     int       `json:"user_id"`
	FollowerId int       `json:"follower_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, userId, followerId int) error {
	query := `INSERT INTO followers(user_id,follower_id) VALUES($1,$2)`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userId, followerId)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
	}

	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, userId, followerId int) error {
	query := `DELETE FROM followers 
	WHERE user_id = $1 AND follower_id = $2`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userId, followerId)

	return err
}
