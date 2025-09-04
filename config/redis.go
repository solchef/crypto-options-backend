package config

import (
	"context"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	Redis *redis.Client
	Ctx   = context.Background()
)

func ConnectRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		if v, err := strconv.Atoi(dbStr); err == nil {
			db = v
		}
	}

	Redis = redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
}
