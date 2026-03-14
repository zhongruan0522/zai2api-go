package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "admin123"
	defaultJWTSecret     = "your-secret-key-change-in-production"
)

type Config struct {
	AdminUsername string
	AdminPassword string
	JWTSecret     string
	DB            *gorm.DB
	OCRDailyLimit int // OCR Token 默认每日限额，0 表示无限制

	GinMode               string
	AppEnv                string
	AllowInsecureDefaults bool

	CORSAllowOrigins []string
	TrustedProxies   []string

	AdminMaxBodyBytes    int64
	ImageMaxBodyBytes    int64
	OCRMaxBodyBytes      int64
	UpstreamMaxRespBytes int64
}

func Load() *Config {
	ginMode := getEnv("GIN_MODE", "debug")
	appEnv := getEnv("APP_ENV", "")
	isProd := strings.EqualFold(strings.TrimSpace(appEnv), "production") || strings.EqualFold(strings.TrimSpace(ginMode), "release")

	corsDefault := ""
	if !isProd {
		corsDefault = "http://localhost:3000,http://127.0.0.1:3000"
	}

	cfg := &Config{
		AdminUsername: getEnv("ADMIN_USERNAME", defaultAdminUsername),
		AdminPassword: getEnv("ADMIN_PASSWORD", defaultAdminPassword),
		JWTSecret:     getEnv("JWT_SECRET", defaultJWTSecret),
		OCRDailyLimit: getEnvInt("OCR_DAILY_LIMIT", 0),

		GinMode:               ginMode,
		AppEnv:                appEnv,
		AllowInsecureDefaults: getEnvBool("ALLOW_INSECURE_DEFAULTS", false),

		CORSAllowOrigins: getEnvCSV("CORS_ALLOW_ORIGINS", corsDefault),
		TrustedProxies:   getEnvCSV("TRUSTED_PROXIES", ""),

		AdminMaxBodyBytes:    getEnvInt64("ADMIN_MAX_BODY_BYTES", 2<<20),
		ImageMaxBodyBytes:    getEnvInt64("IMAGE_MAX_BODY_BYTES", 1<<20),
		OCRMaxBodyBytes:      getEnvInt64("OCR_MAX_BODY_BYTES", 25<<20),
		UpstreamMaxRespBytes: getEnvInt64("UPSTREAM_MAX_RESPONSE_BYTES", 10<<20),
	}

	return cfg
}

func (c *Config) IsProduction() bool {
	if strings.EqualFold(strings.TrimSpace(c.AppEnv), "production") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(c.GinMode), "release")
}

func (c *Config) Validate() error {
	if c.AllowInsecureDefaults || !c.IsProduction() {
		return nil
	}

	if c.JWTSecret == defaultJWTSecret || len(strings.TrimSpace(c.JWTSecret)) < 32 {
		return fmt.Errorf("insecure JWT_SECRET: set JWT_SECRET to a random 32+ character value (or set ALLOW_INSECURE_DEFAULTS=true for dev)")
	}
	if c.AdminPassword == defaultAdminPassword || len(strings.TrimSpace(c.AdminPassword)) < 12 {
		return fmt.Errorf("insecure ADMIN_PASSWORD: set ADMIN_PASSWORD to a strong password (>= 12 chars) (or set ALLOW_INSECURE_DEFAULTS=true for dev)")
	}

	return nil
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

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		switch strings.ToLower(value) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return defaultValue
}

func getEnvCSV(key, defaultValue string) []string {
	raw := strings.TrimSpace(getEnv(key, defaultValue))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
