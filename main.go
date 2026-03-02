package main

import (
	"fmt"
	"net/http"
	"os"

	"zai2api-go/audio"
	"zai2api-go/chat"
	"zai2api-go/chatagent"
	"zai2api-go/image"
	"zai2api-go/ocr"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 注册各模块路由
	// 绘图模块: /image/v1/*
	image.RegisterRoutes(r.Group("/image/v1"))

	// 音频模块: /audio/v1/*
	audio.RegisterRoutes(r.Group("/audio/v1"))

	// OCR模块: /ocr/v1/*
	ocr.RegisterRoutes(r.Group("/ocr/v1"))

	// 聊天模块: /chat/v1/*
	chat.RegisterRoutes(r.Group("/chat/v1"))

	// Chat-Agent模块: /chat-agent/v1/*
	chatagent.RegisterRoutes(r.Group("/chat-agent/v1"))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
