package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"zai2api-go/database"
	"zai2api-go/models"

	"github.com/gin-gonic/gin"
)

const maxBatchSize = 500

func validateBatchSize(tokens []string, c *gin.Context) bool {
	if len(tokens) > maxBatchSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("单次最多导入 %d 个 Token，当前 %d 个，请分批导入", maxBatchSize, len(tokens))})
		return false
	}
	return true
}

// TokenResponse 通用响应结构
type TokenResponse struct {
	ID             uint       `json:"id"`
	Token          string     `json:"token"`
	ImportedAt     time.Time  `json:"imported_at"`
	LastUsedAt     *time.Time `json:"last_used_at"`
	Enabled        bool       `json:"enabled"`
	TotalCallCount int        `json:"total_call_count"`
	DailyCallCount int        `json:"daily_call_count"`
}

// TokenCreateRequest 创建 Token 请求
type TokenCreateRequest struct {
	Tokens []string `json:"tokens" binding:"required"`
}

// TokenBatchRequest 批量操作请求
type TokenBatchRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// GetAudioTokens 获取所有 Audio Token
func GetAudioTokens(c *gin.Context) {
	var tokens []models.AudioToken
	database.DB.Order("id desc").Find(&tokens)
	c.JSON(http.StatusOK, tokens)
}

// GetOCRTokens 获取所有 OCR Token
func GetOCRTokens(c *gin.Context) {
	var tokens []models.OCRToken
	database.DB.Order("id desc").Find(&tokens)
	c.JSON(http.StatusOK, tokens)
}

// GetChatTokens 获取所有 Chat Token
func GetChatTokens(c *gin.Context) {
	var tokens []models.ChatToken
	database.DB.Order("id desc").Find(&tokens)
	c.JSON(http.StatusOK, tokens)
}

// CreateAudioTokens 批量创建 Audio Token
func CreateAudioTokens(c *gin.Context) {
	var req TokenCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validateBatchSize(req.Tokens, c) {
		return
	}

	now := time.Now()
	var created []models.AudioToken
	var duplicates []string

	for _, tokenStr := range req.Tokens {
		tokenStr = strings.TrimSpace(tokenStr)
		if tokenStr == "" {
			continue
		}

		var existing models.AudioToken
		if err := database.DB.Where("token = ?", tokenStr).First(&existing).Error; err == nil {
			duplicates = append(duplicates, tokenStr)
			continue
		}

		token := models.AudioToken{
			Token:      tokenStr,
			ImportedAt: now,
			Enabled:    true,
		}
		if err := database.DB.Create(&token).Error; err != nil {
			continue
		}
		created = append(created, token)
	}

	c.JSON(http.StatusOK, gin.H{
		"created":    len(created),
		"duplicates": len(duplicates),
		"data":       created,
	})
}

// CreateOCRTokens 批量创建 OCR Token
func CreateOCRTokens(c *gin.Context) {
	var req TokenCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validateBatchSize(req.Tokens, c) {
		return
	}

	now := time.Now()
	var created []models.OCRToken
	var duplicates []string

	for _, tokenStr := range req.Tokens {
		tokenStr = strings.TrimSpace(tokenStr)
		if tokenStr == "" {
			continue
		}

		var existing models.OCRToken
		if err := database.DB.Where("token = ?", tokenStr).First(&existing).Error; err == nil {
			duplicates = append(duplicates, tokenStr)
			continue
		}

		token := models.OCRToken{
			Token:      tokenStr,
			ImportedAt: now,
			Enabled:    true,
		}
		if err := database.DB.Create(&token).Error; err != nil {
			continue
		}
		created = append(created, token)
	}

	c.JSON(http.StatusOK, gin.H{
		"created":    len(created),
		"duplicates": len(duplicates),
		"data":       created,
	})
}

// CreateChatTokens 批量创建 Chat Token
func CreateChatTokens(c *gin.Context) {
	var req TokenCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validateBatchSize(req.Tokens, c) {
		return
	}

	now := time.Now()
	var created []models.ChatToken
	var duplicates []string

	for _, tokenStr := range req.Tokens {
		tokenStr = strings.TrimSpace(tokenStr)
		if tokenStr == "" {
			continue
		}

		var existing models.ChatToken
		if err := database.DB.Where("token = ?", tokenStr).First(&existing).Error; err == nil {
			duplicates = append(duplicates, tokenStr)
			continue
		}

		token := models.ChatToken{
			Token:      tokenStr,
			ImportedAt: now,
			Enabled:    true,
		}
		if err := database.DB.Create(&token).Error; err != nil {
			continue
		}
		created = append(created, token)
	}

	c.JSON(http.StatusOK, gin.H{
		"created":    len(created),
		"duplicates": len(duplicates),
		"data":       created,
	})
}

// DeleteAudioToken 删除单个 Audio Token
func DeleteAudioToken(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := database.DB.Delete(&models.AudioToken{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// DeleteOCRToken 删除单个 OCR Token
func DeleteOCRToken(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := database.DB.Delete(&models.OCRToken{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// DeleteChatToken 删除单个 Chat Token
func DeleteChatToken(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := database.DB.Delete(&models.ChatToken{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// BatchDeleteAudioTokens 批量删除 Audio Token
func BatchDeleteAudioTokens(c *gin.Context) {
	var req TokenBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Delete(&models.AudioToken{}, req.IDs)
	c.JSON(http.StatusOK, gin.H{"deleted": len(req.IDs)})
}

// BatchDeleteOCRTokens 批量删除 OCR Token
func BatchDeleteOCRTokens(c *gin.Context) {
	var req TokenBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Delete(&models.OCRToken{}, req.IDs)
	c.JSON(http.StatusOK, gin.H{"deleted": len(req.IDs)})
}

// BatchDeleteChatTokens 批量删除 Chat Token
func BatchDeleteChatTokens(c *gin.Context) {
	var req TokenBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Delete(&models.ChatToken{}, req.IDs)
	c.JSON(http.StatusOK, gin.H{"deleted": len(req.IDs)})
}

// ToggleAudioToken 切换 Audio Token 启用状态
func ToggleAudioToken(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var token models.AudioToken
	if err := database.DB.First(&token, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	token.Enabled = !token.Enabled
	database.DB.Save(&token)
	c.JSON(http.StatusOK, token)
}

// ToggleOCRToken 切换 OCR Token 启用状态
func ToggleOCRToken(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var token models.OCRToken
	if err := database.DB.First(&token, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	token.Enabled = !token.Enabled
	database.DB.Save(&token)
	c.JSON(http.StatusOK, token)
}

// ToggleChatToken 切换 Chat Token 启用状态
func ToggleChatToken(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var token models.ChatToken
	if err := database.DB.First(&token, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	token.Enabled = !token.Enabled
	database.DB.Save(&token)
	c.JSON(http.StatusOK, token)
}

// BatchToggleAudioTokens 批量切换 Audio Token 启用状态
func BatchToggleAudioTokens(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Enable bool   `json:"enable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&models.AudioToken{}).Where("id IN ?", req.IDs).Update("enabled", req.Enable)
	c.JSON(http.StatusOK, gin.H{"updated": len(req.IDs)})
}

// BatchToggleOCRTokens 批量切换 OCR Token 启用状态
func BatchToggleOCRTokens(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Enable bool   `json:"enable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&models.OCRToken{}).Where("id IN ?", req.IDs).Update("enabled", req.Enable)
	c.JSON(http.StatusOK, gin.H{"updated": len(req.IDs)})
}

// BatchToggleChatTokens 批量切换 Chat Token 启用状态
func BatchToggleChatTokens(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Enable bool   `json:"enable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&models.ChatToken{}).Where("id IN ?", req.IDs).Update("enabled", req.Enable)
	c.JSON(http.StatusOK, gin.H{"updated": len(req.IDs)})
}

// GetImageTokens 获取所有 Image Token
func GetImageTokens(c *gin.Context) {
	var tokens []models.ImageToken
	database.DB.Order("id desc").Find(&tokens)
	c.JSON(http.StatusOK, tokens)
}

// CreateImageTokens 批量创建 Image Token
func CreateImageTokens(c *gin.Context) {
	var req TokenCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validateBatchSize(req.Tokens, c) {
		return
	}

	now := time.Now()
	var created []models.ImageToken
	var duplicates []string

	for _, tokenStr := range req.Tokens {
		tokenStr = strings.TrimSpace(tokenStr)
		if tokenStr == "" {
			continue
		}

		var existing models.ImageToken
		if err := database.DB.Where("token = ?", tokenStr).First(&existing).Error; err == nil {
			duplicates = append(duplicates, tokenStr)
			continue
		}

		token := models.ImageToken{
			Token:      tokenStr,
			ImportedAt: now,
			Enabled:    true,
		}
		if err := database.DB.Create(&token).Error; err != nil {
			continue
		}
		created = append(created, token)
	}

	c.JSON(http.StatusOK, gin.H{
		"created":    len(created),
		"duplicates": len(duplicates),
		"data":       created,
	})
}

// DeleteImageToken 删除单个 Image Token
func DeleteImageToken(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := database.DB.Delete(&models.ImageToken{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// BatchDeleteImageTokens 批量删除 Image Token
func BatchDeleteImageTokens(c *gin.Context) {
	var req TokenBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Delete(&models.ImageToken{}, req.IDs)
	c.JSON(http.StatusOK, gin.H{"deleted": len(req.IDs)})
}

// ToggleImageToken 切换 Image Token 启用状态
func ToggleImageToken(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var token models.ImageToken
	if err := database.DB.First(&token, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	token.Enabled = !token.Enabled
	database.DB.Save(&token)
	c.JSON(http.StatusOK, token)
}

// BatchToggleImageTokens 批量切换 Image Token 启用状态
func BatchToggleImageTokens(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Enable bool   `json:"enable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&models.ImageToken{}).Where("id IN ?", req.IDs).Update("enabled", req.Enable)
	c.JSON(http.StatusOK, gin.H{"updated": len(req.IDs)})
}
