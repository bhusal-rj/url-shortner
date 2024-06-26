package database

import (
	"context"
	"github.com/go-redis/redis/v8"
	"os"
)

var Ctx = context.Background()

// creation of database client of redis
func CreateClient(dbNo int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DB_ADDR"),
		Password: os.Getenv("DB_ADDR"),
		DB:       dbNo,
	})
	return rdb
}
