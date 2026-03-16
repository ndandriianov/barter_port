ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR/backend" || exit

npx -y @redocly/cli bundle \
  ./docs/doc-first/items/swagger.yaml \
  --output ./tmp/items.openapi.bundle.yaml

oapi-codegen \
  -generate types \
  -package types \
  -o ./internal/contracts/openapi/items/types/types.gen.go \
  ./tmp/items.openapi.bundle.yaml