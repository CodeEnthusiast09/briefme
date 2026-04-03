package services

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

type ScraperService struct {
	httpClient *http.Client
}

func NewScraperService() *ScraperService {
	return &ScraperService{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

var blockedTags = []string{
	"script",
	"style",
	"nav",
	"header",
	"footer",
	"aside",
	"iframe",
	"noscript",
	"figure",
	"form",
	"button",
}

// ScrapeURL fetches the page at the given URL, strips HTML noise,
// and returns clean readable text suitable for sending to Gemini.
//
// Web scraping is inherently messy — every website structures its HTML
// differently. Our strategy is:
//  1. Remove known noise tags entirely
//  2. Try to find the main article container first (most news sites use
//     semantic tags like <article> or common class names)
//  3. Fall back to <body> if no article container is found
//  4. Extract text, clean whitespace, truncate to a safe length
func (s *ScraperService) ScrapeURL(pageURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BriefMe/1.0)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("URL returned status %d — page may be paywalled or blocked", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	for _, tag := range blockedTags {
		doc.Find(tag).Remove()
	}

	articleSelectors := []string{
		"article",
		"[role='main']",
		".article-body",
		".article-content",
		".article__body",
		".story-body",
		".post-content",
		".entry-content",
		"main",
	}

	var contentNode *goquery.Selection
	for _, selector := range articleSelectors {
		node := doc.Find(selector)
		if node.Length() > 0 {
			contentNode = node.First()
			break
		}
	}

	if contentNode == nil || contentNode.Length() == 0 {
		contentNode = doc.Find("body")
	}

	rawText := contentNode.Text()
	cleaned := cleanText(rawText)

	if cleaned == "" {
		return "", fmt.Errorf("no readable content found at URL — site may require JavaScript or login")
	}

	const maxChars = 12000
	if utf8.RuneCountInString(cleaned) > maxChars {
		runes := []rune(cleaned)
		cleaned = string(runes[:maxChars])
	}

	return cleaned, nil
}

func cleanText(raw string) string {
	lines := strings.Split(raw, "\n")

	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, " ")
}
