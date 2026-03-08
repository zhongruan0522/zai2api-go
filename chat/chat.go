package chat

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ZAIChatBaseURL = "https://chat.z.ai"
	HMACSecret     = "key-@@@@)))()((9))-xxxx&&&%%%%%"
	FEVersion      = "prod-fe-1.0.231"
	ClientVersion  = "0.0.1"
	UserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
)

// 智谱模型响应结构
type ZAIModelsResponse struct {
	Data []ZAIModel `json:"data"`
}

type ZAIModel struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	OwnedBy        string                 `json:"owned_by,omitempty"`
	Object         string                 `json:"object,omitempty"`
	Created        int64                  `json:"created,omitempty"`
	MaxContext     int                    `json:"max_context,omitempty"`
	Params         map[string]interface{} `json:"params,omitempty"`
}

// 从智谱API获取模型列表
func fetchZAIModels(token string) ([]gin.H, error) {
	req, err := http.NewRequest("GET", ZAIChatBaseURL+"/api/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Origin", ZAIChatBaseURL)
	req.Header.Set("Referer", ZAIChatBaseURL+"/")
	req.Header.Set("User-Agent", UserAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("fetch models failed: status=%d body=%s", resp.StatusCode, string(body)[:min(500, len(body))])
	}

	var zaiResp ZAIModelsResponse
	if err := json.Unmarshal(body, &zaiResp); err != nil {
		return nil, err
	}

	// 转换为OpenAI格式
	var modelList []gin.H
	for _, m := range zaiResp.Data {
		modelList = append(modelList, gin.H{
			"id":       m.ID,
			"object":   "model",
			"created":  m.Created,
			"owned_by": "zhipu",
		})
	}

	// 如果没有获取到模型，返回默认列表
	if len(modelList) == 0 {
		return getDefaultModels(), nil
	}

	return modelList, nil
}

// 默认模型列表（作为fallback）
func getDefaultModels() []gin.H {
	return []gin.H{
		{"id": "glm-5", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
		{"id": "glm-4.7v", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
		{"id": "glm-4-long", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
		{"id": "glm-4-flash", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
		{"id": "glm-4-plus", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
		{"id": "glm-4-air", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
		{"id": "glm-4-airx", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
		{"id": "glm-4v", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
		{"id": "glm-4v-plus", "object": "model", "created": 1700000000, "owned_by": "zhipu"},
	}
}

// OpenAI 请求结构
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Stream      bool            `json:"stream"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Tools       []Tool          `json:"tools,omitempty"`
}

type OpenAIMessage struct {
	Role         string          `json:"role"`
	Content      json.RawMessage `json:"content"` // 支持字符串或数组格式
	ToolCalls    []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID   string          `json:"tool_call_id,omitempty"`
}

type Tool struct {
	Type     string     `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ToolCall struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// 智谱 API 结构
type ZAIChatCreateRequest struct {
	Chat struct {
		ID              string                 `json:"id"`
		Title           string                 `json:"title"`
		Models          []string               `json:"models"`
		Params          map[string]interface{} `json:"params"`
		History         *ZAIChatHistory        `json:"history"`
		Tags            []string               `json:"tags"`
		Flags           []string               `json:"flags"`
		Features        []ZAIFeature           `json:"features"`
		MCPServers      []interface{}          `json:"mcp_servers"`
		EnableThinking  bool                   `json:"enable_thinking"`
		AutoWebSearch   bool                   `json:"auto_web_search"`
		MessageVersion  int                    `json:"message_version"`
		Extra           map[string]interface{} `json:"extra"`
		Timestamp       int64                  `json:"timestamp"`
	} `json:"chat"`
}

type ZAIChatHistory struct {
	Messages  map[string]ZAIMessage `json:"messages"`
	CurrentID string                `json:"currentId"`
}

type ZAIMessage struct {
	ID           string   `json:"id"`
	ParentID     *string  `json:"parentId"`
	ChildrenIDs  []string `json:"childrenIds"`
	Role         string   `json:"role"`
	Content      string   `json:"content"`
	Timestamp    int64    `json:"timestamp"`
	Models       []string `json:"models"`
}

type ZAIFeature struct {
	Type     string `json:"type"`
	Server   string `json:"server"`
	Status   string `json:"status"`
}

type ZAIChatCreateResponse struct {
	ID string `json:"id"`
}

type ZAICompletionRequest struct {
	Stream                       bool                   `json:"stream"`
	Model                        string                 `json:"model"`
	Messages                     []ZAIBlockMessage      `json:"messages"`
	SignaturePrompt              string                 `json:"signature_prompt"`
	Params                       map[string]interface{} `json:"params"`
	Extra                        map[string]interface{} `json:"extra"`
	Features                     ZAICompletionFeatures  `json:"features"`
	Variables                    map[string]string      `json:"variables"`
	ChatID                       string                 `json:"chat_id"`
	ID                           string                 `json:"id"`
	CurrentUserMessageID         string                 `json:"current_user_message_id"`
	CurrentUserMessageParentID   *string                `json:"current_user_message_parent_id"`
	BackgroundTasks              ZAIBackgroundTasks     `json:"background_tasks"`
	Tools                        []ZAIBlockTool         `json:"tools,omitempty"`
}

type ZAIBlockMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type ZAICompletionFeatures struct {
	ImageGeneration bool     `json:"image_generation"`
	WebSearch       bool     `json:"web_search"`
	AutoWebSearch   bool     `json:"auto_web_search"`
	PreviewMode     bool     `json:"preview_mode"`
	Flags           []string `json:"flags"`
	EnableThinking  bool     `json:"enable_thinking"`
}

type ZAIBackgroundTasks struct {
	TitleGeneration bool `json:"title_generation"`
	TagsGeneration  bool `json:"tags_generation"`
}

type ZAIBlockTool struct {
	Type string `json:"type"`
	Function struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Parameters  map[string]interface{} `json:"parameters"`
	} `json:"function"`
}

// 智谱 SSE 响应结构
type ZAISSEEvent struct {
	Data *ZAISSEData `json:"data"`
}

type ZAISSEData struct {
	Phase        string `json:"phase"`
	DeltaContent string `json:"delta_content"`
	Done         bool   `json:"done"`
}

// OpenAI 流式响应结构
type OpenAIStreamResponse struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []OpenAIStreamChoice     `json:"choices"`
}

type OpenAIStreamChoice struct {
	Index        int                    `json:"index"`
	Delta        map[string]interface{} `json:"delta"`
	FinishReason *string                `json:"finish_reason"`
}

// 从 Authorization header 提取 token
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	// 支持 Bearer token 格式
	re := regexp.MustCompile(`(?i)^Bearer\s+(.+)$`)
	matches := re.FindStringSubmatch(authHeader)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	return "", fmt.Errorf("invalid Authorization header format")
}

// 生成 HMAC-SHA256 签名
func generateSignature(sortedPayload, prompt, timestamp string) string {
	// Base64 encode prompt
	b64Prompt := base64.StdEncoding.EncodeToString([]byte(prompt))

	// Build message
	message := fmt.Sprintf("%s|%s|%s", sortedPayload, b64Prompt, timestamp)

	// Calculate time bucket (5 minutes)
	timeBucket := int64(0)
	if ts := parseIntFromString(timestamp); ts > 0 {
		timeBucket = ts / (5 * 60 * 1000)
	}

	// First HMAC: HMAC-SHA256(secret, time_bucket)
	mac1 := hmac.New(sha256.New, []byte(HMACSecret))
	mac1.Write([]byte(fmt.Sprintf("%d", timeBucket)))
	derivedKey := hex.EncodeToString(mac1.Sum(nil))

	// Second HMAC: HMAC-SHA256(derived_key, message)
	mac2 := hmac.New(sha256.New, []byte(derivedKey))
	mac2.Write([]byte(message))
	signature := hex.EncodeToString(mac2.Sum(nil))

	return signature
}

func parseIntFromString(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

// 构建查询参数和签名
func buildQueryAndSignature(token, userID, prompt, chatID string) (string, string) {
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	requestID := uuid.New().String()

	// Core params
	core := map[string]string{
		"timestamp":  timestamp,
		"requestId":  requestID,
		"user_id":    userID,
	}

	// Build sorted payload
	var keys []string
	for k := range core {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sortedParts []string
	for _, k := range keys {
		sortedParts = append(sortedParts, fmt.Sprintf("%s,%s", k, core[k]))
	}
	sortedPayload := strings.Join(sortedParts, ",")

	// Generate signature
	signature := generateSignature(sortedPayload, prompt, timestamp)

	// Build full query params
	now := time.Now().UTC()
	extra := map[string]string{
		"version":            ClientVersion,
		"platform":           "web",
		"token":              token,
		"user_agent":         UserAgent,
		"language":           "zh-CN",
		"languages":          "zh-CN",
		"timezone":           "Asia/Shanghai",
		"cookie_enabled":     "true",
		"screen_width":       "1920",
		"screen_height":      "1080",
		"screen_resolution":  "1920x1080",
		"viewport_height":    "919",
		"viewport_width":     "944",
		"viewport_size":      "944x919",
		"color_depth":        "24",
		"pixel_ratio":        "1.25",
		"current_url":        fmt.Sprintf("%s/c/%s", ZAIChatBaseURL, chatID),
		"pathname":           fmt.Sprintf("/c/%s", chatID),
		"search":             "",
		"hash":               "",
		"host":               "chat.z.ai",
		"hostname":           "chat.z.ai",
		"protocol":           "https:",
		"referrer":           "",
		"title":              "Z.ai - Free AI Chatbot & Agent powered by GLM-5 & GLM-4.7",
		"timezone_offset":    "-480",
		"local_time":         now.Format("2006-01-02T15:04:05.000Z"),
		"utc_time":           now.Format("Mon, 02 Jan 2006 15:04:05 GMT"),
		"is_mobile":          "false",
		"is_touch":           "false",
		"max_touch_points":   "10",
		"browser_name":       "Chrome",
		"os_name":            "Windows",
		"signature_timestamp": timestamp,
	}

	// Merge all params
	allParams := make(url.Values)
	for k, v := range core {
		allParams.Set(k, v)
	}
	for k, v := range extra {
		allParams.Set(k, v)
	}

	return allParams.Encode(), signature
}

// 提取消息内容为字符串
func extractMessageContent(content json.RawMessage) string {
	// 尝试解析为字符串
	var strContent string
	if err := json.Unmarshal(content, &strContent); err == nil {
		return strContent
	}

	// 尝试解析为数组格式
	var arrayContent []map[string]interface{}
	if err := json.Unmarshal(content, &arrayContent); err == nil {
		var texts []string
		for _, item := range arrayContent {
			if itemType, ok := item["type"].(string); ok && itemType == "text" {
				if text, ok := item["text"].(string); ok {
					texts = append(texts, text)
				}
			}
		}
		return strings.Join(texts, "\n")
	}

	return string(content)
}

// 扁平化消息为智谱API需要的格式
// 把所有消息合并成 <ROLE>content</ROLE> 格式，作为单个user消息发送
func flattenMessagesForZAI(messages []OpenAIMessage) []ZAIBlockMessage {
	var parts []string
	for _, msg := range messages {
		role := strings.ToUpper(msg.Role)
		content := extractMessageContent(msg.Content)
		parts = append(parts, fmt.Sprintf("<%s>%s</%s>", role, content, role))
	}

	// 返回单个user消息，包含所有对话历史
	flattenedContent, _ := json.Marshal(strings.Join(parts, "\n"))
	return []ZAIBlockMessage{
		{
			Role:    "user",
			Content: flattenedContent,
		},
	}
}

// 提取用户消息作为签名prompt
func extractPromptFromMessages(messages []OpenAIMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return extractMessageContent(messages[i].Content)
		}
	}
	return ""
}

// 创建智谱聊天会话
func createZAIChat(token, model, initialContent string) (string, error) {
	msgID := uuid.New().String()
	ts := time.Now().Unix()

	// 截断过长的初始内容
	maxInitLen := 500
	initContent := initialContent
	if len(initContent) > maxInitLen {
		initContent = initContent[:maxInitLen] + "..."
	}

	reqBody := ZAIChatCreateRequest{}
	reqBody.Chat.ID = ""
	reqBody.Chat.Title = "新聊天"
	reqBody.Chat.Models = []string{model}
	reqBody.Chat.Params = make(map[string]interface{})
	reqBody.Chat.History = &ZAIChatHistory{
		Messages: map[string]ZAIMessage{
			msgID: {
				ID:          msgID,
				ParentID:    nil,
				ChildrenIDs: []string{},
				Role:        "user",
				Content:     initContent,
				Timestamp:   ts,
				Models:      []string{model},
			},
		},
		CurrentID: msgID,
	}
	reqBody.Chat.Tags = []string{}
	reqBody.Chat.Flags = []string{}
	reqBody.Chat.Features = []ZAIFeature{
		{Type: "tool_selector", Server: "tool_selector_h", Status: "hidden"},
	}
	reqBody.Chat.MCPServers = []interface{}{}
	reqBody.Chat.EnableThinking = true
	reqBody.Chat.AutoWebSearch = false
	reqBody.Chat.MessageVersion = 1
	reqBody.Chat.Extra = make(map[string]interface{})
	reqBody.Chat.Timestamp = time.Now().UnixMilli()

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", ZAIChatBaseURL+"/api/v1/chats/new", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Origin", ZAIChatBaseURL)
	req.Header.Set("Referer", ZAIChatBaseURL+"/")
	req.Header.Set("User-Agent", UserAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("create chat failed: status=%d body=%s", resp.StatusCode, string(body)[:min(500, len(body))])
	}

	var chatResp ZAIChatCreateResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	return chatResp.ID, nil
}

// 删除智谱聊天会话
func deleteZAIChat(token, chatID string) {
	req, err := http.NewRequest("DELETE", ZAIChatBaseURL+"/api/v1/chats/"+chatID, nil)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Origin", ZAIChatBaseURL)
	req.Header.Set("Referer", ZAIChatBaseURL+"/")

	client := &http.Client{Timeout: 10 * time.Second}
	client.Do(req)
}

// 发送聊天请求并流式返回
func streamZAIChatCompletion(token, chatID, model string, req OpenAIRequest, prompt string) (<-chan string, error) {
	outputChan := make(chan string, 100)

	// 获取用户ID（从token解析或使用默认值）
	userID := "guest"

	queryString, signature := buildQueryAndSignature(token, userID, prompt, chatID)

	msgID := uuid.New().String()
	userMsgID := uuid.New().String()

	// 使用扁平化的消息格式（Python版本的逻辑）
	zaiMessages := flattenMessagesForZAI(req.Messages)

	// 当前时间（东八区）
	now := time.Now().In(time.FixedZone("CST", 8*3600))
	variables := map[string]string{
		"{{USER_NAME}}":      "Guest",
		"{{USER_LOCATION}}":  "Unknown",
		"{{CURRENT_DATETIME}}": now.Format("2006-01-02 15:04:05"),
		"{{CURRENT_DATE}}":   now.Format("2006-01-02"),
		"{{CURRENT_TIME}}":   now.Format("15:04:05"),
		"{{CURRENT_WEEKDAY}}": now.Format("Monday"),
		"{{CURRENT_TIMEZONE}}": "Asia/Shanghai",
		"{{USER_LANGUAGE}}":  "zh-CN",
	}

	zaiReq := ZAICompletionRequest{
		Stream:                     true,
		Model:                      model,
		Messages:                   zaiMessages,
		SignaturePrompt:            prompt,
		Params:                     make(map[string]interface{}),
		Extra:                      make(map[string]interface{}),
		Variables:                  variables,
		ChatID:                     chatID,
		ID:                         msgID,
		CurrentUserMessageID:       userMsgID,
		CurrentUserMessageParentID: nil,
		Features: ZAICompletionFeatures{
			ImageGeneration: false,
			WebSearch:       false,
			AutoWebSearch:   false,
			PreviewMode:     true,
			Flags:           []string{},
			EnableThinking:  true,
		},
		BackgroundTasks: ZAIBackgroundTasks{
			TitleGeneration: true,
			TagsGeneration:  true,
		},
	}

	// 转换工具定义
	if len(req.Tools) > 0 {
		var zaiTools []ZAIBlockTool
		for _, tool := range req.Tools {
			if tool.Type == "function" {
				zaiTool := ZAIBlockTool{Type: "function"}
				zaiTool.Function.Name = tool.Function.Name
				zaiTool.Function.Description = tool.Function.Description
				zaiTool.Function.Parameters = tool.Function.Parameters
				zaiTools = append(zaiTools, zaiTool)
			}
		}
		zaiReq.Tools = zaiTools
	}

	jsonData, err := json.Marshal(zaiReq)
	if err != nil {
		close(outputChan)
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v2/chat/completions?%s", ZAIChatBaseURL, queryString)
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		close(outputChan)
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Accept-Language", "zh-CN")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("X-FE-Version", FEVersion)
	httpReq.Header.Set("X-Signature", signature)
	httpReq.Header.Set("Origin", ZAIChatBaseURL)
	httpReq.Header.Set("Referer", ZAIChatBaseURL+"/")
	httpReq.Header.Set("User-Agent", UserAgent)

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		close(outputChan)
		return nil, err
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		close(outputChan)
		return nil, fmt.Errorf("chat completion failed: status=%d body=%s", resp.StatusCode, string(body)[:min(500, len(body))])
	}

	go func() {
		defer close(outputChan)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			raw := strings.TrimPrefix(line, "data: ")
			if strings.TrimSpace(raw) == "[DONE]" {
				break
			}

			var event map[string]interface{}
			if err := json.Unmarshal([]byte(raw), &event); err != nil {
				continue
			}

			// 提取 data 字段
			data, ok := event["data"].(map[string]interface{})
			if !ok {
				continue
			}

			// 转换为 JSON 字符串发送
			dataJSON, _ := json.Marshal(data)
			outputChan <- string(dataJSON)

			// 检查是否完成
			if done, ok := data["done"].(bool); ok && done {
				break
			}
		}
	}()

	return outputChan, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 获取模型列表
func handleListModels(c *gin.Context) {
	// 尝试从Authorization header获取token（可选）
	token, _ := extractToken(c)

	// 动态获取模型列表
	modelList, err := fetchZAIModels(token)
	if err != nil {
		// 获取失败时使用默认列表
		modelList = getDefaultModels()
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   modelList,
	})
}

// 获取单个模型
func handleGetModel(c *gin.Context) {
	modelID := c.Param("model")

	// 尝试从Authorization header获取token（可选）
	token, _ := extractToken(c)

	// 动态获取模型列表
	modelList, err := fetchZAIModels(token)
	if err != nil {
		modelList = getDefaultModels()
	}

	for _, m := range modelList {
		if m["id"] == modelID {
			c.JSON(http.StatusOK, m)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
}

// Chat completions 处理函数
func handleChatCompletions(c *gin.Context) {
	var req OpenAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从 Authorization header 获取 token
	token, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 默认模型
	if req.Model == "" {
		req.Model = "glm-5"
	}

	// 提取用户消息作为签名prompt
	prompt := extractPromptFromMessages(req.Messages)
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No user message found"})
		return
	}

	// 创建聊天会话
	chatID, err := createZAIChat(token, req.Model, prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create chat: %v", err)})
		return
	}

	// 确保在最后删除聊天会话
	defer deleteZAIChat(token, chatID)

	// 发送聊天请求
	dataChan, err := streamZAIChatCompletion(token, chatID, req.Model, req, prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start chat: %v", err)})
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	responseID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	created := time.Now().Unix()

	// 发送初始角色
	initialResp := OpenAIStreamResponse{
		ID:      responseID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []OpenAIStreamChoice{
			{
				Index: 0,
				Delta: map[string]interface{}{
					"role": "assistant",
				},
				FinishReason: nil,
			},
		},
	}
	initialJSON, _ := json.Marshal(initialResp)
	c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(initialJSON)))
	c.Writer.Flush()

	// 流式发送内容
	for dataJSON := range dataChan {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
			continue
		}

		phase, _ := data["phase"].(string)
		deltaContent, _ := data["delta_content"].(string)

		if deltaContent == "" {
			continue
		}

		// 根据阶段处理内容
		// thinking阶段使用reasoning_content字段，answer阶段使用content字段
		var delta map[string]interface{}
		if phase == "thinking" {
			delta = map[string]interface{}{
				"reasoning_content": deltaContent,
			}
		} else {
			delta = map[string]interface{}{
				"content": deltaContent,
			}
		}

		streamResp := OpenAIStreamResponse{
			ID:      responseID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   req.Model,
			Choices: []OpenAIStreamChoice{
				{
					Index:        0,
					Delta:        delta,
					FinishReason: nil,
				},
			},
		}

		respJSON, _ := json.Marshal(streamResp)
		c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(respJSON)))
		c.Writer.Flush()
	}

	// 发送结束标记
	finishReason := "stop"
	finalResp := OpenAIStreamResponse{
		ID:      responseID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []OpenAIStreamChoice{
			{
				Index:        0,
				Delta:        map[string]interface{}{},
				FinishReason: &finishReason,
			},
		},
	}
	finalJSON, _ := json.Marshal(finalResp)
	c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(finalJSON)))
	c.Writer.WriteString("data: [DONE]\n\n")
	c.Writer.Flush()
}

// RegisterRoutes 注册聊天模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/models", handleListModels)
	r.GET("/models/:model", handleGetModel)
	r.POST("/chat/completions", handleChatCompletions)
}
