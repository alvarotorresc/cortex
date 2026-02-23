package main

import (
	"log"
	"os"

	"github.com/alvarotorresc/cortex/internal/config"
	"github.com/alvarotorresc/cortex/internal/db"
	pluginpkg "github.com/alvarotorresc/cortex/internal/plugin"
	"github.com/alvarotorresc/cortex/internal/server"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded: port=%d, data=%s, plugins=%s",
		cfg.Port, cfg.DataDir, cfg.PluginDir)

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize host database
	hostDB, err := db.NewHostDB(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize host database: %v", err)
	}
	defer hostDB.Close()

	// Initialize plugin system
	registry := pluginpkg.NewRegistry()
	loader := pluginpkg.NewLoader(cfg.PluginDir, cfg.DataDir, registry)

	// Load all plugins from the plugins directory
	if err := loader.LoadAll(); err != nil {
		log.Printf("Warning: error loading plugins: %v", err)
	}

	// Ensure plugins are unloaded on exit.
	// The server.Start function handles SIGINT/SIGTERM for HTTP shutdown.
	// We defer plugin cleanup so it runs after the server stops.
	defer func() {
		log.Println("Unloading plugins...")
		loader.UnloadAll()
	}()

	if err := server.Start(cfg, registry, hostDB); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
