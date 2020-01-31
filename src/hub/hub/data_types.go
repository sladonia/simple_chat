package hub

import "github.com/gorilla/websocket"

type UserConnection struct {
	Name string
	Conn *websocket.Conn
}
