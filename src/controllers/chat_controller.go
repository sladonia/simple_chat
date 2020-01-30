package controllers

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"simple_chat/src/datasources/redisdb"
	"simple_chat/src/domains/message_domain"
	"simple_chat/src/hub/in_memory_hub"
	"simple_chat/src/services"
	"strings"
)

var (
	NameError                              = errors.New("name can not be empty")
	ChatController ChatControllerInterface = &chatController{}
	upgrader                               = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type ChatControllerInterface interface {
	HandleChat(w http.ResponseWriter, r *http.Request)
	HandleLogIn(w http.ResponseWriter, r *http.Request)
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

func (c *chatController) HandleLogIn(w http.ResponseWriter, r *http.Request) {
	type msg struct {
		Err string
	}

	log.Printf("processing logging form request from %s", r.Host)
	tmpl, err := template.ParseFiles("html/login.html")
	if err != nil {
		log.Printf("error parsing template: %s", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		tmpl.Execute(w, msg{Err: ""})
		return
	} else if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Printf("error parsing form: %s", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	name := r.Form.Get("name")
	name, err = ValidateName(name)
	if err != nil {
		_ = tmpl.Execute(w, msg{Err: "name can not be empty"})
		return
	}

	err = services.ChatService.AddUser(redisdb.RedisClient, name)
	if err != nil {
		if errors.Is(err, services.UserExistsError) {
			_ = tmpl.Execute(w, msg{Err: "name already occupied"})
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		return
	}
	url := fmt.Sprintf("/chat?name=%s", name)
	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}

func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", NameError
	}
	return name, nil
}
