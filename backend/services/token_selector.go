package services

import (
	"math/rand"
	"time"
	"zai2api-go/database"
	"zai2api-go/models"

	"gorm.io/gorm"
)

// TokenSelector Token 选择器
type TokenSelector struct{}

// NewTokenSelector 创建 Token 选择器
func NewTokenSelector() *TokenSelector {
	return &TokenSelector{}
}

// SelectOCRToken 随机选择一个可用的 OCR Token
func (s *TokenSelector) SelectOCRToken() (*models.OCRToken, error) {
	var tokens []models.OCRToken

	// 查询所有启用的 Token
	if err := database.DB.Where("enabled = ?", true).Find(&tokens).Error; err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// 过滤出当日未超限的 Token
	var availableTokens []models.OCRToken
	for _, token := range tokens {
		// 如果 DailyLimit 为 0 表示无限制
		if token.DailyLimit == 0 || token.DailyCallCount < token.DailyLimit {
			availableTokens = append(availableTokens, token)
		}
	}

	if len(availableTokens) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// 随机选择一个
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	selected := availableTokens[r.Intn(len(availableTokens))]

	return &selected, nil
}

// IncrementOCRCallCount 增加 OCR Token 的调用计数
func (s *TokenSelector) IncrementOCRCallCount(tokenID uint) error {
	now := time.Now()
	return database.DB.Model(&models.OCRToken{}).Where("id = ?", tokenID).Updates(map[string]interface{}{
		"total_call_count": gorm.Expr("total_call_count + 1"),
		"daily_call_count": gorm.Expr("daily_call_count + 1"),
		"last_used_at":     &now,
	}).Error
}

// SelectAudioToken 随机选择一个可用的 Audio Token
func (s *TokenSelector) SelectAudioToken() (*models.AudioToken, error) {
	var tokens []models.AudioToken

	if err := database.DB.Where("enabled = ?", true).Find(&tokens).Error; err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	selected := tokens[r.Intn(len(tokens))]

	return &selected, nil
}

// IncrementAudioCallCount 增加 Audio Token 的调用计数
func (s *TokenSelector) IncrementAudioCallCount(tokenID uint) error {
	now := time.Now()
	return database.DB.Model(&models.AudioToken{}).Where("id = ?", tokenID).Updates(map[string]interface{}{
		"total_call_count": gorm.Expr("total_call_count + 1"),
		"daily_call_count": gorm.Expr("daily_call_count + 1"),
		"last_used_at":     &now,
	}).Error
}

// SelectChatToken 随机选择一个可用的 Chat Token
func (s *TokenSelector) SelectChatToken() (*models.ChatToken, error) {
	var tokens []models.ChatToken

	if err := database.DB.Where("enabled = ?", true).Find(&tokens).Error; err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	selected := tokens[r.Intn(len(tokens))]

	return &selected, nil
}

// IncrementChatCallCount 增加 Chat Token 的调用计数
func (s *TokenSelector) IncrementChatCallCount(tokenID uint) error {
	now := time.Now()
	return database.DB.Model(&models.ChatToken{}).Where("id = ?", tokenID).Updates(map[string]interface{}{
		"total_call_count": gorm.Expr("total_call_count + 1"),
		"daily_call_count": gorm.Expr("daily_call_count + 1"),
		"last_used_at":     &now,
	}).Error
}

// SelectImageToken 随机选择一个可用的 Image Token
func (s *TokenSelector) SelectImageToken() (*models.ImageToken, error) {
	var tokens []models.ImageToken

	if err := database.DB.Where("enabled = ?", true).Find(&tokens).Error; err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	selected := tokens[r.Intn(len(tokens))]

	return &selected, nil
}

// IncrementImageCallCount 增加 Image Token 的调用计数
func (s *TokenSelector) IncrementImageCallCount(tokenID uint) error {
	now := time.Now()
	return database.DB.Model(&models.ImageToken{}).Where("id = ?", tokenID).Updates(map[string]interface{}{
		"total_call_count": gorm.Expr("total_call_count + 1"),
		"daily_call_count": gorm.Expr("daily_call_count + 1"),
		"last_used_at":     &now,
	}).Error
}

// ResetDailyCallCount 重置所有 Token 的当日调用计数（定时任务调用）
func ResetDailyCallCount() error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.OCRToken{}).Where("1 = 1").Update("daily_call_count", 0).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.AudioToken{}).Where("1 = 1").Update("daily_call_count", 0).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.ChatToken{}).Where("1 = 1").Update("daily_call_count", 0).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.ImageToken{}).Where("1 = 1").Update("daily_call_count", 0).Error; err != nil {
			return err
		}
		return nil
	})
}
