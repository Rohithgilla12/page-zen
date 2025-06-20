package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	// Create gin engine without default middleware
	r := gin.New()

	// Use custom middleware
	r.Use(s.RecoveryMiddleware())
	r.Use(s.LoggerMiddleware())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/", s.HelloWorldHandler)
	r.POST("/extract", s.ExtractArticleHandler)
	r.GET("/extract", s.ExtractArticleSimpleHandler)
	r.POST("/opengraph", s.ExtractOpenGraphHandler)
	r.GET("/opengraph", s.ExtractOpenGraphSimpleHandler)

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	s.logger.Info("HelloWorldHandler called")

	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}
