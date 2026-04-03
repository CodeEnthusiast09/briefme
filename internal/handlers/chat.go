package handlers

import (
	"fmt"
	"net/http"

	"github.com/CodeEnthusiast09/briefme-backend/internal/models"
	"github.com/CodeEnthusiast09/briefme-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	gemini *services.GeminiService
}

func NewChatHandler(gemini *services.GeminiService) *ChatHandler {
	return &ChatHandler{gemini: gemini}
}

func (h *ChatHandler) Handle(c *gin.Context) {
	var req models.ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid request: %s", err.Error()),
		})
		return
	}

	for i, msg := range req.History {
		if msg.Role != "user" && msg.Role != "model" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf(
					"history[%d].role must be 'user' or 'model', got '%s'",
					i, msg.Role,
				),
			})
			return
		}
		if msg.Content == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("history[%d].content must not be empty", i),
			})
			return
		}
	}

	reply, err := h.gemini.Chat(
		c.Request.Context(),
		req.Message,
		req.History,
		req.Context,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("chat failed: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		Reply: reply,
	})
}
