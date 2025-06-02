package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AlieNoori/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Hour

func (s *UserStore) Get(ctx context.Context, userId int) (*store.User, error) {
	cacheKey := fmt.Sprintf("user/%d", userId)

	userJSON, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if userJSON != "" {
		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user/%d", user.ID)

	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	err = s.rdb.Set(ctx, cacheKey, userJSON, UserExpTime).Err()

	return err
}
