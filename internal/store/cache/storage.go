package cache

import (
	"context"

	"github.com/AlieNoori/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users interface {
		Get(context.Context, int) (*store.User, error)
		Set(context.Context, *store.User) error
	}
	// RateLimit interface {
	// 	Get(context.Context, int) (string, error)
	// 	Set(context.Context, *store.User) error
	// }
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb},
	}
}
