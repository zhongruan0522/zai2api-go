package ocr

func ConvertResponse(upstream *UpstreamResponse) *APIResponse {
	wordsResult := make([]WordsResultItem, 0, len(upstream.Data.Layout))

	for _, block := range upstream.Data.Layout {
		if block.BlockLabel != "text" {
			continue
		}

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

	return &APIResponse{
		TaskID:         upstream.Data.TaskID,
		Message:        "成功",
		Status:         "succeeded",
		WordsResultNum: len(wordsResult),
		WordsResult:    wordsResult,
	}
}
