# Cortex

> Self-hosted personal hub with a plugin system.

[![CI](https://github.com/alvarotorresc/cortex/actions/workflows/ci.yml/badge.svg)](https://github.com/alvarotorresc/cortex/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

## What is Cortex

Cortex is a self-hosted, offline-first personal hub. Deploy it on your local network via Docker Compose and access it from any device (phone, laptop, tablet). All data stays local -- zero cloud dependencies.

The core value is the **plugin system**: a platform where modular tools (finance tracker, notes, habit tracker, bookmarks, etc.) can be installed, uninstalled, and created by anyone.

## Tech Stack

- **Backend:** Go (chi HTTP router)
- **Plugin system:** HashiCorp go-plugin + gRPC
- **Frontend:** SvelteKit (Svelte 5) + Tailwind CSS
- **Database:** SQLite (one per plugin, isolated)
- **Deployment:** Docker Compose

## Local Development

### Prerequisites

- Go >= 1.23
- Node.js >= 20 + pnpm (for frontend)
- golangci-lint
- lefthook
- Protocol Buffers compiler (protoc) -- for plugin development

### Installation

```bash
git clone https://github.com/alvarotorresc/cortex.git
cd cortex
go mod download
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CORTEX_PORT` | HTTP server port | `8080` |
| `CORTEX_DATA_DIR` | Runtime data directory | `./data` |
| `CORTEX_PLUGIN_DIR` | Plugin binaries directory | `./plugins` |
| `CORTEX_FRONTEND_DIR` | Frontend build directory | `./frontend/build` |

### Available Commands

| Command | Description |
|---------|-------------|
| `make build` | Compile the Cortex binary |
| `make run` | Build and run the server |
| `make test` | Run all tests with race detection |
| `make lint` | Run golangci-lint |
| `make fmt` | Format all Go source files |
| `make clean` | Remove build artifacts |

## Architecture

```
Docker Compose
├── host (Go server, port 8080)
│   ├── Plugin Registry       -- register/unregister plugins
│   ├── Plugin Loader         -- launch subprocesses via go-plugin
│   ├── gRPC Client Manager   -- communicate with each plugin
│   ├── Reverse Proxy         -- /api/plugins/{id}/* -> gRPC
│   ├── SQLite per plugin     -- data/plugins/{id}/db.sqlite
│   └── Asset Server          -- /plugins/{id}/assets/*
│
├── frontend (SvelteKit, served by Go in production)
│   ├── Shell                 -- sidebar, topbar, dark mode, i18n
│   ├── Dashboard             -- grid of plugin widgets
│   └── Plugin pages          -- /plugins/{id} -> load plugin UI
│
└── plugins/                  -- installed plugin binaries + assets
```

## License

MIT -- see [LICENSE](./LICENSE)
