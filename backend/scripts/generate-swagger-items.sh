#!/usr/bin/env bash

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR/backend" || exit

swag init \
  -g main.go \
  -d ./cmd/items,./internal/items \
  -o ./docs/items \
  --parseDependency \
  --parseInternal

npx -y openapi-to-postmanv2 -s docs/items/swagger.json -o docs/items/postman.json -p