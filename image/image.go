package image

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	ZAIImageBaseURL = "https://image.z.ai"
	GenerateAPI     = "https://image.z.ai/api/proxy/images/generate"
)

// OpenAI 请求结构
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// 智谱请求结构
type ZAIRequest struct {
	Prompt           string `json:"prompt"`
	Ratio            string `json:"ratio"`
	Resolution       string `json:"resolution"`
	RmLabelWatermark bool   `json:"rm_label_watermark"`
}

// 智谱响应结构
type ZAIResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      struct {
		Image struct {
			ImageID     string `json:"image_id"`
			Prompt      string `json:"prompt"`
			Size        string `json:"size"`
			Ratio       string `json:"ratio"`
			Resolution  string `json:"resolution"`
			ImageURL    string `json:"image_url"`
			Status      string `json:"status"`
		} `json:"image"`
	} `json:"data"`
	Timestamp int64 `json:"timestamp"`
}

// OpenAI 流式响应结构
type OpenAIStreamChoice struct {
	Index        int                    `json:"index"`
	Delta        map[string]interface{} `json:"delta"`
	FinishReason *string                `json:"finish_reason"`
}

type OpenAIStreamResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIStreamChoice `json:"choices"`
}

// 解析模型名称获取分辨率和比例
func parseModel(model string) (resolution, ratio string) {
	resolution = "1K"
	ratio = "21:9"

	// 匹配分辨率
	if strings.Contains(model, "2k") || strings.Contains(model, "2K") {
		resolution = "2K"
	}

	// 比例后缀映射
	ratioMap := map[string]string{
		"-1-1":  "1:1",
		"-3-4":  "3:4",
		"-4-3":  "4:3",
		"-16-9": "16:9",
		"-9-16": "9:16",
		"-21-9": "21:9",
		"-9-21": "9:21",
	}

	modelLower := strings.ToLower(model)
	for suffix, r := range ratioMap {
		if strings.HasSuffix(modelLower, suffix) {
			ratio = r
			break
		}
	}

	return
}

// 从消息中提取用户提示词
func extractPrompt(messages []OpenAIMessage) string {
	for _, msg := range messages {
		if msg.Role == "user" {
			return msg.Content
		}
	}
	return ""
}

// 下载图片并转为Base64
func imageToBase64(url string) (string, string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// 检测 MIME 类型
	mimeType := http.DetectContentType(data)

	base64Data := base64.StdEncoding.EncodeToString(data)
	return base64Data, mimeType, nil
}

// 调用智谱API生成图片
func generateImage(token, prompt, ratio, resolution string) (*ZAIResponse, error) {
	reqBody := ZAIRequest{
		Prompt:           prompt,
		Ratio:            ratio,
		Resolution:       resolution,
		RmLabelWatermark: true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", GenerateAPI, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Origin", ZAIImageBaseURL)
	req.Header.Set("Referer", ZAIImageBaseURL+"/")
	req.Header.Set("User-Agent", "Mozilla/5.0 AppleWebKit/537.36 Chrome/143 Safari/537")
	req.Header.Set("Cookie", "session="+token)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var zaiResp ZAIResponse
	if err := json.Unmarshal(body, &zaiResp); err != nil {
		return nil, err
	}

	return &zaiResp, nil
}

func handleChatCompletions(c *gin.Context) {
	var req OpenAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从 Authorization header 获取 token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		return
	}

	// 提取 Bearer token
	re := regexp.MustCompile(`(?i)^Bearer\s+(.+)$`)
	matches := re.FindStringSubmatch(authHeader)
	if len(matches) < 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		return
	}
	token := matches[1]

	// 解析模型参数
	resolution, ratio := parseModel(req.Model)

	// 提取提示词
	prompt := extractPrompt(req.Messages)
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No user message found"})
		return
	}

	// 调用智谱API
	zaiResp, err := generateImage(token, prompt, ratio, resolution)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if zaiResp.Code != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("ZAI API error: %s (code: %d)", zaiResp.Message, zaiResp.Code),
		})
		return
	}

	imageURL := zaiResp.Data.Image.ImageURL
	if imageURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No image URL in response"})
		return
	}

	// 下载图片并转为Base64
	base64Data, mimeType, err := imageToBase64(imageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to download image: %v", err)})
		return
	}

	// 构建Markdown格式图片
	imageMarkdown := fmt.Sprintf("![image](data:%s;base64,%s)", mimeType, base64Data)

	// 构建OpenAI标准流式响应
	finishReason := "stop"
	resp1 := OpenAIStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []OpenAIStreamChoice{
			{
				Index: 0,
				Delta: map[string]interface{}{
					"role":    "assistant",
					"content": imageMarkdown,
				},
				FinishReason: nil,
			},
		},
	}

	resp2 := OpenAIStreamResponse{
		ID:      resp1.ID,
		Object:  "chat.completion.chunk",
		Created: resp1.Created,
		Model:   req.Model,
		Choices: []OpenAIStreamChoice{
			{
				Index:        0,
				Delta:        map[string]interface{}{},
				FinishReason: &finishReason,
			},
		},
	}

	// 发送SSE响应
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	jsonData1, _ := json.Marshal(resp1)
	c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(jsonData1)))

	jsonData2, _ := json.Marshal(resp2)
	c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(jsonData2)))

	c.Writer.WriteString("data: [DONE]\n\n")
	c.Writer.Flush()
}

// 模型列表数据
var modelList = []gin.H{
	{"id": "gemini-3-pro-image-1k", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-2k", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-1k-1-1", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-1k-3-4", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-1k-4-3", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-1k-16-9", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-1k-9-16", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-1k-21-9", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-1k-9-21", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-2k-1-1", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-2k-3-4", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-2k-4-3", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-2k-16-9", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-2k-9-16", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-2k-21-9", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	{"id": "gemini-3-pro-image-2k-9-21", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
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

// RegisterRoutes 注册绘图模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/models", handleListModels)
	r.GET("/models/:model", handleGetModel)
	r.POST("/chat/completions", handleChatCompletions)
}
