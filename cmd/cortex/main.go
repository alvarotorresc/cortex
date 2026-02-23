package main

import (
	"log"
	"os"

	"github.com/alvarotorresc/cortex/internal/config"
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

	if err := server.Start(cfg); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
