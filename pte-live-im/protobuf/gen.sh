#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

mkdir -p protobuf/imapi

protoc \
  -I protobuf \
  --go_out=protobuf/imapi --go_opt=paths=source_relative \
  --go-grpc_out=protobuf/imapi --go-grpc_opt=paths=source_relative \
  protobuf/im_api.proto

echo "generated protobuf/imapi/*.go"
