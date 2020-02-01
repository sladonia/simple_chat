package hub

import (
	"fmt"
	"github.com/gorilla/websocket"
	"simple_chat/src/datasources/redisdb"
	"simple_chat/src/logger"
	"simple_chat/src/services"
)

var (
	Hub = NewChatHub()
)

type ChatHub struct {
	Clients        map[string]*websocket.Conn
	AddClientCh    chan UserConnection
	RemoveClientCh chan string
	BroadcastCh    chan string
}

func NewChatHub() *ChatHub {
	return &ChatHub{
		Clients:        make(map[string]*websocket.Conn),
		AddClientCh:    make(chan UserConnection),
		RemoveClientCh: make(chan string),
		BroadcastCh:    make(chan string),
	}
}

func (h *ChatHub) AddClient(uc UserConnection) {
	logger.Logger.Debugf("adding client name: %s", uc.Name)
	err := services.ChatService.AddUser(redisdb.RedisClient, uc.Name)
	if err != nil {
		logger.Logger.Errorw("unable to add client", "client_name", uc.Name, "err", err)
		return
	}
	h.Clients[uc.Name] = uc.Conn
	err = services.ChatService.PublishMessage(redisdb.RedisClient,
		fmt.Sprintf("%s joined the chat", uc.Name))
	if err != nil {
		logger.Logger.Errorw("unable to publish message", "err", err)
	}
}

func (h ChatHub) RemoveClient(clientName string) {
	logger.Logger.Infof("removing client name: %s", clientName)
	delete(h.Clients, clientName)
	_ = services.ChatService.RemoveUser(redisdb.RedisClient, clientName)
	err := services.ChatService.PublishMessage(redisdb.RedisClient,
		fmt.Sprintf("%s left the chat", clientName))
	if err != nil {
		logger.Logger.Errorw("unable to publish message", "err", err)
	}
}

func (h *ChatHub) Broadcast(msg string) {
	logger.Logger.Debugf("start message broadcasting: %s", msg)
	for _, conn := range h.Clients {
		err := conn.WriteMessage(1, []byte(msg))
		if err != nil {
			logger.Logger.Errorw("unable to write msg", "err", err.Error())
		}
	}
}

func (h *ChatHub) ShutDown() {
	users := make([]string, 0, 10)
	for user := range h.Clients {
		users = append(users, user)
	}
	err := services.ChatService.RemoveUsers(redisdb.RedisClient, users...)
	if err != nil {
		logger.Logger.Errorw("error removing users from redis", "err", err)
	}
}

func (h *ChatHub) Run() {
	go services.ChatService.SubscribeToMessageChannel(redisdb.RedisClient, h.BroadcastCh)
	for {
		select {
		case userConn := <-h.AddClientCh:
			h.AddClient(userConn)
		case clientName := <-h.RemoveClientCh:
			h.RemoveClient(clientName)
		case msg := <-h.BroadcastCh:
			h.Broadcast(msg)
			err := services.ChatService.ArchiveMessage(redisdb.RedisClient, msg)
			if err != nil {
				logger.Logger.Errorw("unable to archive the message", "err", err)
			}
		}
	}
}
