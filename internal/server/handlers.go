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
	URL         string               `json:"url"`
	Title       string               `json:"title"`
	Content     string               `json:"content"`
	Markdown    string               `json:"markdown,omitempty"`
	Author      string               `json:"author,omitempty"`
	Excerpt     string               `json:"excerpt,omitempty"`
	Length      int                  `json:"length"`
	PublishedAt string               `json:"published_at,omitempty"`
	OpenGraph   *utils.OpenGraphData `json:"open_graph,omitempty"`
	Success     bool                 `json:"success"`
	Message     string               `json:"message,omitempty"`
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
		OpenGraph:   cleanedArticle.OpenGraph,
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
		OpenGraph:   cleanedArticle.OpenGraph,
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

// OpenGraphRequest represents the request body for Open Graph extraction
type OpenGraphRequest struct {
	URL string `json:"url" binding:"required"`
}

// OpenGraphResponse represents the response for Open Graph extraction
type OpenGraphResponse struct {
	URL       string               `json:"url"`
	OpenGraph *utils.OpenGraphData `json:"open_graph,omitempty"`
	Success   bool                 `json:"success"`
	Message   string               `json:"message,omitempty"`
}

// ExtractOpenGraphHandler handles Open Graph extraction requests via POST
func (s *Server) ExtractOpenGraphHandler(c *gin.Context) {
	s.logger.Info("ExtractOpenGraphHandler called")

	var req OpenGraphRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Errorw("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, OpenGraphResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	s.logger.Infow("Processing Open Graph extraction request", "url", req.URL)

	// Extract Open Graph data
	openGraphData := utils.GetOpenGraphData(req.URL)

	if openGraphData == nil || openGraphData.Title == "" {
		s.logger.Warnw("Failed to extract Open Graph data", "url", req.URL)
		c.JSON(http.StatusInternalServerError, OpenGraphResponse{
			URL:     req.URL,
			Success: false,
			Message: "Failed to extract Open Graph data",
		})
		return
	}

	response := OpenGraphResponse{
		URL:       req.URL,
		OpenGraph: openGraphData,
		Success:   true,
	}

	s.logger.Infow("Successfully extracted Open Graph data",
		"url", req.URL,
		"title", openGraphData.Title,
		"description_length", len(openGraphData.Description),
	)

	c.JSON(http.StatusOK, response)
}

// ExtractOpenGraphSimpleHandler handles simple GET requests for Open Graph extraction
func (s *Server) ExtractOpenGraphSimpleHandler(c *gin.Context) {
	s.logger.Info("ExtractOpenGraphSimpleHandler called")

	url := c.Query("url")
	if url == "" {
		s.logger.Warn("URL parameter missing")
		c.JSON(http.StatusBadRequest, OpenGraphResponse{
			Success: false,
			Message: "URL parameter is required",
		})
		return
	}

	s.logger.Infow("Processing simple Open Graph extraction", "url", url)

	// Extract Open Graph data
	openGraphData := utils.GetOpenGraphData(url)

	if openGraphData == nil || openGraphData.Title == "" {
		s.logger.Warnw("Failed to extract Open Graph data", "url", url)
		c.JSON(http.StatusInternalServerError, OpenGraphResponse{
			URL:     url,
			Success: false,
			Message: "Failed to extract Open Graph data",
		})
		return
	}

	response := OpenGraphResponse{
		URL:       url,
		OpenGraph: openGraphData,
		Success:   true,
	}

	s.logger.Infow("Successfully extracted Open Graph data via GET",
		"url", url,
		"title", openGraphData.Title,
		"description_length", len(openGraphData.Description),
	)

	c.JSON(http.StatusOK, response)
}
