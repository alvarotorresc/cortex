package plugin_test

import (
	"fmt"
	"sync"
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

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := plugin.NewRegistry()
	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Half the goroutines register plugins concurrently.
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			id := fmt.Sprintf("plugin-%d", index)
			registry.Register(id, nil, &plugin.Manifest{ID: id, Name: id})
		}(i)
	}

	// The other half read from the registry concurrently.
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			id := fmt.Sprintf("plugin-%d", index)
			registry.Get(id)
			registry.List()
		}(i)
	}

	wg.Wait()

	// After all goroutines finish, every plugin should be registered.
	list := registry.List()
	if len(list) != goroutines {
		t.Errorf("expected %d plugins, got %d", goroutines, len(list))
	}
}

func TestRegistry_UnregisterNonexistent(t *testing.T) {
	registry := plugin.NewRegistry()

	// Should not panic when unregistering a plugin that was never registered.
	registry.Unregister("nonexistent-plugin")

	list := registry.List()
	if len(list) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(list))
	}
}

func TestRegistry_OverwriteExisting(t *testing.T) {
	registry := plugin.NewRegistry()

	registry.Register("my-plugin", nil, &plugin.Manifest{ID: "my-plugin", Name: "Original"})
	registry.Register("my-plugin", nil, &plugin.Manifest{ID: "my-plugin", Name: "Overwritten"})

	entry, ok := registry.Get("my-plugin")
	if !ok {
		t.Fatal("expected plugin to be registered")
	}
	if entry.Manifest.Name != "Overwritten" {
		t.Errorf("expected name 'Overwritten', got '%s'", entry.Manifest.Name)
	}

	// Only one entry should exist, not two.
	list := registry.List()
	if len(list) != 1 {
		t.Errorf("expected 1 plugin after overwrite, got %d", len(list))
	}
}

func TestRegistry_GetNonexistent(t *testing.T) {
	registry := plugin.NewRegistry()

	_, ok := registry.Get("does-not-exist")
	if ok {
		t.Fatal("expected ok to be false for nonexistent plugin")
	}
}

func TestRegistry_ListEmpty(t *testing.T) {
	registry := plugin.NewRegistry()

	list := registry.List()
	if list == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(list) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(list))
	}
}
