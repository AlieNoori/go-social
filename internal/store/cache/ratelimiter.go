package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RateLimitStore struct {
	rdb *redis.Client
}

const prefix = "ratelimit:"

func (s *RateLimitStore) Get(ctx context.Context, ip string) (int, error) {
	key := prefix + ip
	res, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return -1, err
	}

	count, err := strconv.Atoi(res)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (s *RateLimitStore) Set(ctx context.Context, ip string, ttl time.Duration) error {
	key := prefix + ip

	err := s.rdb.Set(ctx, key, 1, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *RateLimitStore) Incrementor(ctx context.Context, ip string) error {
	key := prefix + ip

	err := s.rdb.Incr(ctx, key).Err()

	return err
}
