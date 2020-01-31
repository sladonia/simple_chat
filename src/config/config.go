package config

import "github.com/jinzhu/configor"

const configPath = "config.yml"

var Config Configuration

type Configuration struct {
	Address     string `env:"SERVICE_ADDRESS"`
	LogLevel    string `env:"LOG_LEVEL"`
	RedisConfig RedisConfig
}

type RedisConfig struct {
	Address  string `env:"REDIS_ADDRESS"`
	PoolSize int    `env:"REDIS_POOL_SIZE"`
}

func Load() error {
	return configor.Load(&Config, configPath)
}
