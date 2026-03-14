package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zai2api-go/database"
	"zai2api-go/models"

	"github.com/gin-gonic/gin"
)

// APIKeyCreateRequest 创建 API Key 请求
type APIKeyCreateRequest struct {
	Services string `json:"services"` // 服务类型：ocr,audio,chat,image 或 * 表示全部
}

// APIKeyBatchRequest 批量操作请求
type APIKeyBatchRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// GetAPIKeys 获取所有 API Key
func GetAPIKeys(c *gin.Context) {
	var keys []models.APIKey
	database.DB.Order("id desc").Find(&keys)
	c.JSON(http.StatusOK, keys)
}

// CreateAPIKey 创建 API Key
func CreateAPIKey(c *gin.Context) {
	var req APIKeyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成随机 Key
	keyBytes := make([]byte, 24)
	if _, err := rand.Read(keyBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate key failed"})
		return
	}
	key := "sk-" + hex.EncodeToString(keyBytes)

	// 处理服务类型
	services := strings.TrimSpace(req.Services)
	if services == "" {
		services = "*" // 默认全部服务
	}

	apiKey := models.APIKey{
		Key:      key,
		Services: services,
		Enabled:  true,
	}

	if err := database.DB.Create(&apiKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// DeleteAPIKey 删除 API Key
func DeleteAPIKey(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := database.DB.Delete(&models.APIKey{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ToggleAPIKey 切换 API Key 启用状态
func ToggleAPIKey(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var key models.APIKey
	if err := database.DB.First(&key, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	key.Enabled = !key.Enabled
	database.DB.Save(&key)
	c.JSON(http.StatusOK, key)
}

// BatchDeleteAPIKeys 批量删除 API Key
func BatchDeleteAPIKeys(c *gin.Context) {
	var req APIKeyBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Delete(&models.APIKey{}, req.IDs)
	c.JSON(http.StatusOK, gin.H{"deleted": len(req.IDs)})
}

// BatchToggleAPIKeys 批量切换 API Key 启用状态
func BatchToggleAPIKeys(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Enable bool   `json:"enable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&models.APIKey{}).Where("id IN ?", req.IDs).Update("enabled", req.Enable)
	c.JSON(http.StatusOK, gin.H{"updated": len(req.IDs)})
}

// GetRequestLogs 获取请求日志
func GetRequestLogs(c *gin.Context) {
	var logs []models.RequestLog

	page := 1
	pageSize := 50
	const maxPageSize = 200
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	channel := c.Query("channel")

	query := database.DB.Model(&models.RequestLog{})
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}

	var total int64
	query.Count(&total)

	query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetRequestLogStats 获取请求统计
func GetRequestLogStats(c *gin.Context) {
	type Stats struct {
		Total   int64 `json:"total"`
		Success int64 `json:"success"`
		Failed  int64 `json:"failed"`
		Today   int64 `json:"today"`
		OCR     int64 `json:"ocr"`
		Audio   int64 `json:"audio"`
		Chat    int64 `json:"chat"`
		Image   int64 `json:"image"`
	}

	var stats Stats

	// 总数
	database.DB.Model(&models.RequestLog{}).Count(&stats.Total)

	// 成功数
	database.DB.Model(&models.RequestLog{}).Where("success = ?", true).Count(&stats.Success)

	// 失败数
	database.DB.Model(&models.RequestLog{}).Where("success = ?", false).Count(&stats.Failed)

	// 今日请求数
	today := time.Now().Format("2006-01-02")
	database.DB.Model(&models.RequestLog{}).Where("DATE(created_at) = ?", today).Count(&stats.Today)

	// 按渠道统计
	database.DB.Model(&models.RequestLog{}).Where("channel = ?", "ocr").Count(&stats.OCR)
	database.DB.Model(&models.RequestLog{}).Where("channel = ?", "audio").Count(&stats.Audio)
	database.DB.Model(&models.RequestLog{}).Where("channel = ?", "chat").Count(&stats.Chat)
	database.DB.Model(&models.RequestLog{}).Where("channel = ?", "image").Count(&stats.Image)

	c.JSON(http.StatusOK, stats)
}
