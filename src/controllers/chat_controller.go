package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"simple_chat/src/config"
	"simple_chat/src/datasources/redisdb"
	"simple_chat/src/domains/message_domain"
	"simple_chat/src/domains/request_domains"
	"simple_chat/src/hub/hub"
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
	GetMessages(w http.ResponseWriter, r *http.Request)
	ServeChatHTML(w http.ResponseWriter, r *http.Request)
}

type chatController struct{}

func (c *chatController) GetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed ", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var requestData request_domains.GetOldMessagesRequest
	err := decoder.Decode(&requestData)
	if err != nil {
		log.Print("unable to decode message")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	messages, err := services.ChatService.GetFromToMessages(redisdb.RedisClient, requestData.From, requestData.To)
	if err != nil {
		log.Println("unable to get old messages from redis")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(messages)
	if err != nil {
		log.Println("unable to marshal messages")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

func (c *chatController) HandleChat(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	log.Println(queryParams)
	values, ok := queryParams["name"]
	if !ok || len(values) < 1 {
		log.Print("no name for chat provided")
		http.Error(w, "chat name should be provided", http.StatusBadRequest)
		return
	}
	name := values[0]

	isFree, err := services.ChatService.UsernameIsFree(redisdb.RedisClient, name)
	if !isFree || err != nil {
		log.Printf("name %s is occupied", name)
		http.Error(w, "name is occupied", http.StatusBadRequest)
		return
	}

	log.Printf("creating websocket connection with %s", r.Host)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("unable to upgrade connecion. err: %s", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	hub.Hub.AddClientCh <- hub.UserConnection{
		Name: name,
		Conn: conn,
	}

	for {
		msg := message_domain.Message{}
		err = conn.ReadJSON(&msg)
		if err != nil {
			err, ok := err.(*websocket.CloseError)
			if ok {
				hub.Hub.RemoveClientCh <- name
				return
			}
			log.Printf("unble to read message. err: %s", err.Error())
		}
		log.Printf("received messae: %s", msg.String())

		err := services.ChatService.PublishMessage(redisdb.RedisClient, msg.String())
		if err != nil {
			log.Printf("error publishinbg message: %s", err)
		}
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

	isFree, err := services.ChatService.UsernameIsFree(redisdb.RedisClient, name)
	if !isFree {
		_ = tmpl.Execute(w, msg{Err: "name already occupied"})
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	url := fmt.Sprintf("/chat?name=%s", name)
	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}

func (c *chatController) ServeChatHTML(w http.ResponseWriter, r *http.Request) {
	log.Printf("serving chat.html")

	tmpl, err := template.ParseFiles("html/chat.html")
	if err != nil {
		log.Printf("error parsing template: %s", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, config.Config)
}

func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", NameError
	}
	return name, nil
}
