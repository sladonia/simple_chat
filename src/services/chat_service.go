package services

import (
	"errors"
	"github.com/go-redis/redis/v7"
)

const (
	messagesList    = "messages_list"
	messagesChannel = "messages_channel"
	usersSet        = "users"
)

var (
	UserExistsError                           = errors.New("username occupied")
	ChatService     RedisChatServiceInterface = &redisChatService{}
)

type RedisChatServiceInterface interface {
	AddUser(client *redis.Client, username string) error
	ArchiveMessage(client *redis.Client, msg string) error
	GetLastNMessages(client *redis.Client, n int64) ([]string, error)
	GetFromToMessages(client *redis.Client, from, to int64) ([]string, error)
	PublishMessage(client *redis.Client, msg string) error
	SubscribeToMessageChannel(client *redis.Client, msgCh chan<- string)
}

type redisChatService struct{}

func (s *redisChatService) AddUser(client *redis.Client, username string) error {
	exists, err := client.SIsMember(usersSet, username).Result()
	if err != nil {
		return err
	} else if exists {
		return UserExistsError
	}
	_, err = client.SAdd(usersSet, username).Result()
	return err
}

func (s *redisChatService) ArchiveMessage(client *redis.Client, msg string) error {
	_, err := client.LPush(messagesList, msg).Result()
	return err
}

func (s *redisChatService) GetLastNMessages(client *redis.Client, n int64) ([]string, error) {
	return client.LRange(messagesList, 0, n).Result()
}

func (s *redisChatService) GetFromToMessages(client *redis.Client, from, to int64) ([]string, error) {
	return client.LRange(messagesList, from, to).Result()
}

func (s *redisChatService) PublishMessage(client *redis.Client, msg string) error {
	_, err := client.Publish(messagesChannel, msg).Result()
	return err
}

func (s *redisChatService) SubscribeToMessageChannel(client *redis.Client, msgCh chan<- string) {
	pubSub := client.Subscribe(messagesChannel)

	for {
		msg, _ := pubSub.ReceiveMessage()
		msgCh <- msg.Payload
	}
}
