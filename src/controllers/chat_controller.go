package controllers

import "net/http"

var (
	ChatController ChatControllerInterface = &chatController{}
)

type ChatControllerInterface interface {
	HandleChat(w http.ResponseWriter, r *http.Request)
}

type chatController struct{}

func (c *chatController) HandleChat(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
