package models

import (
	"time"

	"gorm.io/gorm"
)

// ServiceType 服务类型
type ServiceType string

const (
	ServiceOCR   ServiceType = "ocr"
	ServiceAudio ServiceType = "audio"
	ServiceChat  ServiceType = "chat"
	ServiceImage ServiceType = "image"
)

// AudioToken 音频服务 Token
type AudioToken struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Token          string         `json:"token" gorm:"uniqueIndex;size:1024;not null"`
	ImportedAt     time.Time      `json:"imported_at" gorm:"not null"`
	LastUsedAt     *time.Time     `json:"last_used_at"`
	Enabled        bool           `json:"enabled" gorm:"default:true"`
	TotalCallCount int64          `json:"total_call_count" gorm:"default:0"`
	DailyCallCount int64          `json:"daily_call_count" gorm:"default:0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// OCRToken OCR 服务 Token
type OCRToken struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Token          string         `json:"token" gorm:"uniqueIndex;size:1024;not null"`
	ImportedAt     time.Time      `json:"imported_at" gorm:"not null"`
	LastUsedAt     *time.Time     `json:"last_used_at"`
	Enabled        bool           `json:"enabled" gorm:"default:true"`
	TotalCallCount int64          `json:"total_call_count" gorm:"default:0"`
	DailyCallCount int64          `json:"daily_call_count" gorm:"default:0"`
	DailyLimit     int64          `json:"daily_limit" gorm:"default:0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// ChatToken 聊天服务 Token
type ChatToken struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Token          string         `json:"token" gorm:"uniqueIndex;size:1024;not null"`
	ImportedAt     time.Time      `json:"imported_at" gorm:"not null"`
	LastUsedAt     *time.Time     `json:"last_used_at"`
	Enabled        bool           `json:"enabled" gorm:"default:true"`
	TotalCallCount int64          `json:"total_call_count" gorm:"default:0"`
	DailyCallCount int64          `json:"daily_call_count" gorm:"default:0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// ImageToken 图像服务 Token
type ImageToken struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Token          string         `json:"token" gorm:"uniqueIndex;size:1024;not null"`
	ImportedAt     time.Time      `json:"imported_at" gorm:"not null"`
	LastUsedAt     *time.Time     `json:"last_used_at"`
	Enabled        bool           `json:"enabled" gorm:"default:true"`
	TotalCallCount int64          `json:"total_call_count" gorm:"default:0"`
	DailyCallCount int64          `json:"daily_call_count" gorm:"default:0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// APIKey 统一 API 密钥表
type APIKey struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Key       string         `json:"key" gorm:"uniqueIndex;size:64;not null"`     // API 密钥值
	Services  string         `json:"services" gorm:"size:50;not null;default:''"` // 服务类型：ocr,audio,chat,image 或组合
	CreatedAt time.Time      `json:"created_at"`
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (AudioToken) TableName() string {
	return "audio_token"
}

func (OCRToken) TableName() string {
	return "ocr_token"
}

func (ChatToken) TableName() string {
	return "chat_token"
}

func (ImageToken) TableName() string {
	return "image_token"
}

func (APIKey) TableName() string {
	return "api_key"
}
