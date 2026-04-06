ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR/backend" || exit

npx -y @redocly/cli bundle \
  ./docs/doc-first/chats/swagger.yaml \
  --output ./tmp/chats.openapi.bundle.yaml

oapi-codegen \
  -generate types \
  -package types \
  -o ./contracts/openapi/chats/types/types.gen.go \
  ./tmp/chats.openapi.bundle.yaml