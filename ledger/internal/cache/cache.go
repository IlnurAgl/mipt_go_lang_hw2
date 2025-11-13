package cache

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

func Connect() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	context := context.Background()
	_, err := client.Ping(context).Result()
	if err != nil {
		println(err.Error())
		return nil, err
	}
	println("Connected to redis")
	return client, nil
}

func Close(client *redis.Client) error {
	err := client.Close()
	if err != nil {
		return err
	}
	return nil
}
