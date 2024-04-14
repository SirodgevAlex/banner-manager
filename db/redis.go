package db

import (
	"banner-manager/internal/models"
	"context"
	"fmt"
	"time"
    "encoding/json"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

//User

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

//Banner

var rdbBanner *redis.Client

func InitRedisBanner() {
	rdbBanner = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})
}

func CloseRedisBanner() {
	rdbBanner.Close()
}

func CacheBannerInRedis(tagID string, featureID string, banner *models.Banner) error {
    bannerKey := fmt.Sprintf("banner:feature_%s:tag_%s", featureID, tagID)

    jsonData, err := json.Marshal(banner)
    if err != nil {
        return fmt.Errorf("ошибка при сериализации баннера в JSON: %v", err)
    }

    err = rdbBanner.Set(rdbBanner.Context(), bannerKey, jsonData, 0).Err()
    if err != nil {
        return fmt.Errorf("ошибка при сохранении баннера в Redis: %v", err)
    }

    return nil
}

func GetBannerFromRedis(tagID string, featureID string) (*models.Banner, error) {
    bannerKey := fmt.Sprintf("banner:%s:%s", featureID, tagID)

    jsonData, err := rdbBanner.Get(rdbBanner.Context(), bannerKey).Bytes()
    if err != nil {
        if err == redis.Nil {
            return nil, fmt.Errorf("баннер для feature_id=%s и tag_id=%s не найден", featureID, tagID)
        }
        return nil, fmt.Errorf("ошибка при получении баннера из Redis: %v", err)
    }

    var banner models.Banner
    err = json.Unmarshal(jsonData, &banner)
    if err != nil {
        return nil, fmt.Errorf("ошибка при десериализации баннера из JSON: %v", err)
    }

    return &banner, nil
}

func DeleteBannerFromRedis(tagID string, featureID string) error {
    bannerKey := fmt.Sprintf("banner:%s:%s", tagID, featureID)

    _, err := rdbBanner.Del(ctx, bannerKey).Result()
    if err != nil {
        return err
    }

    return nil
}

func UpdateBannerFromRedis(tagID string, featureID string, banner *models.Banner) error {
    if err := DeleteBannerFromRedis(tagID, featureID); err != nil {
        return fmt.Errorf("failed to delete previous banner from Redis: %v", err)
    }

    if err := CacheBannerInRedis(tagID, featureID, banner); err != nil {
        return fmt.Errorf("failed to cache updated banner in Redis: %v", err)
    }

    return nil
}


