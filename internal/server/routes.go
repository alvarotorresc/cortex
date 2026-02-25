package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/alvarotorresc/cortex/internal/config"
	"github.com/alvarotorresc/cortex/internal/db"
	"github.com/alvarotorresc/cortex/internal/plugin"
)

// HealthResponse is the JSON structure returned by the health check endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// NewRouter creates and configures a chi router with middleware and routes.
// It wires the plugin registry, loader, host database, and static asset serving.
func NewRouter(cfg *config.Config, registry *plugin.Registry, loader *plugin.Loader, hostDB *db.HostDB) *chi.Mux {
	router := chi.NewRouter()

	// Middleware stack
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	router.Get("/api/health", handleHealth)

	// Plugin API routes (list, install, uninstall, reload, widget data, proxy)
	pluginAPIRoutes(router, registry, loader)

	// Dashboard layout routes (host-level)
	dashboardRoutes(router, hostDB)

	// Serve plugin frontend assets
	pluginAssetsDir := http.Dir(cfg.PluginDir)
	router.Handle("/plugins/*", http.StripPrefix("/plugins/", http.FileServer(pluginAssetsDir)))

	// Serve main frontend (SvelteKit build)
	frontendDir := http.Dir(cfg.FrontendDir)
	router.Handle("/*", http.FileServer(frontendDir))

	return router
}

// handleHealth returns the server health status.
func handleHealth(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	response := HealthResponse{Status: "ok"}
	_ = json.NewEncoder(writer).Encode(response)
}
