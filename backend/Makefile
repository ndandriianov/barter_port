swagger-auth:
	bash ./scripts/generate-swagger-auth.sh

infra-up:
	docker compose --profile infra up -d
	@echo "Waiting for Postgres to become ready..."
	@until docker compose exec -T db pg_isready -U postgres -d postgres >/dev/null 2>&1; do sleep 1; done
	$(MAKE) migrate-up

infra-down:
	docker compose --profile infra down

infra-down-v:
	docker compose --profile infra down -v

app-up:
	docker compose --profile app up -d --build

app-down:
	docker compose --profile app down

app-down-v:
	docker compose --profile app down -v

local-up:
	@trap 'kill 0' INT TERM EXIT; \
	CONFIG_COMMON=./config/common.yaml APP_ENV=local go run ./cmd/auth & \
	CONFIG_COMMON=./config/common.yaml APP_ENV=local CONFIG_SERVICE=./config/items.yaml go run ./cmd/items & \
	CONFIG_COMMON=./config/common.yaml APP_ENV=local CONFIG_SERVICE=./config/users.yaml go run ./cmd/users & \
	wait

create-migration:
	migrate create -ext sql -dir migrations/$(db) -seq $(name)

migrate-up:
	migrate -path migrations/auth -database "postgres://postgres:$${DB_PASSWORD:-postgres}@localhost:5432/auth_db?sslmode=disable" up
	migrate -path migrations/users -database "postgres://postgres:$${DB_PASSWORD:-postgres}@localhost:5432/users_db?sslmode=disable" up



PROTO_FILES := $(shell find contracts/grpc -name "*.proto")

grpcgen:
	protoc \
 	--go_out=. \
 	--go_opt=paths=source_relative \
 	--go-grpc_out=. \
 	--go-grpc_opt=paths=source_relative \
 	$(PROTO_FILES)
