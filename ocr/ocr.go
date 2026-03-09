package ocr

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	ZAIOCRBaseURL = "https://ocr.z.ai"
	OCRAPI        = "https://ocr.z.ai/api/v1/z-ocr/tasks/process"
)

// 全局正则，避免每次请求重复编译
var bearerTokenRe = regexp.MustCompile(`(?i)^Bearer\s+(.+)$`)

// 智谱OCR响应结构
type ZAIOCRResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      struct {
		TaskID          string             `json:"task_id"`
		Status          string             `json:"status"`
		FileName        string             `json:"file_name"`
		FileSize        int64              `json:"file_size"`
		FileType        string             `json:"file_type"`
		FileURL         string             `json:"file_url"`
		CreatedAt       string             `json:"created_at"`
		MarkdownContent string             `json:"markdown_content"`
		JsonContent     string             `json:"json_content"`
		Layout          []ZAIOCRLayoutItem `json:"layout"`
		DataInfo        ZAIOCRDataInfo     `json:"data_info"`
	} `json:"data"`
	Timestamp int64 `json:"timestamp"`
}

type ZAIOCRLayoutItem struct {
	BlockContent string  `json:"block_content"`
	BBox         []int   `json:"bbox"`
	BlockID      int     `json:"block_id"`
	PageIndex    int     `json:"page_index"`
	BlockLabel   string  `json:"block_label"`
	Score        float64 `json:"score"`
}

type ZAIOCRDataInfo struct {
	Pages    []ZAIOCRPageInfo `json:"pages"`
	NumPages int              `json:"num_pages"`
}

type ZAIOCRPageInfo struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// 对外接口响应结构
type OCRApiResponse struct {
	TaskID         string            `json:"task_id"`
	Message        string            `json:"message"`
	Status         string            `json:"status"`
	WordsResultNum int               `json:"words_result_num"`
	WordsResult    []WordsResultItem `json:"words_result"`
}

type WordsResultItem struct {
	Location    LocationInfo    `json:"location"`
	Words       string          `json:"words"`
	Probability *ProbabilityInfo `json:"probability,omitempty"`
}

type LocationInfo struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ProbabilityInfo struct {
	Average  float64 `json:"average"`
	Variance float64 `json:"variance"`
	Min      float64 `json:"min"`
}

// 错误响应结构（与文档保持一致）
type OCRErrorResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// 模型列表数据
var modelList = []gin.H{
	{"id": "zhipu-ocr", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
}

// 状态映射：智谱状态 -> 对外状态
func mapStatus(s string) string {
	switch s {
	case "completed", "succeeded", "success":
		return "succeeded"
	case "failed", "error":
		return "failed"
	case "processing", "pending":
		return "processing"
	default:
		return "failed"
	}
}

// 获取模型列表
func handleListModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   modelList,
	})
}

// 获取单个模型
func handleGetModel(c *gin.Context) {
	modelID := c.Param("model")
	for _, m := range modelList {
		if m["id"] == modelID {
			c.JSON(http.StatusOK, m)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
}

// 从Authorization header提取token
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	matches := bearerTokenRe.FindStringSubmatch(authHeader)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid Authorization header format")
	}
	return matches[1], nil
}

// 调用智谱OCR API
func callZAIOCRApi(token string, file *multipart.FileHeader) (*ZAIOCRResponse, error) {
	// 打开上传的文件
	fileContent, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer fileContent.Close()

	// 创建管道，实现流式传输
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// 在goroutine中写入文件
	go func() {
		defer pw.Close()
		defer writer.Close()

		part, err := writer.CreateFormFile("file", file.Filename)
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		if _, err := io.Copy(part, fileContent); err != nil {
			pw.CloseWithError(err)
			return
		}
	}()

	// 创建HTTP请求
	req, err := http.NewRequest("POST", OCRAPI, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", ZAIOCRBaseURL)
	req.Header.Set("Referer", ZAIOCRBaseURL+"/")
	req.Header.Set("User-Agent", "Mozilla/5.0 AppleWebKit/537.36 Chrome/143 Safari/537")
	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var zaiResp ZAIOCRResponse
	if err := json.Unmarshal(body, &zaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &zaiResp, nil
}

// 将智谱OCR layout转换为对外words_result格式
func convertLayoutToWordsResult(layout []ZAIOCRLayoutItem) []WordsResultItem {
	var wordsResult []WordsResultItem

	for _, item := range layout {
		// 过滤非文本块（只保留 text 和 title）
		if item.BlockLabel != "text" && item.BlockLabel != "title" {
			continue
		}

		// 过滤空内容
		if item.BlockContent == "" {
			continue
		}

		// bbox合法性检查
		if len(item.BBox) < 4 {
			continue
		}

		x1, y1, x2, y2 := item.BBox[0], item.BBox[1], item.BBox[2], item.BBox[3]

		// 坐标合法性检查
		if x2 < x1 || y2 < y1 {
			continue
		}

		wordsResult = append(wordsResult, WordsResultItem{
			Location: LocationInfo{
				Left:   x1,
				Top:    y1,
				Width:  x2 - x1,
				Height: y2 - y1,
			},
			Words: item.BlockContent,
			Probability: &ProbabilityInfo{
				Average:  item.Score,
				Variance: 0,
				Min:      item.Score,
			},
		})
	}

	return wordsResult
}

// 处理OCR文件上传
func handleFilesOCR(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, OCRErrorResponse{
			Code:      40001,
			Message:   "缺少文件参数",
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 提取token
	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, OCRErrorResponse{
			Code:      40101,
			Message:   err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 调用智谱OCR API
	zaiResp, err := callZAIOCRApi(token, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, OCRErrorResponse{
			Code:      50001,
			Message:   err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 检查智谱API响应
	if zaiResp.Code != 200 {
		c.JSON(http.StatusBadRequest, OCRErrorResponse{
			Code:      zaiResp.Code,
			Message:   zaiResp.Message,
			Timestamp: zaiResp.Timestamp,
		})
		return
	}

	// 转换layout为words_result格式
	wordsResult := convertLayoutToWordsResult(zaiResp.Data.Layout)

	// 构建响应（透传上游状态）
	response := OCRApiResponse{
		TaskID:         zaiResp.Data.TaskID,
		Message:        zaiResp.Message,
		Status:         mapStatus(zaiResp.Data.Status),
		WordsResultNum: len(wordsResult),
		WordsResult:    wordsResult,
	}

	c.JSON(http.StatusOK, response)
}

// RegisterRoutes 注册OCR模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/models", handleListModels)
	r.GET("/models/:model", handleGetModel)
	r.POST("/files/ocr", handleFilesOCR)
}
