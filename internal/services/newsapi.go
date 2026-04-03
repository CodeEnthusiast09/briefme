package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const newsAPIBaseURL = "https://newsapi.org/v2/everything"

type NewsAPIService struct {
	apiKey     string
	httpClient *http.Client
}

type newsAPIResponse struct {
	Status   string           `json:"status"`
	Articles []newsAPIArticle `json:"articles"`
}

type newsAPIArticle struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	URL         string `json:"url"`
	Author      string `json:"author"`
	PublishedAt string `json:"publishedAt"`
	Source      struct {
		Name string `json:"name"`
	} `json:"source"`
}

func NewNewsAPIService(apiKey string) *NewsAPIService {
	return &NewsAPIService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *NewsAPIService) FetchArticles(topic string) (string, []struct{ Title, URL string }, error) {
	params := url.Values{}
	params.Set("qInTitle", topic)
	params.Set("language", "en")
	params.Set("sortBy", "publishedAt")
	params.Set("pageSize", "3")
	params.Set("apiKey", n.apiKey)

	fullURL := fmt.Sprintf("%s?%s", newsAPIBaseURL, params.Encode())

	resp, err := n.httpClient.Get(fullURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to call NewsAPI: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("NewsAPI returned status %d", resp.StatusCode)
	}

	var newsResp newsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode NewsAPI response: %w", err)
	}

	if newsResp.Status != "ok" {
		return "", nil, fmt.Errorf("NewsAPI returned non-ok status: %s", newsResp.Status)
	}

	if len(newsResp.Articles) == 0 {
		return "", nil, fmt.Errorf("no articles found for topic: %s", topic)
	}

	combinedContent := ""
	sources := make([]struct{ Title, URL string }, 0, len(newsResp.Articles))

	for i, article := range newsResp.Articles {
		articleText := fmt.Sprintf(
			"Article %d:\nSource: %s\nAuthor: %s\nPublished: %s\nTitle: %s\nDescription: %s\nContent: %s\n\n",
			i+1,
			article.Source.Name,
			article.Author,
			article.PublishedAt,
			article.Title,
			article.Description,
			article.Content,
		)

		combinedContent += articleText

		sources = append(sources, struct{ Title, URL string }{
			Title: article.Title,
			URL:   article.URL,
		})
	}

	return combinedContent, sources, nil
}
