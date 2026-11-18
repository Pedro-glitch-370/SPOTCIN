package database

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func OpenRedis() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	RDB = redis.NewClient(&redis.Options{Addr: addr})
	return RDB.Ping(context.Background()).Err()
}