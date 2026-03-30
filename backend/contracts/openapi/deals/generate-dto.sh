ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR/backend" || exit

npx -y @redocly/cli bundle \
  ./docs/doc-first/deals/swagger.yaml \
  --output ./tmp/deals.openapi.bundle.yaml

oapi-codegen \
  -generate types \
  -package types \
  -o ./contracts/openapi/deals/types/types.gen.go \
  ./tmp/deals.openapi.bundle.yaml