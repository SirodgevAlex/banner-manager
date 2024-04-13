package db

import (
    "context"
    "time"
	"fmt"
    "github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var rdb *redis.Client

func InitRedis() {
    rdb = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })
}

func CloseRedis() {
    rdb.Close()
}

func AddRedisToken(token string, expiration time.Duration) error {
    return rdb.Set(ctx, token, true, expiration).Err()
}

func IsRedisTokenExists(token string) (bool, error) {
    exists, err := rdb.Exists(ctx, token).Result()
    if err != nil {
        return false, err
    }

    return exists > 0, nil
}

func DeleteRedisToken(token string) error {
    return rdb.Del(ctx, token).Err()
}

func PeriodicallyCleanExpiredRedisTokens(interval time.Duration, stop <-chan struct{}) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            tokens, err := rdb.Keys(ctx, "*").Result()
            if err != nil {
                fmt.Println("Ошибка при получении ключей:", err)
                continue
            }

            for _, token := range tokens {
                ttl := rdb.TTL(ctx, token).Val()
                if ttl < 0 {
                    if err := rdb.Del(ctx, token).Err(); err != nil {
                        fmt.Println("Ошибка при удалении ключа:", err)
                        continue
                    }
                }
            }
        case <-stop:
            return
        }
    }
}
