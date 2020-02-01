package services

import (
	"errors"
	"github.com/go-redis/redis/v7"
)

const (
	messagesList    = "messages_list"
	messagesChannel = "messages_channel"
	UsersSet        = "users"
)

var (
	UserExistsError                             = errors.New("username occupied")
	UserNotExistError                           = errors.New("user does not exist")
	ChatService       RedisChatServiceInterface = &redisChatService{}
)

type RedisChatServiceInterface interface {
	AddUser(client *redis.Client, username string) error
	ArchiveMessage(client *redis.Client, msg string) error
	GetLastNMessages(client *redis.Client, n int64) ([]string, error)
	GetFromToMessages(client *redis.Client, from, to int64) ([]string, error)
	PublishMessage(client *redis.Client, msg string) error
	SubscribeToMessageChannel(client *redis.Client, msgCh chan<- string)
	RemoveUser(client *redis.Client, username string) error
	UsernameIsFree(client *redis.Client, username string) (bool, error)
	RemoveUsers(client *redis.Client, users ...string) error
}

type redisChatService struct{}

func (s *redisChatService) AddUser(client *redis.Client, username string) error {
	exists, err := client.SIsMember(UsersSet, username).Result()
	if err != nil {
		return err
	} else if exists {
		return UserExistsError
	}
	_, err = client.SAdd(UsersSet, username).Result()
	return err
}

func (s *redisChatService) RemoveUser(client *redis.Client, username string) error {
	numRemoved, err := client.SRem(UsersSet, username).Result()
	if err != nil {
		return err
	} else if numRemoved != 1 {
		return UserNotExistError
	}
	return nil
}

func (s *redisChatService) UsernameIsFree(client *redis.Client, username string) (bool, error) {
	exist, err := client.SIsMember(UsersSet, username).Result()
	return !exist, err
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

func (s *redisChatService) RemoveUsers(client *redis.Client, users ...string) error {
	if len(users) == 0 {
		return nil
	}
	_, err := client.SRem(UsersSet, users).Result()
	return err
}
