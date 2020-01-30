package chat_service

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
	service = RedisChatService{}
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

	err := service.SendNewMessage(client, "42")
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

func TestProcessMessage(t *testing.T) {
	defer client.FlushAll()

	tstMsg := "42"
	err := service.SendNewMessage(client, tstMsg)
	assert.Nil(t, err)

	msg, err := service.ProcessNewMessage(client, time.Second)
	assert.Nil(t, err)
	assert.Equal(t, msg, tstMsg)
}

func TestGetLastNMessages(t *testing.T) {
	defer client.FlushAll()

	service.SendNewMessage(client, "first")
	service.SendNewMessage(client, "second")
	service.SendNewMessage(client, "third")
	service.SendNewMessage(client, "fourth")

	service.ProcessNewMessage(client, time.Second)
	service.ProcessNewMessage(client, time.Second)
	service.ProcessNewMessage(client, time.Second)

	msgs, err := service.GetLastNMessages(client, 4)
	assert.Nil(t, err)
	assert.Equal(t, len(msgs), 3)
}

func TestFromToMessages(t *testing.T) {
	defer client.FlushAll()

	service.SendNewMessage(client, "first")
	service.SendNewMessage(client, "second")
	service.SendNewMessage(client, "third")
	service.SendNewMessage(client, "fourth")

	service.ProcessNewMessage(client, time.Second)
	service.ProcessNewMessage(client, time.Second)
	service.ProcessNewMessage(client, time.Second)

	msgs, err := service.GetFromToMessages(client, 1, 2)
	assert.Nil(t, err)
	assert.Equal(t, "second", msgs[0])
	assert.Equal(t, "first", msgs[1])
}
