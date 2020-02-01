package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"io/ioutil"
	"net/http"
	"simple_chat/src/config"
	"simple_chat/src/datasources/redisdb"
	"simple_chat/src/domains/message_domain"
	"simple_chat/src/domains/request_domains"
	"simple_chat/src/hub/hub"
	"simple_chat/src/logger"
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
	logger.Logger.Infow("processing request", "method", r.Method, "path", r.URL.Path)
	if r.Method != "POST" {
		http.Error(w, "method not allowed ", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var requestData request_domains.GetOldMessagesRequest
	err := decoder.Decode(&requestData)
	if err != nil {
		bytes, _ := ioutil.ReadAll(r.Body)
		logger.Logger.Errorw("unable to parse request data", "data", string(bytes))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	messages, err := services.ChatService.GetFromToMessages(redisdb.RedisClient, requestData.From, requestData.To)
	if err != nil {
		logger.Logger.Error("unable to get old messages from redis")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(messages)
	if err != nil {
		logger.Logger.Errorw("unable to marshal messages", "data", messages)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		logger.Logger.Errorw("error writing response", "err", err)
	}
}

func (c *chatController) HandleChat(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("processing request", "method", r.Method, "path", r.URL.Path)
	queryParams := r.URL.Query()
	values, ok := queryParams["name"]
	if !ok || len(values) < 1 {
		logger.Logger.Debug("no username for provided")
		http.Error(w, "username should be provided", http.StatusBadRequest)
		return
	}
	name := values[0]

	isFree, err := services.ChatService.UsernameIsFree(redisdb.RedisClient, name)
	if !isFree || err != nil {
		logger.Logger.Debugf("username \"%s\" is occupied", name)
		http.Error(w, "username is occupied", http.StatusBadRequest)
		return
	}

	logger.Logger.Infof("creating websocket connection. host: %s. username: %s", r.Host, name)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.Errorw("unable to upgrade connection", "err", err.Error())
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
			logger.Logger.Errorw("unable to read message", "err", err.Error())
		}
		logger.Logger.Debugf("received message: %s", msg.String())

		err := services.ChatService.PublishMessage(redisdb.RedisClient, msg.String())
		if err != nil {
			logger.Logger.Errorw("error publishing message", "err", err)
		}
	}
}

func (c *chatController) HandleLogIn(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("processing request", "method", r.Method, "path", r.URL.Path)
	type msg struct {
		Err string
	}

	tmpl, err := template.ParseFiles("html/login.html")
	if err != nil {
		logger.Logger.Errorw("error parsing template", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		tmpl.Execute(w, msg{Err: ""})
		return
	} else if r.Method != "POST" {
		logger.Logger.Debug("method not allowed. returning 405")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.Logger.Errorw("error parsing form", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	name := r.Form.Get("name")
	name, err = ValidateName(name)
	if err != nil {
		err = tmpl.Execute(w, msg{Err: "name can not be empty"})
		if err != nil {
			logger.Logger.Errorw("template error", "err", err)
		}
		return
	}

	isFree, err := services.ChatService.UsernameIsFree(redisdb.RedisClient, name)
	if !isFree {
		err = tmpl.Execute(w, msg{Err: "name already occupied"})
		if err != nil {
			logger.Logger.Errorw("template error", "err", err)
		}
		return
	}
	if err != nil {
		logger.Logger.Errorw("redis error", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	url := fmt.Sprintf("/chat?name=%s", name)
	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}

func (c *chatController) ServeChatHTML(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("processing request", "method", r.Method, "path", r.URL.Path)

	tmpl, err := template.ParseFiles("html/chat.html")
	if err != nil {
		logger.Logger.Errorw("template error", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, config.Config)
	if err != nil {
		logger.Logger.Errorw("template error", "err", err)
	}
}

func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", NameError
	}
	return name, nil
}
