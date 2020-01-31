package hub

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"simple_chat/src/datasources/redisdb"
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
	log.Printf("adding client name: %s", uc.Name)
	err := services.ChatService.AddUser(redisdb.RedisClient, uc.Name)
	if err != nil {
		log.Printf("anable to add client: %s", uc.Name)
		return
	}
	h.Clients[uc.Name] = uc.Conn
	services.ChatService.PublishMessage(redisdb.RedisClient,
		fmt.Sprintf("%s joined the chat", uc.Name))
}

func (h ChatHub) RemoveClient(clientName string) {
	log.Printf("removeing client name: %s", clientName)
	delete(h.Clients, clientName)
	_ = services.ChatService.RemoveUser(redisdb.RedisClient, clientName)
	services.ChatService.PublishMessage(redisdb.RedisClient,
		fmt.Sprintf("%s left the chat", clientName))
}

func (h *ChatHub) Broadcast(msg string) {
	log.Printf("start message broadcasting: %s", msg)
	for _, conn := range h.Clients {
		err := conn.WriteMessage(1, []byte(msg))
		if err != nil {
			log.Printf("unble to write msg. err: %s", err.Error())
		}
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
			services.ChatService.ArchiveMessage(redisdb.RedisClient, msg)
		}
	}
}
