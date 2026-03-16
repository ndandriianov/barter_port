ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR/backend" || exit

docker run --rm -v $PWD:/work -w /work caddy:2 caddy fmt --overwrite Caddyfile