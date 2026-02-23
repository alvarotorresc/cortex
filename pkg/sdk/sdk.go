// Package sdk provides a minimal API for building Cortex plugins.
//
// Plugin authors implement the CortexPlugin interface and call sdk.Serve
// to start the plugin process. The SDK handles all gRPC and go-plugin
// wiring automatically.
//
// Example:
//
//	func main() {
//		sdk.Serve(&MyPlugin{})
//	}
package sdk

import (
	goplugin "github.com/hashicorp/go-plugin"

	cortexplugin "github.com/alvarotorresc/cortex/internal/plugin"
)

// Re-export types so plugin authors only import the SDK package.
type (
	// CortexPlugin is the interface that all plugins must implement.
	CortexPlugin = cortexplugin.CortexPlugin

	// Manifest represents a plugin's metadata.
	Manifest = cortexplugin.Manifest

	// APIRequest represents an incoming API request routed to a plugin.
	APIRequest = cortexplugin.APIRequest

	// APIResponse represents a plugin's response to an API request.
	APIResponse = cortexplugin.APIResponse
)

// Serve starts the plugin subprocess and serves over gRPC.
// This function blocks until the host process disconnects.
// Plugin authors call this as the only line in main():
//
//	sdk.Serve(&MyPlugin{})
func Serve(impl CortexPlugin) {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: cortexplugin.Handshake,
		Plugins: map[string]goplugin.Plugin{
			"cortex_plugin": &cortexplugin.CortexGRPCPlugin{Impl: impl},
		},
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}
