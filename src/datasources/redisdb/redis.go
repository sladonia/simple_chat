package redisdb

import "github.com/go-redis/redis/v7"

var (
	redisClient *redis.Client
)

func InitRedisClient(options *redis.Options) error {
	redisClient = redis.NewClient(options)
	_, err := redisClient.Ping().Result()
	return err
}
