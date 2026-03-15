package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"zai2api-go/config"
	"zai2api-go/database"
	"zai2api-go/models"
	"zai2api-go/ocr"
	"zai2api-go/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OCRHandler struct {
	tokenSelector *services.TokenSelector
	dailyLimit    int
	maxRespBytes  int64
}

func NewOCRHandler(cfg *config.Config) *OCRHandler {
	return &OCRHandler{
		tokenSelector: services.NewTokenSelector(),
		dailyLimit:    cfg.OCRDailyLimit,
		maxRespBytes:  cfg.UpstreamMaxRespBytes,
	}
}

func (h *OCRHandler) ProcessOCR(c *gin.Context) {
	requestID := uuid.New().String()
	sourceIP := c.ClientIP()

	apiKeyID, valid := h.validateAPIKey(c)
	if !valid {
		h.logRequest(requestID, sourceIP, apiKeyID, 0, false, "401", "invalid api key")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	token, err := h.tokenSelector.SelectOCRToken()
	if err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, 0, false, "503", "no available token")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available token"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, "400", "file required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, "400", "open file failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "open file failed"})
		return
	}
	defer file.Close()

	respBody, err := ocr.SendRequest(file, fileHeader.Filename, token.Token, h.maxRespBytes)
	if err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, "502", err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	var upstreamResp ocr.UpstreamResponse
	if err = json.Unmarshal(respBody, &upstreamResp); err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, "500", "parse response failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse response failed"})
		return
	}

	if upstreamResp.Code != 200 {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, fmt.Sprintf("%d", upstreamResp.Code), upstreamResp.Message)
		c.JSON(http.StatusOK, gin.H{
			"code":      upstreamResp.Code,
			"message":   upstreamResp.Message,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	_ = h.tokenSelector.IncrementOCRCallCount(token.ID)

	result := ocr.ConvertResponse(&upstreamResp)

	h.logRequest(requestID, sourceIP, apiKeyID, token.ID, true, "", "")

	c.JSON(http.StatusOK, result)
}

func (h *OCRHandler) validateAPIKey(c *gin.Context) (uint, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, false
	}

	key := parts[1]
	var apiKey models.APIKey
	if err := database.DB.Where("key = ? AND enabled = ?", key, true).First(&apiKey).Error; err != nil {
		return 0, false
	}

	if !hasService(apiKey.Services, "ocr") {
		return 0, false
	}

	return apiKey.ID, true
}

func hasService(servicesStr, service string) bool {
	if servicesStr == "" || servicesStr == "*" {
		return true
	}
	parts := strings.Split(servicesStr, ",")
	for _, p := range parts {
		if strings.TrimSpace(p) == service {
			return true
		}
	}
	return false
}

func (h *OCRHandler) logRequest(requestID, sourceIP string, apiKeyID, tokenID uint, success bool, errorCode, errorMsg string) {
	errorCode = truncateString(errorCode, 20)
	errorMsg = truncateString(errorMsg, 500)
	log := models.OCRLog{
		BaseLog: models.BaseLog{
			RequestID: requestID,
			CreatedAt: time.Now(),
			SourceIP:  sourceIP,
			APIKeyID:  apiKeyID,
			TokenID:   tokenID,
			Success:   success,
			ErrorCode: errorCode,
			ErrorMsg:  errorMsg,
		},
	}
	database.DB.Create(&log)
}
