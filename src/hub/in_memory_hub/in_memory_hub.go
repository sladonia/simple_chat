package in_memory_hub

import (
	"github.com/gorilla/websocket"
	"log"
	"simple_chat/src/domains/message_domain"
)

var (
	Hub = NewInMemoryHub()
)

func init() {
	log.Printf("starting hub")
	go Hub.Run()
}

type InMemoryHub struct {
	Clients        map[string]*websocket.Conn
	AddClientCh    chan *websocket.Conn
	RemoveClientCh chan *websocket.Conn
	BroadcastCh    chan message_domain.Message
}

func NewInMemoryHub() *InMemoryHub {
	return &InMemoryHub{
		Clients:        make(map[string]*websocket.Conn),
		AddClientCh:    make(chan *websocket.Conn),
		RemoveClientCh: make(chan *websocket.Conn),
		BroadcastCh:    make(chan message_domain.Message),
	}
}

func (h *InMemoryHub) AddClient(conn *websocket.Conn) {
	addr := conn.LocalAddr().String()
	log.Printf("adding client address: %s", addr)
	h.Clients[addr] = conn
}

func (h InMemoryHub) RemoveClient(conn *websocket.Conn) {
	addr := conn.LocalAddr().String()
	log.Printf("removeing client address: %s", addr)
	delete(h.Clients, addr)
}

func (h *InMemoryHub) Broadcast(msg message_domain.Message) {
	log.Printf("start message broadcasting: %s", msg.Text)
	for _, conn := range h.Clients {
		err := conn.WriteJSON(msg)
		if err != nil {
			log.Printf("unble to write msg. err: %s", err.Error())
		}
	}
}

func (h *InMemoryHub) Run() {
	for {
		select {
		case conn := <-h.AddClientCh:
			h.AddClient(conn)
		case conn := <-h.RemoveClientCh:
			h.RemoveClient(conn)
		case msg := <-h.BroadcastCh:
			h.Broadcast(msg)
		}
	}
}
