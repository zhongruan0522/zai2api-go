package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
	"zai2api-go/config"
	"zai2api-go/database"
	"zai2api-go/models"
	"zai2api-go/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ocrUpstreamURL = "https://ocr.z.ai/api/v1/z-ocr/tasks/process"

// OCRHandler OCR 处理器
type OCRHandler struct {
	tokenSelector *services.TokenSelector
	dailyLimit    int // OCR Token 每日限额
}

// NewOCRHandler 创建 OCR 处理器
func NewOCRHandler(cfg *config.Config) *OCRHandler {
	return &OCRHandler{
		tokenSelector: services.NewTokenSelector(),
		dailyLimit:    cfg.OCRDailyLimit,
	}
}

// ProcessOCR 处理 OCR 请求
func (h *OCRHandler) ProcessOCR(c *gin.Context) {
	requestID := uuid.New().String()

	// 获取源 IP
	sourceIP := c.ClientIP()

	// 验证 API Key
	apiKeyID, valid := h.validateAPIKey(c)
	if !valid {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, 0, false, "401", "invalid api key")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	// 选择一个可用的 Token
	token, err := h.tokenSelector.SelectOCRToken()
	if err != nil {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, 0, false, "503", "no available token")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available token"})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, token.ID, false, "400", "file required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	// 构建上游请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, token.ID, false, "500", "create form file failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create form file failed"})
		return
	}

	if _, err = io.Copy(part, file); err != nil {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, token.ID, false, "500", "copy file failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "copy file failed"})
		return
	}

	if err = writer.Close(); err != nil {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, token.ID, false, "500", "close writer failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "close writer failed"})
		return
	}

	// 发送请求到上游
	req, err := http.NewRequest("POST", ocrUpstreamURL, body)
	if err != nil {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, token.ID, false, "500", "create request failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create request failed"})
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://ocr.z.ai")
	req.Header.Set("Referer", "https://ocr.z.ai/")
	req.Header.Set("User-Agent", "Mozilla/5.0 AppleWebKit/537.36 Chrome/143 Safari/537.36")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, token.ID, false, "502", "upstream request failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream request failed"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, token.ID, false, "500", "read response failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read response failed"})
		return
	}

	// 解析上游响应
	var upstreamResp OCRUpstreamResponse
	if err = json.Unmarshal(respBody, &upstreamResp); err != nil {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, token.ID, false, "500", "parse response failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse response failed"})
		return
	}

	// 检查上游错误
	if upstreamResp.Code != 200 {
		h.logRequest(requestID, "ocr", sourceIP, apiKeyID, token.ID, false, fmt.Sprintf("%d", upstreamResp.Code), upstreamResp.Message)
		c.JSON(http.StatusOK, gin.H{
			"code":      upstreamResp.Code,
			"message":   upstreamResp.Message,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// 增加调用计数
	_ = h.tokenSelector.IncrementOCRCallCount(token.ID)

	// 转换响应格式
	result := h.convertResponse(&upstreamResp)

	// 记录成功日志
	h.logRequest(requestID, "ocr", sourceIP, apiKeyID, token.ID, true, "", "")

	c.JSON(http.StatusOK, result)
}

// validateAPIKey 验证 API Key
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

	// 检查服务类型
	if !h.hasService(apiKey.Services, "ocr") {
		return 0, false
	}

	return apiKey.ID, true
}

// hasService 检查是否包含指定服务
func (h *OCRHandler) hasService(services, service string) bool {
	if services == "" || services == "*" {
		return true
	}
	parts := strings.Split(services, ",")
	for _, p := range parts {
		if strings.TrimSpace(p) == service {
			return true
		}
	}
	return false
}

// logRequest 记录请求日志
func (h *OCRHandler) logRequest(requestID, channel, sourceIP string, apiKeyID, tokenID uint, success bool, errorCode, errorMsg string) {
	log := models.RequestLog{
		RequestID: requestID,
		CreatedAt: time.Now(),
		Channel:   channel,
		SourceIP:  sourceIP,
		APIKeyID:  apiKeyID,
		TokenID:   tokenID,
		Success:   success,
		ErrorCode: errorCode,
		ErrorMsg:  errorMsg,
	}
	database.DB.Create(&log)
}

// OCRUpstreamResponse 上游响应结构
type OCRUpstreamResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Data      struct {
		TaskID          string         `json:"task_id"`
		Status          string         `json:"status"`
		FileName        string         `json:"file_name"`
		FileSize        int64          `json:"file_size"`
		FileType        string         `json:"file_type"`
		FileURL         string         `json:"file_url"`
		CreatedAt       string         `json:"created_at"`
		MarkdownContent string         `json:"markdown_content"`
		JsonContent     string         `json:"json_content"`
		Layout          []LayoutBlock  `json:"layout"`
		DataInfo        *DataInfo      `json:"data_info"`
	} `json:"data"`
}

// LayoutBlock 布局块
type LayoutBlock struct {
	BlockContent string  `json:"block_content"`
	BBox         []int   `json:"bbox"`
	BlockID      int     `json:"block_id"`
	PageIndex    int     `json:"page_index"`
	BlockLabel   string  `json:"block_label"`
	Score        float64 `json:"score"`
}

// DataInfo 数据信息
type DataInfo struct {
	Pages    []PageSize `json:"pages"`
	NumPages int        `json:"num_pages"`
}

// PageSize 页面尺寸
type PageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// OCRAPIResponse 对外响应结构
type OCRAPIResponse struct {
	TaskID        string             `json:"task_id"`
	Message       string             `json:"message"`
	Status        string             `json:"status"`
	WordsResultNum int               `json:"words_result_num"`
	WordsResult   []WordsResultItem  `json:"words_result"`
}

// WordsResultItem 文字识别结果项
type WordsResultItem struct {
	Location    Location    `json:"location"`
	Words       string      `json:"words"`
	Probability Probability `json:"probability"`
}

// Location 位置信息
type Location struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Probability 置信度
type Probability struct {
	Average  float64 `json:"average"`
	Variance float64 `json:"variance"`
	Min      float64 `json:"min"`
}

// convertResponse 转换响应格式
func (h *OCRHandler) convertResponse(upstream *OCRUpstreamResponse) *OCRAPIResponse {
	wordsResult := make([]WordsResultItem, 0, len(upstream.Data.Layout))

	for _, block := range upstream.Data.Layout {
		// 跳过非文本块
		if block.BlockLabel != "text" {
			continue
		}

		// bbox 转换: [x1, y1, x2, y2] -> left, top, width, height
		left := block.BBox[0]
		top := block.BBox[1]
		width := block.BBox[2] - block.BBox[0]
		height := block.BBox[3] - block.BBox[1]

		item := WordsResultItem{
			Location: Location{
				Left:   left,
				Top:    top,
				Width:  width,
				Height: height,
			},
			Words: block.BlockContent,
			Probability: Probability{
				Average:  block.Score,
				Variance: 0,
				Min:      block.Score,
			},
		}
		wordsResult = append(wordsResult, item)
	}

	return &OCRAPIResponse{
		TaskID:         upstream.Data.TaskID,
		Message:        "成功",
		Status:         "succeeded",
		WordsResultNum: len(wordsResult),
		WordsResult:    wordsResult,
	}
}
