package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

func Connect(addr string) error {
	Client = redis.NewClient(&redis.Options{Addr: addr})
	if _, err := Client.Ping(Ctx).Result(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}
	return nil
}

func SetRefreshToken(token string, userID uint, ttl time.Duration) error {
	return Client.Set(Ctx, "refresh:"+token, userID, ttl).Err()
}

func GetRefreshTokenOwner(token string) (uint, error) {
	val, err := Client.Get(Ctx, "refresh:"+token).Result()
	if err != nil {
		return 0, err
	}
	var id uint
	_, err = fmt.Sscanf(val, "%d", &id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func RevokeRefreshToken(token string) error {
	return Client.Del(Ctx, "refresh:"+token).Err()
}

func BlacklistAccessToken(jti string, ttl time.Duration) error {
	return Client.Set(Ctx, "blacklist:access:"+jti, "1", ttl).Err()
}

func IsAccessTokenBlacklisted(jti string) (bool, error) {
	exists, err := Client.Exists(Ctx, "blacklist:access:"+jti).Result()
	return exists > 0, err
}
