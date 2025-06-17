package server

import (
	"net/http"
	"page-zen/internal/utils"

	"github.com/gin-gonic/gin"
)

// ArticleRequest represents the request body for article extraction
type ArticleRequest struct {
	URL string `json:"url" binding:"required"`
}

// ArticleResponse represents the response for article extraction
type ArticleResponse struct {
	URL     string `json:"url"`
	Content string `json:"content"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ExtractArticleHandler handles article extraction requests
func (s *Server) ExtractArticleHandler(c *gin.Context) {
	s.logger.Info("ExtractArticleHandler called")

	var req ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Errorw("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, ArticleResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	s.logger.Infow("Processing article extraction request", "url", req.URL)

	// Extract article content
	content := utils.GetReadableArticle(req.URL)

	if content == "" {
		s.logger.Warnw("Failed to extract article content", "url", req.URL)
		c.JSON(http.StatusInternalServerError, ArticleResponse{
			URL:     req.URL,
			Success: false,
			Message: "Failed to extract article content",
		})
		return
	}

	s.logger.Infow("Successfully extracted article",
		"url", req.URL,
		"content_length", len(content),
	)

	c.JSON(http.StatusOK, ArticleResponse{
		URL:     req.URL,
		Content: content,
		Success: true,
	})
}
