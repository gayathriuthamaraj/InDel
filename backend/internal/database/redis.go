package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/Shravanthi20/InDel/backend/internal/config"
)

var Rdb *redis.Client

func InitRedis(cfg *config.Config) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		// Fallback to simple address if it's not a full URL format (like redis://...)
		opts = &redis.Options{
			Addr: cfg.RedisURL,
		}
	}

	Rdb = redis.NewClient(opts)

	// Ping to verify connection
	_, err = Rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis at %s: %v", cfg.RedisURL, err)
		return nil, err
	}

	log.Printf("Connected successfully to Redis at %s", cfg.RedisURL)
	return Rdb, nil
}
