package redisdb

import "github.com/go-redis/redis/v7"

var (
	RedisClient *redis.Client
)

func InitRedisClient(options *redis.Options) error {
	RedisClient = redis.NewClient(options)
	_, err := RedisClient.Ping().Result()
	return err
}
