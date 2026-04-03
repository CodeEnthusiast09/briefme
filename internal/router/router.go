package router

import (
	"net/http"

	"github.com/CodeEnthusiast09/briefme-backend/internal/handlers"
	"github.com/gin-gonic/gin"
)

func Setup(
	summarizeHandler *handlers.SummarizeHandler,
	chatHandler *handlers.ChatHandler,
	allowedOrigin string,
) *gin.Engine {
	r := gin.New()

	// gin.New() gives you a bare router with no middleware.
	// We add our own explicitly so we control exactly what runs.
	// gin.Default() would add Logger and Recovery automatically,
	// but being explicit is better practice — you know what's active.

	// Recovery middleware catches any panic anywhere in a handler
	// and returns a 500 instead of crashing the entire server.
	// A panic in one request should never take down the whole app.
	r.Use(gin.Recovery())

	// Logger middleware prints each request: method, path, status, latency.
	// Useful during development and in Railway logs.
	r.Use(gin.Logger())

	// CORS middleware.
	// CORS (Cross-Origin Resource Sharing) is a browser security mechanism.
	// When your Vite frontend (on vercel.app) calls your Go backend
	// (on railway.app), the browser sees two different origins and blocks
	// the request unless the backend explicitly says "I allow this origin".
	// This middleware adds the right headers to every response.
	r.Use(corsMiddleware(allowedOrigin))

	// Health check endpoint.
	// Railway's healthcheck hits this to know the app is running.
	// It returns 200 with a simple JSON body — no logic needed.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes grouped under /api.
	// Grouping keeps route definitions clean and makes it easy to add
	// versioning later (e.g. /api/v2/summarize) without touching every route.
	api := r.Group("/api")
	{
		api.POST("/summarize", summarizeHandler.Handle)
		api.POST("/chat", chatHandler.Handle)
	}

	return r
}

// corsMiddleware returns a Gin middleware function that sets CORS headers.
// We write this ourselves rather than pulling in a third-party CORS library
// because our needs are simple — one allowed origin, standard methods.
// Adding a dependency for 5 lines of code is unnecessary.
func corsMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// OPTIONS is a "preflight" request — before the browser sends a POST
		// with a JSON body, it first sends an OPTIONS request to ask:
		// "is this cross-origin request allowed?"
		// We respond with 204 (No Content) and the CORS headers above.
		// If we don't handle OPTIONS, the browser never sends the actual request.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// c.Next() passes control to the next middleware or handler.
		// Without this, the request stops here and never reaches your route handler.
		c.Next()
	}
}
