package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"page-zen/internal/logger"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
)

type Server struct {
	port   int
	logger *zap.SugaredLogger
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Initialize logger
	zapLogger, err := logger.NewSugaredLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	NewServer := &Server{
		port:   port,
		logger: zapLogger,
	}

	// Log server initialization
	NewServer.logger.Infof("Initializing server on port %d", port)

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
