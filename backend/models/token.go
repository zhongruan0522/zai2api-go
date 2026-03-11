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
)

// AudioToken 音频服务 Token
type AudioToken struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Token           string         `json:"token" gorm:"uniqueIndex;size:255;not null"`
	ImportedAt      time.Time      `json:"imported_at" gorm:"not null"`        // 导入日期（带时间）
	LastUsedAt      *time.Time     `json:"last_used_at"`                       // 最后使用日期（仅日期）
	Enabled         bool           `json:"enabled" gorm:"default:true"`        // 是否启用
	TotalCallCount  int            `json:"total_call_count" gorm:"default:0"`  // 总共调用次数
	DailyCallCount  int            `json:"daily_call_count" gorm:"default:0"`  // 当日调用次数
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// OCRToken OCR 服务 Token
type OCRToken struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Token           string         `json:"token" gorm:"uniqueIndex;size:255;not null"`
	ImportedAt      time.Time      `json:"imported_at" gorm:"not null"`
	LastUsedAt      *time.Time     `json:"last_used_at"`
	Enabled         bool           `json:"enabled" gorm:"default:true"`
	TotalCallCount  int            `json:"total_call_count" gorm:"default:0"`
	DailyCallCount  int            `json:"daily_call_count" gorm:"default:0"`
	DailyLimit      int            `json:"daily_limit" gorm:"default:0"`        // 每日限额，0表示无限制
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// ChatToken 聊天服务 Token
type ChatToken struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Token           string         `json:"token" gorm:"uniqueIndex;size:255;not null"`
	ImportedAt      time.Time      `json:"imported_at" gorm:"not null"`
	LastUsedAt      *time.Time     `json:"last_used_at"`
	Enabled         bool           `json:"enabled" gorm:"default:true"`
	TotalCallCount  int            `json:"total_call_count" gorm:"default:0"`
	DailyCallCount  int            `json:"daily_call_count" gorm:"default:0"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// APIKey 统一 API 密钥表
type APIKey struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Key         string         `json:"key" gorm:"uniqueIndex;size:64;not null"`      // API 密钥值
	Services    string         `json:"services" gorm:"size:50;not null;default:''"` // 服务类型：ocr,audio,chat 或组合
	CreatedAt   time.Time      `json:"created_at"`
	Enabled     bool           `json:"enabled" gorm:"default:true"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// RequestLog 请求日志表
type RequestLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	RequestID  string    `json:"request_id" gorm:"index;size:36;not null"` // 请求唯一ID
	CreatedAt  time.Time `json:"created_at"`                               // 请求时间
	Channel    string    `json:"channel" gorm:"size:20;not null"`           // 渠道：ocr, audio, chat
	SourceIP   string    `json:"source_ip" gorm:"size:45"`                  // 源IP地址
	APIKeyID   uint      `json:"api_key_id"`                                // 对应的 API Key ID
	TokenID    uint      `json:"token_id"`                                  // 使用的上游 Token ID
	Success    bool      `json:"success"`                                   // 是否成功
	ErrorCode  string    `json:"error_code" gorm:"size:20"`                 // 错误码
	ErrorMsg   string    `json:"error_msg" gorm:"size:500"`                 // 错误信息
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

func (APIKey) TableName() string {
	return "api_key"
}

func (RequestLog) TableName() string {
	return "request_log"
}
