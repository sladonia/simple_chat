package chat_service

import (
	"errors"
	"github.com/go-redis/redis/v7"
	"time"
)

const (
	oldMessagesList = "old_messages"
	newMessageList  = "new_messages"
	usersSet        = "users"
)

var TimeoutError = errors.New("timeout error")
var UserExistsError = errors.New("username occupied")

type RedisChatServiceInterface interface {
	SendNewMessage(client *redis.Client, msg string) error
	AddUser(client *redis.Client, username string) error
	ProcessNewMessage(client *redis.Client, timeout time.Duration) (string, error)
	GetLastNMessages(client *redis.Client, n int64) ([]string, error)
	GetFromToMessages(client *redis.Client, from, to int64) ([]string, error)
}

type RedisChatService struct{}

func (s *RedisChatService) SendNewMessage(client *redis.Client, msg string) error {
	_, err := client.LPush(newMessageList, msg).Result()
	return err
}

func (s *RedisChatService) AddUser(client *redis.Client, username string) error {
	exists, err := client.SIsMember(usersSet, username).Result()
	if err != nil {
		return err
	} else if exists {
		return UserExistsError
	}
	_, err = client.SAdd(usersSet, username).Result()
	return err
}

func (s *RedisChatService) ProcessNewMessage(client *redis.Client, timeout time.Duration) (string, error) {
	res, err := client.BRPopLPush(newMessageList, oldMessagesList, timeout).Result()
	if errors.Is(err, redis.Nil) {
		return "", TimeoutError
	}
	return res, err
}

func (s *RedisChatService) GetLastNMessages(client *redis.Client, n int64) ([]string, error) {
	return client.LRange(oldMessagesList, 0, n).Result()
}

func (s *RedisChatService) GetFromToMessages(client *redis.Client, from, to int64) ([]string, error) {
	return client.LRange(oldMessagesList, from, to).Result()
}
