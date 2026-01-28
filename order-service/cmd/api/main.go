package main

import (
	"order-service/internal/config"
	"order-service/internal/logger"
	"order-service/internal/server"
)

func main() {
	cfg := config.NewConfig()
	log := logger.NewLogger(cfg.Logger)

	log.Info("Initializing server")
	httpServer := server.NewServer(cfg, log)

	log.Info("Starting server on :" + cfg.Port)
	if err := httpServer.Start(); err != nil {
		log.Error("Server failed to start", "error", err)
		panic(err)
	}
}
