package router

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"zai2api-go/auth"
	"zai2api-go/config"
	"zai2api-go/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatalf("invalid TRUSTED_PROXIES: %v", err)
	}

	if len(cfg.CORSAllowOrigins) > 0 {
		corsCfg := cors.Config{
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: false,
			MaxAge:           12 * time.Hour,
		}
		for _, o := range cfg.CORSAllowOrigins {
			if o == "*" {
				corsCfg.AllowAllOrigins = true
				corsCfg.AllowOrigins = nil
				break
			}
		}
		if !corsCfg.AllowAllOrigins {
			corsCfg.AllowOrigins = cfg.CORSAllowOrigins
		}
		r.Use(cors.New(corsCfg))
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	ocrHandler := handlers.NewOCRHandler(cfg)
	ocr := r.Group("/ocr/v1")
	{
		ocr.Use(limitBodySize(cfg.OCRMaxBodyBytes))
		ocr.POST("/files/ocr", ocrHandler.ProcessOCR)
	}

	imageHandler := handlers.NewImageHandler()
	image := r.Group("/image/v1")
	{
		image.Use(limitBodySize(cfg.ImageMaxBodyBytes))
		image.POST("/images/generations", imageHandler.GenerateImage)
	}
	imageChat := r.Group("/v1")
	{
		imageChat.Use(limitBodySize(cfg.ImageMaxBodyBytes))
		imageChat.POST("/chat/completions", imageHandler.ChatGenerateImage)
	}

	api := r.Group("/api")
	{
		api.Use(limitBodySize(cfg.AdminMaxBodyBytes))
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
	rg.GET("/logs/ocr", handlers.GetOCRLogs)
	rg.GET("/logs/ocr/stats", handlers.GetOCRLogStats)
	rg.GET("/logs/chat", handlers.GetChatLogs)
	rg.GET("/logs/chat/stats", handlers.GetChatLogStats)
	rg.GET("/logs/image", handlers.GetImageLogs)
	rg.GET("/logs/image/stats", handlers.GetImageLogStats)

	rg.GET("/logs/monitor/summary", handlers.GetMonitorSummary)
	rg.GET("/logs/monitor/daily", handlers.GetMonitorDaily)
	rg.GET("/logs/monitor/hourly", handlers.GetMonitorHourly)
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
	if _, err := os.Stat(indexFile); err != nil {
		return
	}
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		reqPath := c.Request.URL.Path
		rel := strings.TrimPrefix(path.Clean(reqPath), "/")
		if rel == "" || rel == "." {
			c.File(indexFile)
			return
		}
		// Basic hardening against path tricks on Windows-style paths.
		if strings.Contains(rel, "\\") || strings.Contains(rel, ":") {
			c.File(indexFile)
			return
		}

		candidates := []string{
			filepath.Join(frontendDir, rel),
			filepath.Join(frontendDir, rel, "index.html"),
			filepath.Join(frontendDir, rel+".html"),
		}
		for _, f := range candidates {
			if info, err := os.Stat(f); err == nil && !info.IsDir() {
				c.File(f)
				return
			}
		}
		c.File(indexFile)
	})
}
