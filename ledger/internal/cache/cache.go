package cache

import (
	"context"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func Connect() (*redis.Client, error) {
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		db = 0
	}
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
	context := context.Background()
	_, err = client.Ping(context).Result()
	if err != nil {
		println(err.Error())
		return nil, err
	}
	println("Connected to redis")
	return client, nil
}
