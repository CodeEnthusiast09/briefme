package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/CodeEnthusiast09/briefme-backend/internal/models"
	"google.golang.org/genai"
)

type GeminiService struct {
	client *genai.Client
	model  string
}

func NewGeminiService(apiKey string) (*GeminiService, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiService{
		client: client,
		model:  "gemini-2.5-flash",
	}, nil
}

func summarizePrompt(content string) string {
	return fmt.Sprintf(`You are BriefMe, an expert news summarizer. Analyze the following article content and respond with ONLY a valid JSON object — no markdown, no backticks, no explanation outside the JSON.

The JSON must follow this exact structure:
{
  "title": "A clear, concise title for this content",
  "summary": "A 2-3 sentence summary of the core message",
  "key_points": ["point one", "point two", "point three"],
  "sentiment": "positive | neutral | negative",
  "reading_time_saved": "X mins"
}

Rules:
- key_points must have exactly 3 items
- sentiment must be exactly one of: positive, neutral, negative
- reading_time_saved should estimate how long the full content would take to read at 200 words per minute, then format as "X mins"
- Do not include any text outside the JSON object

Article content:
%s`, content)
}

func (g *GeminiService) Summarize(ctx context.Context, content string) (*models.SummarizeResponse, error) {
	prompt := summarizePrompt(content)

	result, err := g.client.Models.GenerateContent(
		ctx,
		g.model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("gemini generate content failed: %w", err)
	}

	raw := result.Text()

	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var response models.SummarizeResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response as JSON: %w\nraw response: %s", err, raw)
	}

	return &response, nil
}

func (g *GeminiService) Chat(ctx context.Context, message string, history []models.ChatMessage, articleContext string) (string, error) {
	contents := make([]*genai.Content, 0, len(history)+1)

	systemContext := fmt.Sprintf(
		`You are BriefMe, an AI assistant helping users understand a specific article or topic.
    Here is the article context you should base your answers on:

    %s

    Only answer questions related to this content. If asked something unrelated, politely redirect the user back to the article.`,
		articleContext,
	)

	contents = append(contents, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: systemContext}},
	})

	contents = append(contents, &genai.Content{
		Role:  "model",
		Parts: []*genai.Part{{Text: "Understood. I have read the article and I am ready to answer your questions about it."}},
	})

	// Append the actual conversation history from the frontend.
	for _, msg := range history {
		contents = append(contents, &genai.Content{
			Role:  msg.Role,
			Parts: []*genai.Part{{Text: msg.Content}},
		})
	}

	// Append the latest user message.
	contents = append(contents, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: message}},
	})

	result, err := g.client.Models.GenerateContent(
		ctx,
		g.model,
		contents,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("gemini chat failed: %w", err)
	}

	reply := strings.TrimSpace(result.Text())
	if reply == "" {
		return "", fmt.Errorf("gemini returned an empty reply")
	}

	return reply, nil
}

func (g *GeminiService) ExtractKeywords(ctx context.Context, query string) (string, error) {
	prompt := fmt.Sprintf(
		`Extract 2-3 concise news search keywords from the following question.
    Focus on the subject matter only — the country, person, company, or event being asked about.
    Strip question words (what, how, why, when, is, are) and vague words (happening, going on, situation, latest).
    Return ONLY the keywords as a single line of plain text. No explanation, no punctuation, no quotes.

    Examples:
    "what is happening in Nigeria?" → "Nigeria news"
    "how is the tech industry doing?" → "tech industry"
    "tell me about the war in Sudan" → "Sudan war"

    Question: %s`,
		query,
	)
	result, err := g.client.Models.GenerateContent(
		ctx,
		g.model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("gemini keyword extraction failed: %w", err)
	}

	keywords := strings.TrimSpace(result.Text())
	if keywords == "" {
		return query, nil
	}

	return keywords, nil
}
