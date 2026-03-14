package image

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type GenerateRequest struct {
	Prompt           string `json:"prompt"`
	Ratio            string `json:"ratio"`
	Resolution       string `json:"resolution"`
	RmLabelWatermark bool   `json:"rm_label_watermark"`
}

type UpstreamResponse struct {
	Code      int               `json:"code"`
	Message   string            `json:"message"`
	Data      UpstreamImageData `json:"data"`
	Timestamp int64             `json:"timestamp"`
}

type UpstreamImageData struct {
	Image UpstreamImage `json:"image"`
}

type UpstreamImage struct {
	ImageID    string    `json:"image_id"`
	Prompt     string    `json:"prompt"`
	Size       string    `json:"size"`
	Ratio      string    `json:"ratio"`
	Resolution string    `json:"resolution"`
	ImageURL   string    `json:"image_url"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
}

type OpenAIResponse struct {
	Created int64            `json:"created"`
	Data    []OpenAIDataItem `json:"data"`
}

type OpenAIDataItem struct {
	URL           string `json:"url,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

var modelPattern = regexp.MustCompile(`^gemini-3-pro-image(?:-(\d+[kK]))?(?:-(.+))?$`)

func ParseModelToParams(model string) (resolution, ratio string) {
	matches := modelPattern.FindStringSubmatch(model)
	if len(matches) < 3 {
		return "1K", "1:1"
	}

	resolution = "1K"
	if matches[1] != "" {
		resolution = strings.ToLower(matches[1])
	}

	ratio = "1:1"
	if matches[2] != "" {
		ratio = normalizeRatio(matches[2])
	}

	return resolution, ratio
}

func normalizeRatio(r string) string {
	r = strings.ReplaceAll(r, "-", ":")
	if !strings.Contains(r, ":") {
		switch r {
		case "1to1":
			return "1:1"
		case "3to4":
			return "3:4"
		case "4to3":
			return "4:3"
		case "16to9":
			return "16:9"
		case "9to16":
			return "9:16"
		case "21to9":
			return "21:9"
		case "9to21":
			return "9:21"
		}
	}
	return r
}

func ConvertToChatResponse(upstream *UpstreamResponse, model string) *ChatResponse {
	imageURL := upstream.Data.Image.ImageURL

	return &ChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", upstream.Data.Image.ImageID),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []ChatChoice{
			{
				Index:        0,
				FinishReason: "stop",
				Message: ChatMessage{
					Role:    "assistant",
					Content: fmt.Sprintf("![image](%s)", imageURL),
				},
			},
		},
	}
}
