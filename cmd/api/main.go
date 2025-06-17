package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"page-zen/internal/logger"
	"page-zen/internal/server"

	"go.uber.org/zap"
)

func gracefulShutdown(apiServer *http.Server, logger *zap.SugaredLogger, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	logger.Info("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Errorw("Server forced to shutdown with error", "error", err)
	}

	logger.Info("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	// Initialize logger
	zapLogger, err := logger.NewSugaredLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting Page Zen API server")

	server := server.NewServer()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, zapLogger, done)

	zapLogger.Info("Server starting to listen and serve")
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		zapLogger.Fatalw("HTTP server error", "error", err)
	}

	// Wait for the graceful shutdown to complete
	<-done
	zapLogger.Info("Graceful shutdown complete.")
}
