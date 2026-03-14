package image

import "time"

func ConvertResponse(upstream *UpstreamResponse) *OpenAIResponse {
	return &OpenAIResponse{
		Created: time.Now().Unix(),
		Data: []OpenAIDataItem{
			{
				URL:           upstream.Data.Image.ImageURL,
				RevisedPrompt: upstream.Data.Image.Prompt,
			},
		},
	}
}
