package plugin

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	goplugin "github.com/hashicorp/go-plugin"
)

// Loader discovers and launches plugin subprocesses.
type Loader struct {
	pluginDir string
	dataDir   string
	registry  *Registry
}

// NewLoader creates a loader that scans pluginDir for plugins
// and stores runtime data in dataDir.
func NewLoader(pluginDir string, dataDir string, registry *Registry) *Loader {
	return &Loader{
		pluginDir: pluginDir,
		dataDir:   dataDir,
		registry:  registry,
	}
}

// LoadAll discovers plugins in pluginDir and starts them.
// Each plugin directory must contain a "plugin" binary and a "manifest.json" file.
func (l *Loader) LoadAll() error {
	entries, err := os.ReadDir(l.pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("No plugins directory found, skipping plugin loading")
			return nil
		}
		return fmt.Errorf("reading plugin directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if err := l.LoadPlugin(entry.Name()); err != nil {
			log.Printf("Failed to load plugin %s: %v", entry.Name(), err)
		}
	}

	return nil
}

// LoadPlugin starts a single plugin by its directory name.
func (l *Loader) LoadPlugin(id string) error {
	pluginPath := filepath.Join(l.pluginDir, id)
	binaryPath := filepath.Join(pluginPath, "plugin")
	manifestPath := filepath.Join(pluginPath, "manifest.json")

	// Read manifest from disk
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("reading manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("parsing manifest: %w", err)
	}

	// Ensure plugin data directory exists
	dataPath := filepath.Join(l.dataDir, "plugins", id)
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	// Launch plugin subprocess via go-plugin
	client := goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig:  Handshake,
		Plugins:          PluginMap,
		Cmd:              exec.Command(binaryPath),
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("connecting to plugin: %w", err)
	}

	raw, err := rpcClient.Dispense("cortex_plugin")
	if err != nil {
		client.Kill()
		return fmt.Errorf("dispensing plugin: %w", err)
	}

	cortexPlugin, ok := raw.(CortexPlugin)
	if !ok {
		client.Kill()
		return fmt.Errorf("plugin does not implement CortexPlugin interface")
	}

	// Run database migrations
	databasePath := filepath.Join(dataPath, "db.sqlite")
	if err := cortexPlugin.Migrate(databasePath); err != nil {
		client.Kill()
		return fmt.Errorf("running migrations: %w", err)
	}

	// Register plugin in the registry
	l.registry.Register(id, client, &manifest)
	entry, _ := l.registry.Get(id)
	entry.Plugin = cortexPlugin

	log.Printf("Plugin loaded: %s (%s v%s)", manifest.Name, manifest.ID, manifest.Version)
	return nil
}

// UnloadPlugin stops and unregisters a plugin by ID.
func (l *Loader) UnloadPlugin(id string) error {
	entry, ok := l.registry.Get(id)
	if !ok {
		return fmt.Errorf("plugin %s not found", id)
	}

	if entry.Plugin != nil {
		if err := entry.Plugin.Teardown(); err != nil {
			log.Printf("Warning: teardown failed for plugin %s: %v", id, err)
		}
	}

	l.registry.Unregister(id)
	log.Printf("Plugin unloaded: %s", id)
	return nil
}

// UnloadAll stops all registered plugins.
func (l *Loader) UnloadAll() {
	for _, manifest := range l.registry.List() {
		if err := l.UnloadPlugin(manifest.ID); err != nil {
			log.Printf("Error unloading plugin %s: %v", manifest.ID, err)
		}
	}
}
