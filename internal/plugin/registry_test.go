package plugin_test

import (
	"testing"

	"github.com/alvarotorresc/cortex/internal/plugin"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := plugin.NewRegistry()

	manifest := &plugin.Manifest{
		ID:      "test-plugin",
		Name:    "Test Plugin",
		Version: "0.1.0",
	}

	registry.Register("test-plugin", nil, manifest)

	entry, ok := registry.Get("test-plugin")
	if !ok {
		t.Fatal("expected plugin to be registered")
	}
	if entry.Manifest.Name != "Test Plugin" {
		t.Errorf("expected name 'Test Plugin', got '%s'", entry.Manifest.Name)
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := plugin.NewRegistry()

	manifest := &plugin.Manifest{ID: "test-plugin", Name: "Test"}
	registry.Register("test-plugin", nil, manifest)
	registry.Unregister("test-plugin")

	_, ok := registry.Get("test-plugin")
	if ok {
		t.Fatal("expected plugin to be unregistered")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := plugin.NewRegistry()

	registry.Register("a", nil, &plugin.Manifest{ID: "a", Name: "A"})
	registry.Register("b", nil, &plugin.Manifest{ID: "b", Name: "B"})

	list := registry.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(list))
	}
}
