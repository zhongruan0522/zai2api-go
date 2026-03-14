package ocr

type UpstreamResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Data      struct {
		TaskID          string        `json:"task_id"`
		Status          string        `json:"status"`
		FileName        string        `json:"file_name"`
		FileSize        int64         `json:"file_size"`
		FileType        string        `json:"file_type"`
		FileURL         string        `json:"file_url"`
		CreatedAt       string        `json:"created_at"`
		MarkdownContent string        `json:"markdown_content"`
		JsonContent     string        `json:"json_content"`
		Layout          []LayoutBlock `json:"layout"`
		DataInfo        *DataInfo     `json:"data_info"`
	} `json:"data"`
}

type LayoutBlock struct {
	BlockContent string  `json:"block_content"`
	BBox         []int   `json:"bbox"`
	BlockID      int     `json:"block_id"`
	PageIndex    int     `json:"page_index"`
	BlockLabel   string  `json:"block_label"`
	Score        float64 `json:"score"`
}

type DataInfo struct {
	Pages    []PageSize `json:"pages"`
	NumPages int        `json:"num_pages"`
}

type PageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type APIResponse struct {
	TaskID         string            `json:"task_id"`
	Message        string            `json:"message"`
	Status         string            `json:"status"`
	WordsResultNum int               `json:"words_result_num"`
	WordsResult    []WordsResultItem `json:"words_result"`
}

type WordsResultItem struct {
	Location    Location    `json:"location"`
	Words       string      `json:"words"`
	Probability Probability `json:"probability"`
}

type Location struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Probability struct {
	Average  float64 `json:"average"`
	Variance float64 `json:"variance"`
	Min      float64 `json:"min"`
}
