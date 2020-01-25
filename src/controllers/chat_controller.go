package controllers

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
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

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("unble to read message. err: %s", err.Error())
			return
		}
		log.Printf("received messae of type %d. msg: %s", messageType, string(p))
		err = conn.WriteMessage(messageType, p)
		if err != nil {
			log.Printf("unble to write message. err: %s", err.Error())
			return
		}
		log.Printf("message sent msg: %s", string(p))
	}
}
