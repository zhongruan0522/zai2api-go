package main

import (
	"log"
	"os"
	"zai2api-go/auth"
	"zai2api-go/common"
	"zai2api-go/config"
	"zai2api-go/database"
	"zai2api-go/router"
)

func main() {
	cfg := config.Load()

	database.Init(cfg)
	auth.Init(cfg)
	common.StartDailyResetScheduler()

	r := router.Setup(cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
