package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/alvarotorresc/cortex/internal/plugin"
)

// pluginAPIRoutes registers all plugin-related API endpoints.
func pluginAPIRoutes(router chi.Router, registry *plugin.Registry) {
	// List installed plugins
	router.Get("/api/plugins", func(writer http.ResponseWriter, request *http.Request) {
		manifests := registry.List()
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]interface{}{"data": manifests})
	})

	// Plugin widget data
	router.Get("/api/plugins/{pluginID}/widget/{slot}", func(writer http.ResponseWriter, request *http.Request) {
		pluginID := chi.URLParam(request, "pluginID")
		slot := chi.URLParam(request, "slot")

		entry, ok := registry.Get(pluginID)
		if !ok {
			writePluginError(writer, http.StatusNotFound, "NOT_FOUND", "plugin not found")
			return
		}

		if entry.Plugin == nil {
			writePluginError(writer, http.StatusServiceUnavailable, "PLUGIN_UNAVAILABLE", "plugin is not running")
			return
		}

		data, err := entry.Plugin.GetWidgetData(slot)
		if err != nil {
			writePluginError(writer, http.StatusInternalServerError, "PLUGIN_ERROR", "failed to get widget data")
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(data)
	})

	// Proxy all other plugin API requests
	router.HandleFunc("/api/plugins/{pluginID}/*", func(writer http.ResponseWriter, request *http.Request) {
		pluginID := chi.URLParam(request, "pluginID")

		entry, ok := registry.Get(pluginID)
		if !ok {
			writePluginError(writer, http.StatusNotFound, "NOT_FOUND", "plugin not found")
			return
		}

		if entry.Plugin == nil {
			writePluginError(writer, http.StatusServiceUnavailable, "PLUGIN_UNAVAILABLE", "plugin is not running")
			return
		}

		// Extract the sub-path after /api/plugins/{id}/
		fullPath := request.URL.Path
		prefix := "/api/plugins/" + pluginID + "/"
		subPath := strings.TrimPrefix(fullPath, prefix)

		body, _ := io.ReadAll(request.Body)

		query := make(map[string]string)
		for key, values := range request.URL.Query() {
			if len(values) > 0 {
				query[key] = values[0]
			}
		}

		response, err := entry.Plugin.HandleAPI(&plugin.APIRequest{
			Method: request.Method,
			Path:   "/" + subPath,
			Body:   body,
			Query:  query,
		})
		if err != nil {
			writePluginError(writer, http.StatusInternalServerError, "PLUGIN_ERROR", "plugin request failed")
			return
		}

		writer.Header().Set("Content-Type", response.ContentType)
		writer.WriteHeader(response.StatusCode)
		writer.Write(response.Body)
	})
}

// writePluginError writes a standardized error JSON response.
// It never exposes internal error details to the client.
func writePluginError(writer http.ResponseWriter, statusCode int, code string, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	json.NewEncoder(writer).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}
