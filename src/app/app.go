package app

import (
	"github.com/go-redis/redis/v7"
	"log"
	"net/http"
	"os"
	"os/signal"
	"simple_chat/src/config"
	"simple_chat/src/controllers"
	"simple_chat/src/datasources/redisdb"
	"simple_chat/src/hub/hub"
	"simple_chat/src/services"
	"syscall"
)

const (
	port = ":8080"
)

func RunApp() {
	if err := config.Load(); err != nil {
		panic(err)
	}

	opt := &redis.Options{
		Addr:     config.Config.RedisConfig.Address,
		PoolSize: config.Config.RedisConfig.PoolSize,
	}
	if err := redisdb.InitRedisClient(opt); err != nil {
		panic(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	http.HandleFunc("/", controllers.ChatController.HandleLogIn)
	http.HandleFunc("/chat", controllers.ChatController.ServeChatHTML)
	http.HandleFunc("/chat/get-messages", controllers.ChatController.GetMessages)
	http.HandleFunc("/chat-sock", controllers.ChatController.HandleChat)

	log.Printf("start listening on port %s", port)

	go hub.Hub.Run()

	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			panic(err)
		}
	}()

	<-done
	redisdb.RedisClient.Del(services.UsersSet)
	log.Printf("Shutting down gracefully...")
}
