package server

import (
	"net/http"
	"page-zen/internal/utils"

	"github.com/gin-gonic/gin"
)

// ArticleRequest represents the request body for article extraction
type ArticleRequest struct {
	URL             string `json:"url" binding:"required"`
	IncludeMarkdown bool   `json:"include_markdown,omitempty"`
}

// ArticleResponse represents the response for article extraction
type ArticleResponse struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Markdown    string `json:"markdown,omitempty"`
	Author      string `json:"author,omitempty"`
	Excerpt     string `json:"excerpt,omitempty"`
	Length      int    `json:"length"`
	PublishedAt string `json:"published_at,omitempty"`
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
}

// ExtractArticleHandler handles article extraction requests
func (s *Server) ExtractArticleHandler(c *gin.Context) {
	s.logger.Info("ExtractArticleHandler called")

	var req ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Errorw("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, ArticleResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	s.logger.Infow("Processing article extraction request", "url", req.URL, "include_markdown", req.IncludeMarkdown)

	// Extract article content using the enhanced cleaner
	cleanedArticle := utils.GetCleanedArticle(req.URL)

	if cleanedArticle.Title == "" && cleanedArticle.Content == "" {
		s.logger.Warnw("Failed to extract article content", "url", req.URL)
		c.JSON(http.StatusInternalServerError, ArticleResponse{
			URL:     req.URL,
			Success: false,
			Message: "Failed to extract article content",
		})
		return
	}

	response := ArticleResponse{
		URL:         cleanedArticle.URL,
		Title:       cleanedArticle.Title,
		Content:     cleanedArticle.Content,
		Author:      cleanedArticle.Author,
		Excerpt:     cleanedArticle.Excerpt,
		Length:      cleanedArticle.Length,
		PublishedAt: cleanedArticle.PublishedAt,
		Success:     true,
	}

	// Include markdown if requested
	if req.IncludeMarkdown {
		response.Markdown = cleanedArticle.Markdown
	}

	s.logger.Infow("Successfully extracted article",
		"url", req.URL,
		"title", cleanedArticle.Title,
		"content_length", cleanedArticle.Length,
		"include_markdown", req.IncludeMarkdown,
	)

	c.JSON(http.StatusOK, response)
}

// ExtractArticleSimpleHandler handles simple GET requests for article extraction
func (s *Server) ExtractArticleSimpleHandler(c *gin.Context) {
	s.logger.Info("ExtractArticleSimpleHandler called")

	url := c.Query("url")
	if url == "" {
		s.logger.Warn("URL parameter missing")
		c.JSON(http.StatusBadRequest, ArticleResponse{
			Success: false,
			Message: "URL parameter is required",
		})
		return
	}

	includeMarkdown := c.Query("markdown") == "true"

	s.logger.Infow("Processing simple article extraction", "url", url, "include_markdown", includeMarkdown)

	// Extract article content using the enhanced cleaner
	cleanedArticle := utils.GetCleanedArticle(url)

	if cleanedArticle.Title == "" && cleanedArticle.Content == "" {
		s.logger.Warnw("Failed to extract article content", "url", url)
		c.JSON(http.StatusInternalServerError, ArticleResponse{
			URL:     url,
			Success: false,
			Message: "Failed to extract article content",
		})
		return
	}

	response := ArticleResponse{
		URL:         cleanedArticle.URL,
		Title:       cleanedArticle.Title,
		Content:     cleanedArticle.Content,
		Author:      cleanedArticle.Author,
		Excerpt:     cleanedArticle.Excerpt,
		Length:      cleanedArticle.Length,
		PublishedAt: cleanedArticle.PublishedAt,
		Success:     true,
	}

	// Include markdown if requested
	if includeMarkdown {
		response.Markdown = cleanedArticle.Markdown
	}

	s.logger.Infow("Successfully extracted article via GET",
		"url", url,
		"title", cleanedArticle.Title,
		"content_length", cleanedArticle.Length,
		"include_markdown", includeMarkdown,
	)

	c.JSON(http.StatusOK, response)
}
