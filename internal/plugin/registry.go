package plugin

import (
	"sync"

	goplugin "github.com/hashicorp/go-plugin"
)

// RegistryEntry holds a running plugin's client and manifest.
type RegistryEntry struct {
	Client   *goplugin.Client
	Plugin   CortexPlugin
	Manifest *Manifest
}

// Registry manages active plugins in a thread-safe map.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*RegistryEntry
}

// NewRegistry creates an empty plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*RegistryEntry),
	}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(id string, client *goplugin.Client, manifest *Manifest) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.plugins[id] = &RegistryEntry{
		Client:   client,
		Manifest: manifest,
	}
}

// Get retrieves a plugin entry by ID. Returns false if not found.
func (r *Registry) Get(id string) (*RegistryEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.plugins[id]
	return entry, ok
}

// Unregister removes a plugin from the registry and kills its subprocess.
func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry, ok := r.plugins[id]; ok {
		if entry.Client != nil {
			entry.Client.Kill()
		}
		delete(r.plugins, id)
	}
}

// List returns the manifests of all registered plugins.
func (r *Registry) List() []*Manifest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	manifests := make([]*Manifest, 0, len(r.plugins))
	for _, entry := range r.plugins {
		manifests = append(manifests, entry.Manifest)
	}

	return manifests
}
