package server

import (
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware creates a gin middleware for logging requests using zap
func (s *Server) LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom log format using zap
		s.logger.Infow("HTTP Request",
			"client_ip", param.ClientIP,
			"timestamp", param.TimeStamp.Format(time.RFC3339),
			"method", param.Method,
			"path", param.Path,
			"protocol", param.Request.Proto,
			"status_code", param.StatusCode,
			"latency", param.Latency,
			"user_agent", param.Request.UserAgent(),
			"error", param.ErrorMessage,
		)
		return ""
	})
}

// RecoveryMiddleware creates a gin recovery middleware using zap
func (s *Server) RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			s.logger.Errorw("Panic recovered",
				"error", err,
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)
		}
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
	})
}
