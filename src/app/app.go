package app

import (
	"github.com/go-redis/redis/v7"
	"log"
	"net/http"
	"simple_chat/src/controllers"
	"simple_chat/src/datasources/redisdb"
)

const (
	port = ":8080"
)

func RunApp() {
	opt := &redis.Options{
		Addr: "localhost:6379",
	}
	if err := redisdb.InitRedisClient(opt); err != nil {
		panic(err)
	}

	http.HandleFunc("/", controllers.FileController.ServeHTML)
	http.HandleFunc("/chat", controllers.ChatController.HandleChat)

	log.Printf("start listening on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(err)
	}
}
