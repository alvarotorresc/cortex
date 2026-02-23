# Cortex

## Descripcion
Self-hosted personal hub with a plugin system. Deploy on local network via Docker Compose, access from any device. All data stays local.

## Stack
- **Lenguaje:** Go
- **Framework:** chi (HTTP router) + HashiCorp go-plugin (plugin system) + gRPC
- **Base de datos:** SQLite (one per plugin, isolated)
- **Frontend:** SvelteKit (Svelte 5) + Tailwind CSS
- **Deploy:** Docker Compose (self-hosted, local network)

## Estructura del Proyecto
```
cortex/
├── cmd/cortex/          # Entry point
├── internal/
│   ├── config/          # Configuration loading and validation
│   ├── server/          # HTTP server, router, middleware
│   ├── plugin/          # Plugin system (registry, loader, gRPC)
│   └── db/              # Host database utilities
├── pkg/sdk/             # Public Plugin SDK for plugin authors
├── plugins/             # Installed plugin binaries (runtime)
├── proto/               # Protocol Buffers definitions
├── frontend/            # SvelteKit application
├── data/                # Runtime data (gitignored)
├── docs/                # Project documentation
└── .github/workflows/   # CI/CD
```

## Comandos
- **Build:** `make build`
- **Run:** `make run`
- **Test:** `make test`
- **Lint:** `make lint`
- **Format:** `make fmt`
- **Clean:** `make clean`

## Convenciones del Proyecto
- Go module: `github.com/alvarotorresc/cortex`
- Import order: stdlib, third-party, internal (`github.com/alvarotorresc/cortex/...`)
- Config via environment variables with `CORTEX_` prefix, validated at startup
- Plugin communication via gRPC (Protocol Buffers)
- Each plugin gets its own SQLite database in `data/plugins/{id}/`
- REST API responses follow `{ data, meta }` for success and `{ error: { code, message, details } }` for errors
- Frontend assets served by Go server in production, SvelteKit dev server in development

## Permisos de Claude

### Permitido (sin preguntar)
- Leer cualquier archivo del proyecto
- Ejecutar tests, linters, type checks
- Crear/editar archivos de codigo
- Buscar en la codebase
- Ejecutar builds locales
- Lectura de Git (status, log, diff, show)

### Permitido (pidiendo confirmacion)
- Instalar dependencias nuevas
- Modificar CI/CD
- Borrar archivos

### Prohibido
- Escritura de Git (commit, push, merge, etc.)
- Crear/comentar PRs e issues
- Modificar secrets/env en produccion
- Ejecutar migraciones en produccion
- Publicar packages
- Incluir binarios en el repo
