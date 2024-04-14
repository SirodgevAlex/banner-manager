package db

import (
	"context"
	"fmt"
	"time"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var rdbUser *redis.Client

func InitRedisUser() {
	rdbUser = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

func CloseRedisUser() {
	rdbUser.Close()
}

func AddRedisToken(token string, expiration time.Duration) error {
	return rdbUser.Set(ctx, token, true, expiration).Err()
}

func IsRedisTokenExists(token string) (bool, error) {
	exists, err := rdbUser.Exists(ctx, token).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

func DeleteRedisToken(token string) error {
	return rdbUser.Del(ctx, token).Err()
}

func PeriodicallyCleanExpiredRedisTokens(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tokens, err := rdbUser.Keys(ctx, "*").Result()
			if err != nil {
				fmt.Println("Ошибка при получении ключей:", err)
				continue
			}

			for _, token := range tokens {
				ttl := rdbUser.TTL(ctx, token).Val()
                fmt.Println(ttl)
				if ttl == 0 {
					if err := rdbUser.Del(ctx, token).Err(); err != nil {
						fmt.Println("Ошибка при удалении ключа:", err)
						continue
					} else {
						fmt.Println("Ключ", token, "удален")
					}
				}
			}
		case <-stop:
			return
		}
	}
}
