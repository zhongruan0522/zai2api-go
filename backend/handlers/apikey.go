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
	"gorm.io/gorm"
)

// APIKeyCreateRequest 创建 API Key 请求
type APIKeyCreateRequest struct {
	Services string `json:"services"` // 服务类型：ocr,chat,image 或 * 表示全部
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

func GetOCRLogs(c *gin.Context) {
	var logs []models.OCRLog
	paginateLogs(c, database.DB.Model(&models.OCRLog{}), &logs)
}

func GetChatLogs(c *gin.Context) {
	var logs []models.ChatLog
	paginateLogs(c, database.DB.Model(&models.ChatLog{}), &logs)
}

func GetImageLogs(c *gin.Context) {
	var logs []models.ImageLog
	paginateLogs(c, database.DB.Model(&models.ImageLog{}), &logs)
}

func paginateLogs(c *gin.Context, query *gorm.DB, result interface{}) {
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

	var total int64
	query.Count(&total)
	query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(result)

	c.JSON(http.StatusOK, gin.H{
		"data":      result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type ChannelStats struct {
	Total   int64 `json:"total"`
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
	Today   int64 `json:"today"`
}

func GetOCRLogStats(c *gin.Context) {
	channelStats(c, &models.OCRLog{})
}

func GetChatLogStats(c *gin.Context) {
	channelStats(c, &models.ChatLog{})
}

func GetImageLogStats(c *gin.Context) {
	channelStats(c, &models.ImageLog{})
}

func channelStats(c *gin.Context, sample interface{}) {
	var stats ChannelStats

	database.DB.Model(sample).Count(&stats.Total)
	database.DB.Model(sample).Where("success = ?", true).Count(&stats.Success)
	database.DB.Model(sample).Where("success = ?", false).Count(&stats.Failed)

	today := time.Now().Format("2006-01-02")
	database.DB.Model(sample).Where("DATE(created_at) = ?", today).Count(&stats.Today)

	c.JSON(http.StatusOK, stats)
}
