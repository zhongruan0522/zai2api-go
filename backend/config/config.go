package config

import (
	"os"
	"strconv"

	"gorm.io/gorm"
)

type Config struct {
	AdminUsername string
	AdminPassword string
	JWTSecret     string
	DB            *gorm.DB
	OCRDailyLimit int // OCR Token 默认每日限额，0 表示无限制
}

func Load() *Config {
	return &Config{
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		OCRDailyLimit: getEnvInt("OCR_DAILY_LIMIT", 0),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
