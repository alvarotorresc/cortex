package plugin

import (
	"github.com/hashicorp/go-plugin"
)

// CortexPlugin is the interface that all plugins must implement.
type CortexPlugin interface {
	GetManifest() (*Manifest, error)
	HandleAPI(request *APIRequest) (*APIResponse, error)
	GetWidgetData(slot string) ([]byte, error)
	Migrate(databasePath string) error
	Teardown() error
}

// Manifest represents a plugin's metadata.
type Manifest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Color       string   `json:"color"`
	Permissions []string `json:"permissions"`
}

// APIRequest represents an incoming API request for a plugin.
type APIRequest struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Body   []byte            `json:"body"`
	Query  map[string]string `json:"query"`
}

// APIResponse represents a plugin's API response.
type APIResponse struct {
	StatusCode  int    `json:"statusCode"`
	Body        []byte `json:"body"`
	ContentType string `json:"contentType"`
}

// Handshake is the shared handshake config for host and plugins.
// Both sides must agree on this for the connection to succeed.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "CORTEX_PLUGIN",
	MagicCookieValue: "cortex-v1",
}

// PluginMap is the map of plugin types the host can dispense.
var PluginMap = map[string]plugin.Plugin{
	"cortex_plugin": &CortexGRPCPlugin{},
}
