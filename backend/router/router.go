package router

import (
	"net/http"
	"os"
	"path/filepath"
	"zai2api-go/auth"
	"zai2api-go/config"
	"zai2api-go/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	ocrHandler := handlers.NewOCRHandler(cfg)
	ocr := r.Group("/ocr/v1")
	{
		ocr.POST("/files/ocr", ocrHandler.ProcessOCR)
	}

	api := r.Group("/api")
	{
		api.GET("/hello", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Hello from Go backend!"})
		})

		api.POST("/login", auth.LoginHandler)

		protected := api.Group("")
		protected.Use(auth.AuthMiddleware())
		{
			protected.GET("/me", func(c *gin.Context) {
				username := c.GetString("username")
				c.JSON(http.StatusOK, gin.H{"username": username})
			})

			registerTokenRoutes(protected)
			registerAPIKeyRoutes(protected)
			registerLogRoutes(protected)
		}
	}

	serveFrontend(r)

	return r
}

func registerTokenRoutes(rg *gin.RouterGroup) {
	rg.GET("/tokens/audio", handlers.GetAudioTokens)
	rg.POST("/tokens/audio", handlers.CreateAudioTokens)
	rg.DELETE("/tokens/audio/:id", handlers.DeleteAudioToken)
	rg.PUT("/tokens/audio/:id/toggle", handlers.ToggleAudioToken)
	rg.POST("/tokens/audio/batch-delete", handlers.BatchDeleteAudioTokens)
	rg.POST("/tokens/audio/batch-toggle", handlers.BatchToggleAudioTokens)

	rg.GET("/tokens/ocr", handlers.GetOCRTokens)
	rg.POST("/tokens/ocr", handlers.CreateOCRTokens)
	rg.DELETE("/tokens/ocr/:id", handlers.DeleteOCRToken)
	rg.PUT("/tokens/ocr/:id/toggle", handlers.ToggleOCRToken)
	rg.POST("/tokens/ocr/batch-delete", handlers.BatchDeleteOCRTokens)
	rg.POST("/tokens/ocr/batch-toggle", handlers.BatchToggleOCRTokens)

	rg.GET("/tokens/chat", handlers.GetChatTokens)
	rg.POST("/tokens/chat", handlers.CreateChatTokens)
	rg.DELETE("/tokens/chat/:id", handlers.DeleteChatToken)
	rg.PUT("/tokens/chat/:id/toggle", handlers.ToggleChatToken)
	rg.POST("/tokens/chat/batch-delete", handlers.BatchDeleteChatTokens)
	rg.POST("/tokens/chat/batch-toggle", handlers.BatchToggleChatTokens)

	rg.GET("/tokens/image", handlers.GetImageTokens)
	rg.POST("/tokens/image", handlers.CreateImageTokens)
	rg.DELETE("/tokens/image/:id", handlers.DeleteImageToken)
	rg.PUT("/tokens/image/:id/toggle", handlers.ToggleImageToken)
	rg.POST("/tokens/image/batch-delete", handlers.BatchDeleteImageTokens)
	rg.POST("/tokens/image/batch-toggle", handlers.BatchToggleImageTokens)
}

func registerAPIKeyRoutes(rg *gin.RouterGroup) {
	rg.GET("/apikeys", handlers.GetAPIKeys)
	rg.POST("/apikeys", handlers.CreateAPIKey)
	rg.DELETE("/apikeys/:id", handlers.DeleteAPIKey)
	rg.PUT("/apikeys/:id/toggle", handlers.ToggleAPIKey)
	rg.POST("/apikeys/batch-delete", handlers.BatchDeleteAPIKeys)
	rg.POST("/apikeys/batch-toggle", handlers.BatchToggleAPIKeys)
}

func registerLogRoutes(rg *gin.RouterGroup) {
	rg.GET("/logs", handlers.GetRequestLogs)
	rg.GET("/logs/stats", handlers.GetRequestLogStats)
}

func serveFrontend(r *gin.Engine) {
	frontendDir := filepath.Join(".", "frontend")
	if _, err := os.Stat(frontendDir); err != nil {
		return
	}

	r.Static("/_next", filepath.Join(frontendDir, "_next"))
	publicDir := filepath.Join(frontendDir, "public")
	if _, err := os.Stat(publicDir); err == nil {
		r.Static("/public", publicDir)
	}

	indexFile := filepath.Join(frontendDir, "index.html")
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		path := c.Request.URL.Path
		htmlFile := filepath.Join(frontendDir, path+".html")
		if info, err := os.Stat(htmlFile); err == nil && !info.IsDir() {
			c.File(htmlFile)
			return
		}
		c.File(indexFile)
	})
}
