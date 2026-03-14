package main

import (
	"log"
	"net/http"
	"os"
	"time"
	"zai2api-go/auth"
	"zai2api-go/common"
	"zai2api-go/config"
	"zai2api-go/database"
	"zai2api-go/router"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	database.Init(cfg)
	auth.Init(cfg)
	common.StartDailyResetScheduler()

	r := router.Setup(cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on http://localhost:%s", port)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
