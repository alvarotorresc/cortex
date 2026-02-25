package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alvarotorresc/cortex/internal/config"
	"github.com/alvarotorresc/cortex/internal/db"
	"github.com/alvarotorresc/cortex/internal/plugin"
)

const (
	readTimeout     = 10 * time.Second
	writeTimeout    = 30 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 10 * time.Second
)

// Start initializes and runs the HTTP server with graceful shutdown.
// It blocks until a termination signal is received (SIGINT or SIGTERM),
// then gracefully shuts down the server.
func Start(cfg *config.Config, registry *plugin.Registry, loader *plugin.Loader, hostDB *db.HostDB) error {
	router := NewRouter(cfg, registry, loader, hostDB)

	server := &http.Server{
		Addr:         cfg.Address(),
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Channel to listen for errors from the server goroutine.
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Cortex server starting on %s", cfg.Address())
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for OS signals.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal or a server error.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Received signal %v, starting graceful shutdown", sig)

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			_ = server.Close()
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		log.Println("Server stopped gracefully")
	}

	return nil
}
