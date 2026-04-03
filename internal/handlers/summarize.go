package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/CodeEnthusiast09/briefme-backend/internal/models"
	"github.com/CodeEnthusiast09/briefme-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type SummarizeHandler struct {
	gemini  *services.GeminiService
	news    *services.NewsAPIService
	scraper *services.ScraperService
}

func NewSummarizeHandler(
	gemini *services.GeminiService,
	news *services.NewsAPIService,
	scraper *services.ScraperService,
) *SummarizeHandler {
	return &SummarizeHandler{
		gemini:  gemini,
		news:    news,
		scraper: scraper,
	}
}

func (h *SummarizeHandler) Handle(c *gin.Context) {
	var req models.SummarizeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid request: %s", err.Error()),
		})
		return
	}

	switch req.Type {
	case "url":
		h.handleURL(c, req.Input)
	case "topic":
		h.handleTopic(c, req.Input)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "type must be 'url' or 'topic'",
		})
	}
}

func (h *SummarizeHandler) handleURL(c *gin.Context, rawURL string) {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "input must be a valid http or https URL",
		})
		return
	}

	// Scrape the article content from the URL.
	content, err := h.scraper.ScrapeURL(rawURL)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": fmt.Sprintf("could not read page: %s", err.Error()),
		})
		return
	}

	// Send to Gemini for structured summarization.
	summary, err := h.gemini.Summarize(c.Request.Context(), content)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "rate limit") {
			status = http.StatusTooManyRequests // 429
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	summary.Sources = []models.Source{
		{
			Title: summary.Title,
			URL:   rawURL,
		},
	}

	c.JSON(http.StatusOK, summary)
}

func (h *SummarizeHandler) handleTopic(c *gin.Context, topic string) {
	// Ask Gemini to extract clean search keywords from the user's query.
	keywords := cleanTopic(topic)

	// Fetch articles from NewsAPI using the extracted keywords.
	combinedContent, rawSources, err := h.news.FetchArticles(keywords)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": fmt.Sprintf("could not fetch news articles: %s", err.Error()),
		})
		return
	}

	// Send the combined article content to Gemini.
	summary, err := h.gemini.Summarize(c.Request.Context(), combinedContent)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "rate limit") {
			status = http.StatusTooManyRequests // 429
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	// Map rawSources to models.Source and attach to the response.
	sources := make([]models.Source, 0, len(rawSources))
	for _, s := range rawSources {
		sources = append(sources, models.Source{
			Title: s.Title,
			URL:   s.URL,
		})
	}
	summary.Sources = sources

	c.JSON(http.StatusOK, summary)
}

// cleanTopic does basic keyword extraction without an LLM call.
// It strips question words and punctuation, leaving the core subject.
func cleanTopic(topic string) string {
	// Remove punctuation
	cleaned := strings.NewReplacer(
		"?", "",
		"!", "",
		".", "",
		",", "",
	).Replace(topic)

	// Strip common question/filler words
	stopWords := map[string]bool{
		"what": true, "is": true, "are": true, "was": true,
		"were": true, "how": true, "why": true, "when": true,
		"who": true, "where": true, "happening": true, "going": true,
		"on": true, "the": true, "a": true, "an": true,
		"in": true, "tell": true, "me": true, "about": true,
		"latest": true, "news": true, "situation": true,
	}

	words := strings.Fields(cleaned)
	keywords := make([]string, 0, len(words))
	for _, word := range words {
		if !stopWords[strings.ToLower(word)] {
			keywords = append(keywords, word)
		}
	}

	result := strings.Join(keywords, " ")
	if result == "" {
		return topic // nothing left after stripping, use original
	}
	return result
}
