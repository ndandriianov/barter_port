ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR/backend" || exit

npx -y @redocly/cli bundle \
  ./docs/doc-first/users/swagger.yaml \
  --output ./tmp/users.openapi.bundle.yaml

oapi-codegen \
  -generate types \
  -package types \
  -o ./internal/contracts/openapi/users/types/types.gen.go \
  ./tmp/users.openapi.bundle.yaml