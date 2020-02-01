package app

import (
	"github.com/go-redis/redis/v7"
	"net/http"
	"os"
	"os/signal"
	"simple_chat/src/config"
	"simple_chat/src/controllers"
	"simple_chat/src/datasources/redisdb"
	"simple_chat/src/hub/hub"
	"simple_chat/src/logger"
	"syscall"
)

const (
	port = ":8080"
)

func RunApp() {
	if err := config.Load(); err != nil {
		panic(err)
	}

	if err := logger.InitLogger(config.Config.ServiceName, config.Config.LogLevel); err != nil {
		panic(err)
	}

	logger.Logger.Debug("logger initialized")

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

	logger.Logger.Infof("start listening on port %s", port)

	go hub.Hub.Run()

	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			panic(err)
		}
	}()

	<-done
	hub.Hub.ShutDown()
	logger.Logger.Info("shutting down gracefully")
	logger.Logger.Sync()
}
