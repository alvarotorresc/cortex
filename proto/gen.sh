#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PROTO_DIR="$PROJECT_ROOT/proto"
OUT_DIR="$PROJECT_ROOT/internal/plugin/proto"
PROTOC="${PROTOC:-protoc}"
GOBIN="${GOBIN:-$(go env GOPATH)/bin}"

mkdir -p "$OUT_DIR"

export PATH="$GOBIN:$PATH"

"$PROTOC" \
  --proto_path="$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$OUT_DIR" \
  --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR/plugin.proto"

echo "Proto files generated in $OUT_DIR"
