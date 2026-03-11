package database

import (
	"fmt"
	"log"
	"os"
	"zai2api-go/config"
	"zai2api-go/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) {
	var err error

	dsn := buildDSN()
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// 自动迁移
	if err = DB.AutoMigrate(
		&models.User{},
		&models.AudioToken{},
		&models.OCRToken{},
		&models.ChatToken{},
		&models.APIKey{},
		&models.RequestLog{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 同步管理员账户
	syncAdminUser(cfg)
	log.Println("Database initialized successfully")
}

func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "zai2api")
	timezone := getEnv("DB_TIMEZONE", "Asia/Shanghai")

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		host, user, password, dbname, port, timezone,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func syncAdminUser(cfg *config.Config) {
	var user models.User
	result := DB.Where("username = ?", cfg.AdminUsername).First(&user)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	if result.Error == gorm.ErrRecordNotFound {
		// 创建新用户
		user = models.User{
			Username: cfg.AdminUsername,
			Password: string(hashedPassword),
		}
		if err = DB.Create(&user).Error; err != nil {
			log.Fatal("Failed to create admin user:", err)
		}
		log.Printf("Admin user '%s' created", cfg.AdminUsername)
	} else {
		// 更新密码（允许通过环境变量修改密码）
		user.Password = string(hashedPassword)
		if err = DB.Save(&user).Error; err != nil {
			log.Fatal("Failed to update admin user:", err)
		}
		log.Printf("Admin user '%s' synced", cfg.AdminUsername)
	}
}
