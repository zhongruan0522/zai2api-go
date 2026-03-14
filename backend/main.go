package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"zai2api-go/auth"
	"zai2api-go/config"
	"zai2api-go/database"
	"zai2api-go/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	database.Init(cfg)

	// 初始化认证模块
	auth.Init(cfg)

	// 初始化 OCR 处理器
	ocrHandler := handlers.NewOCRHandler(cfg)

	r := gin.Default()

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// OCR 对外接口（需要 API Key 认证）
	ocr := r.Group("/ocr/v1")
	{
		ocr.POST("/files/ocr", ocrHandler.ProcessOCR)
	}

	// API 路由组
	api := r.Group("/api")
	{
		api.GET("/hello", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Hello from Go backend!"})
		})

		// 登录接口（无需认证）
		api.POST("/login", auth.LoginHandler)

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(auth.AuthMiddleware())
		{
			protected.GET("/me", func(c *gin.Context) {
				username := c.GetString("username")
				c.JSON(http.StatusOK, gin.H{"username": username})
			})

			// Audio Token 管理
			protected.GET("/tokens/audio", handlers.GetAudioTokens)
			protected.POST("/tokens/audio", handlers.CreateAudioTokens)
			protected.DELETE("/tokens/audio/:id", handlers.DeleteAudioToken)
			protected.PUT("/tokens/audio/:id/toggle", handlers.ToggleAudioToken)
			protected.POST("/tokens/audio/batch-delete", handlers.BatchDeleteAudioTokens)
			protected.POST("/tokens/audio/batch-toggle", handlers.BatchToggleAudioTokens)

			// OCR Token 管理
			protected.GET("/tokens/ocr", handlers.GetOCRTokens)
			protected.POST("/tokens/ocr", handlers.CreateOCRTokens)
			protected.DELETE("/tokens/ocr/:id", handlers.DeleteOCRToken)
			protected.PUT("/tokens/ocr/:id/toggle", handlers.ToggleOCRToken)
			protected.POST("/tokens/ocr/batch-delete", handlers.BatchDeleteOCRTokens)
			protected.POST("/tokens/ocr/batch-toggle", handlers.BatchToggleOCRTokens)

			// Chat Token 管理
			protected.GET("/tokens/chat", handlers.GetChatTokens)
			protected.POST("/tokens/chat", handlers.CreateChatTokens)
			protected.DELETE("/tokens/chat/:id", handlers.DeleteChatToken)
			protected.PUT("/tokens/chat/:id/toggle", handlers.ToggleChatToken)
			protected.POST("/tokens/chat/batch-delete", handlers.BatchDeleteChatTokens)
			protected.POST("/tokens/chat/batch-toggle", handlers.BatchToggleChatTokens)

			// API Key 管理
			protected.GET("/apikeys", handlers.GetAPIKeys)
			protected.POST("/apikeys", handlers.CreateAPIKey)
			protected.DELETE("/apikeys/:id", handlers.DeleteAPIKey)
			protected.PUT("/apikeys/:id/toggle", handlers.ToggleAPIKey)
			protected.POST("/apikeys/batch-delete", handlers.BatchDeleteAPIKeys)
			protected.POST("/apikeys/batch-toggle", handlers.BatchToggleAPIKeys)

			// 请求日志
			protected.GET("/logs", handlers.GetRequestLogs)
			protected.GET("/logs/stats", handlers.GetRequestLogStats)
		}
	}

	// 前端静态文件服务
	frontendDir := filepath.Join(".", "frontend")
	if _, err := os.Stat(frontendDir); err == nil {
		// 静态资源（js/css/images 等）
		r.Static("/_next", filepath.Join(frontendDir, "_next"))
		if publicDir := filepath.Join(frontendDir, "public"); _, err := os.Stat(publicDir); err == nil {
			r.Static("/public", publicDir)
		}

		// SPA 回退：所有未匹配的路由返回 index.html
		indexFile := filepath.Join(frontendDir, "index.html")
		r.NoRoute(func(c *gin.Context) {
			// 非 GET 请求或路径看起来像 API 调用，返回 404
			if c.Request.Method != "GET" {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.File(indexFile)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
