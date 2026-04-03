package models

type SummarizeRequest struct {
	Type  string `json:"type" binding:"required,oneof=url topic"`
	Input string `json:"input" binding:"required"`
}

type Source struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type SummarizeResponse struct {
	Title            string   `json:"title"`
	Summary          string   `json:"summary"`
	KeyPoints        []string `json:"key_points"`
	Sentiment        string   `json:"sentiment"`
	ReadingTimeSaved string   `json:"reading_time_saved"`
	Sources          []Source `json:"sources"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Message string        `json:"message" binding:"required"`
	History []ChatMessage `json:"history"`
	Context string        `json:"context" binding:"required"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}
