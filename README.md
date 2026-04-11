# BriefMe Backend

A Go REST API that summarizes news articles and URLs using Google Gemini, with conversational follow-up support.

## Tech Stack

- **Go** with [Gin](https://github.com/gin-gonic/gin) — HTTP framework
- **Gemini 2.5 Flash** — summarization and chat
- **NewsAPI** — article fetching for topic-based queries
- **goquery** — HTML scraping for URL-based queries
- **Docker** — containerized deployment on Railway

## Project Structure

```
.
├── cmd/
│   └── main.go                  # Entry point, dependency wiring
├── internal/
│   ├── config/config.go         # Environment variable loading
│   ├── handlers/
│   │   ├── summarize.go         # POST /api/summarize
│   │   └── chat.go              # POST /api/chat
│   ├── services/
│   │   ├── gemini.go            # Gemini summarization + chat
│   │   ├── newsapi.go           # NewsAPI article fetching
│   │   └── scraper.go           # URL HTML scraping
│   ├── models/types.go          # Request/response types
│   └── router/router.go         # Route definitions + CORS
├── Dockerfile
└── railway.toml
```

## API Endpoints

### `POST /api/summarize`

Accepts a URL or topic and returns a structured summary.

```json
// Request
{ "type": "url", "input": "https://..." }
{ "type": "topic", "input": "what is happening in Nigeria?" }

// Response
{
  "title": "...",
  "summary": "...",
  "key_points": ["...", "...", "..."],
  "sentiment": "positive | neutral | negative",
  "reading_time_saved": "X mins",
  "sources": [{ "title": "...", "url": "..." }]
}
```

### `POST /api/chat`

Accepts a follow-up message with conversation history and article context.

```json
// Request
{
  "message": "where did this happen?",
  "history": [{ "role": "user", "content": "..." }, { "role": "model", "content": "..." }],
  "context": "original summary text..."
}

// Response
{ "reply": "..." }
```

### `GET /health`

Health check. Returns `{ "status": "ok" }`.

## How the Two Flows Work

**URL flow:** scrape page → strip HTML noise → send to Gemini → return structured summary.

**Topic flow:** clean query keywords locally → search NewsAPI (`qInTitle`) → combine up to 3 articles → send to Gemini → return structured summary with sources.

Gemini is called **once per request** — keyword extraction is done locally to preserve the free tier quota.

## Environment Variables

| Variable                       | Description                                                  |
| ------------------------------ | ------------------------------------------------------------ |
| `GOOGLE_GENERATIVE_AI_API_KEY` | Gemini API key from [AI Studio](https://aistudio.google.com) |
| `NEWS_API_KEY`                 | NewsAPI key from [newsapi.org](https://newsapi.org)          |
| `PORT`                         | Server port (default: `8080`)                                |
| `GIN_MODE`                     | `debug` locally, `release` in production                     |
| `ALLOWED_ORIGIN`               | Frontend URL for CORS (use `*` during development)           |

## Local Development

```bash
git clone https://github.com/CodeEnthusiast09/briefme-backend.git
cd briefme-backend
cp .env.example .env   # fill in your keys
go run ./cmd/
```

Server starts at `http://localhost:8080`.

## Deployment (Railway)

The `Dockerfile` is a multi-stage build — compiles the Go binary in Stage 1, runs it in a clean Alpine image in Stage 2. No database, no migrations, no entrypoint script needed.

Set these environment variables in Railway:

```
GOOGLE_GENERATIVE_AI_API_KEY=...
NEWS_API_KEY=...
PORT=8080
GIN_MODE=release
ALLOWED_ORIGIN=https://your-vercel-url.vercel.app
```

## Rate Limits

Gemini 2.5 Flash free tier: **10 RPM, 250 requests/day**. If you hit the limit, the API returns a `429` with a message to retry after a short wait. Avoid rapid consecutive requests during development.
