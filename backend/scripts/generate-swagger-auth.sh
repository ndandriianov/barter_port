#!/usr/bin/env bash

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR/backend" || exit

swag init \
  -g main.go \
  -d ./cmd/auth,./internal/auth \
  -o ./docs/auth \
  --parseDependency \
  --parseInternal

npx -y openapi-to-postmanv2 -s docs/auth/swagger.json -o docs/auth/postman.json -p