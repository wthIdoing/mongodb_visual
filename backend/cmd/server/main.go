package main

import (
	"log"

	"mongoDB_visual/backend/internal/config"
	apphttp "mongoDB_visual/backend/internal/http"
	"mongoDB_visual/backend/internal/mongodb"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pool := mongodb.NewPool(mongodb.DefaultConnectionConfig(cfg))
	defer func() {
		if disconnectErr := pool.DisconnectAll(); disconnectErr != nil {
			log.Printf("disconnect mongodb: %v", disconnectErr)
		}
	}()

	router := apphttp.NewServer(cfg, pool)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
