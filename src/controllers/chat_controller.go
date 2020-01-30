package controllers

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"simple_chat/src/domains/message_domain"
	"simple_chat/src/hub/in_memory_hub"
)

var (
	ChatController ChatControllerInterface = &chatController{}
	upgrader                               = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type ChatControllerInterface interface {
	HandleChat(w http.ResponseWriter, r *http.Request)
}

type chatController struct{}

func (c *chatController) HandleChat(w http.ResponseWriter, r *http.Request) {
	log.Printf("creating websocket connection with %s", r.Host)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("unable to upgrade connecion. err: %s", err.Error())
		return
	}

	in_memory_hub.Hub.AddClientCh <- conn

	for {
		msg := message_domain.Message{}
		err = conn.ReadJSON(&msg)
		if err != nil {
			err, ok := err.(*websocket.CloseError)
			if ok {
				in_memory_hub.Hub.RemoveClientCh <- conn
				return
			}
			log.Printf("unble to read message. err: %s", err.Error())
		}

		in_memory_hub.Hub.BroadcastCh <- msg
		log.Printf("received messae: %s", msg.String())
	}
}
