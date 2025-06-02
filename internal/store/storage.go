package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          error         = errors.New("resource not found")
	ErrConflict          error         = errors.New("resource already exits")
	queryTimeoutDuration time.Duration = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetById(context.Context, int) (*Post, error)
		Delete(context.Context, int) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int, PaginatedFeedQeury) ([]PostWithMetadata, error)
	}

	Users interface {
		Activate(context.Context, string) error
		Create(context.Context, *sql.Tx, *User) error
		GetById(context.Context, int) (*User, error)
		GetByEmail(context.Context, string) (*User, error)
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		createUserInvitation(context.Context, *sql.Tx, string, time.Duration, int) error
		Delete(context.Context, int) error
	}

	Comments interface {
		Create(context.Context, *Comment) error
		GetByPostId(context.Context, int) ([]Comment, error)
	}

	Followers interface {
		Follow(context.Context, int, int) error
		Unfollow(context.Context, int, int) error
	}
	Roles interface {
		GetByName(context.Context, string) (*Role, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Users:     &UserStore{db},
		Comments:  &CommentStore{db},
		Followers: &FollowerStore{db},
		Roles:     &RoleStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
