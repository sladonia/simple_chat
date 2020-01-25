package app

import (
	"log"
	"net/http"
	"simple_chat/src/controllers"
)

const (
	port = ":8080"
)

func RunApp() {
	http.HandleFunc("/", controllers.FileController.ServeHTML)
	http.HandleFunc("/chat", controllers.ChatController.HandleChat)

	log.Printf("start listening on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(err)
	}
}
