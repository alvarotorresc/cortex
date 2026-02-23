package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/alvarotorresc/cortex/internal/db"
)

// dashboardRoutes registers host-level dashboard layout endpoints.
// These are not plugin routes -- they manage the grid layout across all plugins.
func dashboardRoutes(router chi.Router, hostDB *db.HostDB) {
	// GET /api/dashboard/layout -- returns all widget positions
	router.Get("/api/dashboard/layout", func(writer http.ResponseWriter, request *http.Request) {
		layouts, err := hostDB.GetDashboardLayouts()
		if err != nil {
			writeDashboardError(writer, http.StatusInternalServerError, "DB_ERROR", "failed to get dashboard layouts")
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]interface{}{"data": layouts})
	})

	// PUT /api/dashboard/layout -- save all widget positions (full replace)
	router.Put("/api/dashboard/layout", func(writer http.ResponseWriter, request *http.Request) {
		var body struct {
			Widgets []db.WidgetLayout `json:"widgets"`
		}

		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			writeDashboardError(writer, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		for index := range body.Widgets {
			body.Widgets[index].UpdatedAt = now
			if body.Widgets[index].CreatedAt == "" {
				body.Widgets[index].CreatedAt = now
			}
		}

		if err := hostDB.SaveDashboardLayouts(body.Widgets); err != nil {
			writeDashboardError(writer, http.StatusInternalServerError, "DB_ERROR", "failed to save dashboard layouts")
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]interface{}{
			"data": body.Widgets,
			"meta": map[string]interface{}{"saved_at": now},
		})
	})
}

// writeDashboardError writes a standardized error JSON response for dashboard endpoints.
func writeDashboardError(writer http.ResponseWriter, statusCode int, code string, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	json.NewEncoder(writer).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}
