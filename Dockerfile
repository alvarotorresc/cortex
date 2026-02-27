# ============================================================================
# Stage 1: Build Go binaries (host + plugins)
# ============================================================================
FROM golang:1-alpine AS go-builder

RUN apk add --no-cache git

WORKDIR /src

# Cache Go module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy all Go source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY plugins/ plugins/
COPY proto/ proto/

# Build the host binary (static, no CGO — modernc.org/sqlite is pure Go)
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/cortex ./cmd/cortex

# Build plugin binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/plugins/finance-tracker/plugin ./plugins/finance-tracker/backend/
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/plugins/quick-notes/plugin ./plugins/quick-notes/backend/
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/plugins/project-hub/plugin ./plugins/project-hub/backend/

# ============================================================================
# Stage 2: Build Frontend (SvelteKit with adapter-static)
# ============================================================================
FROM node:22-alpine AS frontend-builder

WORKDIR /src/frontend

# Install pnpm
RUN npm install -g pnpm@10

# Cache dependency installation
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy frontend source and build
COPY frontend/ ./
RUN pnpm build

# ============================================================================
# Stage 3: Minimal runtime image
# ============================================================================
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates

# Create non-root user for security
RUN addgroup -S cortex && adduser -S cortex -G cortex

WORKDIR /app

# Copy host binary
COPY --from=go-builder /out/cortex /app/cortex

# Copy plugin binaries
COPY --from=go-builder /out/plugins/finance-tracker/plugin /plugins/finance-tracker/plugin
COPY --from=go-builder /out/plugins/quick-notes/plugin /plugins/quick-notes/plugin
COPY --from=go-builder /out/plugins/project-hub/plugin /plugins/project-hub/plugin

# Copy plugin manifests (migrations are embedded in plugin binaries via go:embed)
COPY plugins/finance-tracker/manifest.json /plugins/finance-tracker/manifest.json
COPY plugins/quick-notes/manifest.json /plugins/quick-notes/manifest.json
COPY plugins/project-hub/manifest.json /plugins/project-hub/manifest.json

# Copy frontend build output
COPY --from=frontend-builder /src/frontend/build /frontend

# Create data directory
RUN mkdir -p /data

# Set ownership
RUN chown -R cortex:cortex /app /plugins /frontend /data

# Runtime environment
ENV CORTEX_PORT=8080
ENV CORTEX_DATA_DIR=/data
ENV CORTEX_PLUGIN_DIR=/plugins
ENV CORTEX_FRONTEND_DIR=/frontend

EXPOSE 8080

VOLUME /data

USER cortex

CMD ["/app/cortex"]
