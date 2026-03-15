package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"zai2api-go/database"
	"zai2api-go/image"
	"zai2api-go/models"
	"zai2api-go/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImageHandler struct {
	tokenSelector *services.TokenSelector
}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{
		tokenSelector: services.NewTokenSelector(),
	}
}

func (h *ImageHandler) GenerateImage(c *gin.Context) {
	requestID := uuid.New().String()
	sourceIP := c.ClientIP()

	apiKeyID, valid := h.validateAPIKey(c)
	if !valid {
		h.logRequest(requestID, sourceIP, apiKeyID, 0, false, "401", "invalid api key")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	var req image.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, 0, false, "400", "invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Prompt == "" {
		h.logRequest(requestID, sourceIP, apiKeyID, 0, false, "400", "prompt is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	token, err := h.tokenSelector.SelectImageToken()
	if err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, 0, false, "503", "no available token")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available token"})
		return
	}

	respBody, err := image.SendRequest(&req, token.Token)
	if err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, "502", err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	var upstreamResp image.UpstreamResponse
	if err = json.Unmarshal(respBody, &upstreamResp); err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, "500", "parse response failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse response failed"})
		return
	}

	if upstreamResp.Code != 200 {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, fmt.Sprintf("%d", upstreamResp.Code), upstreamResp.Message)
		c.JSON(http.StatusOK, gin.H{
			"error": gin.H{
				"message": upstreamResp.Message,
				"type":    "upstream_error",
				"code":    upstreamResp.Code,
			},
		})
		return
	}

	_ = h.tokenSelector.IncrementImageCallCount(token.ID)

	result := image.ConvertResponse(&upstreamResp)

	h.logRequest(requestID, sourceIP, apiKeyID, token.ID, true, "", "")

	c.JSON(http.StatusOK, result)
}

func (h *ImageHandler) ChatGenerateImage(c *gin.Context) {
	requestID := uuid.New().String()
	sourceIP := c.ClientIP()

	apiKeyID, valid := h.validateAPIKey(c)
	if !valid {
		h.logRequest(requestID, sourceIP, apiKeyID, 0, false, "401", "invalid api key")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	var chatReq image.ChatRequest
	if err := c.ShouldBindJSON(&chatReq); err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, 0, false, "400", "invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	prompt := extractPromptFromMessages(chatReq.Messages)
	if prompt == "" {
		h.logRequest(requestID, sourceIP, apiKeyID, 0, false, "400", "prompt is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	resolution, ratio := image.ParseModelToParams(chatReq.Model)

	token, err := h.tokenSelector.SelectImageToken()
	if err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, 0, false, "503", "no available token")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available token"})
		return
	}

	genReq := &image.GenerateRequest{
		Prompt:           prompt,
		Ratio:            ratio,
		Resolution:       resolution,
		RmLabelWatermark: true,
	}

	respBody, err := image.SendRequest(genReq, token.Token)
	if err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, "502", err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	var upstreamResp image.UpstreamResponse
	if err = json.Unmarshal(respBody, &upstreamResp); err != nil {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, "500", "parse response failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse response failed"})
		return
	}

	if upstreamResp.Code != 200 {
		h.logRequest(requestID, sourceIP, apiKeyID, token.ID, false, fmt.Sprintf("%d", upstreamResp.Code), upstreamResp.Message)
		c.JSON(http.StatusOK, gin.H{
			"error": gin.H{
				"message": upstreamResp.Message,
				"type":    "upstream_error",
				"code":    upstreamResp.Code,
			},
		})
		return
	}

	_ = h.tokenSelector.IncrementImageCallCount(token.ID)

	result := image.ConvertToChatResponse(&upstreamResp, chatReq.Model)

	h.logRequest(requestID, sourceIP, apiKeyID, token.ID, true, "", "")

	c.JSON(http.StatusOK, result)
}

func extractPromptFromMessages(messages []image.ChatMessage) string {
	if len(messages) == 0 {
		return ""
	}

	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" && strings.TrimSpace(messages[i].Content) != "" {
			return messages[i].Content
		}
	}

	return strings.TrimSpace(messages[len(messages)-1].Content)
}

func (h *ImageHandler) validateAPIKey(c *gin.Context) (uint, bool) {
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

	if !hasService(apiKey.Services, "image") {
		return 0, false
	}

	return apiKey.ID, true
}

func (h *ImageHandler) logRequest(requestID, sourceIP string, apiKeyID, tokenID uint, success bool, errorCode, errorMsg string) {
	errorCode = truncateString(errorCode, 20)
	errorMsg = truncateString(errorMsg, 500)
	log := models.ImageLog{
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
