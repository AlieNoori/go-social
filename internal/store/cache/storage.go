package cache

import (
	"context"
	"time"

	"github.com/AlieNoori/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users interface {
		Get(context.Context, int) (*store.User, error)
		Set(context.Context, *store.User) error
	}
	RateLimit interface {
		Incrementor(context.Context, string) error
		Get(context.Context, string) (int, error)
		Set(context.Context, string, time.Duration) error
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users:     &UserStore{rdb},
		RateLimit: &RateLimitStore{rdb},
	}
}
