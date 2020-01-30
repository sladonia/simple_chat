package services

import (
	"errors"
	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const (
	address = "localhost:6379"
)

var (
	client  *redis.Client
	service = redisChatService{}
)

func TestMain(m *testing.M) {
	options := &redis.Options{
		Addr: address,
	}
	client = redis.NewClient(options)
	os.Exit(m.Run())
}

func TestSendNewMessage(t *testing.T) {
	defer client.FlushAll()

	err := service.ArchiveMessage(client, "42")
	assert.Nil(t, err)
}

func TestAddUser(t *testing.T) {
	defer client.FlushAll()

	username := "Richard"
	err := service.AddUser(client, username)
	assert.Nil(t, err)
	err = service.AddUser(client, username)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, UserExistsError))
}

func TestGetLastNMessages(t *testing.T) {
	defer client.FlushAll()

	service.ArchiveMessage(client, "first")
	service.ArchiveMessage(client, "second")
	service.ArchiveMessage(client, "third")
	service.ArchiveMessage(client, "fourth")

	msgs, err := service.GetLastNMessages(client, 4)
	assert.Nil(t, err)
	assert.Equal(t, len(msgs), 4)
	assert.Equal(t, "fourth", msgs[0])
}

func TestFromToMessages(t *testing.T) {
	defer client.FlushAll()

	service.ArchiveMessage(client, "first")
	service.ArchiveMessage(client, "second")
	service.ArchiveMessage(client, "third")
	service.ArchiveMessage(client, "fourth")

	msgs, err := service.GetFromToMessages(client, 1, 2)
	assert.Nil(t, err)
	assert.Equal(t, "third", msgs[0])
	assert.Equal(t, "second", msgs[1])
}

func TestPublishMessage(t *testing.T) {
	defer client.FlushAll()

	err := service.PublishMessage(client, "ho-ho-ho motherfucker!")
	assert.Nil(t, err)
}

func TestSubscribeToMessageChannel(t *testing.T) {
	defer client.FlushAll()

	msgCh := make(chan string, 10)
	msg1 := "foo"
	msg2 := "bar"
	_ = msg2

	go service.SubscribeToMessageChannel(client, msgCh)

	time.Sleep(100 * time.Millisecond)

	service.PublishMessage(client, msg1)
	service.PublishMessage(client, msg2)

	assert.Equal(t, msg1, <-msgCh)
	assert.Equal(t, msg2, <-msgCh)
}
