package models

import "time"

type BaseLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	RequestID string    `json:"request_id" gorm:"index;size:36;not null"`
	CreatedAt time.Time `json:"created_at"`
	SourceIP  string    `json:"source_ip" gorm:"size:45"`
	APIKeyID  uint      `json:"api_key_id"`
	TokenID   uint      `json:"token_id"`
	Success   bool      `json:"success"`
	ErrorCode string    `json:"error_code" gorm:"size:20"`
	ErrorMsg  string    `json:"error_msg" gorm:"size:500"`
}

type OCRLog struct {
	BaseLog
}

func (OCRLog) TableName() string {
	return "ocr_log"
}

type ChatLog struct {
	BaseLog
}

func (ChatLog) TableName() string {
	return "chat_log"
}

type ImageLog struct {
	BaseLog
}

func (ImageLog) TableName() string {
	return "image_log"
}
