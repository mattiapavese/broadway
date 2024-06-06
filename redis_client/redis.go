package redisclient

import (
	"github.com/redis/go-redis/v9"
)

func GetRedisClient(ip string, port string, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     ip + ":" + port,
		Password: password,
		DB:       0,
	})
}
