swagger-auth:
	bash ./scripts/generate-swagger-auth.sh

infra-up:
	docker compose --profile infra up -d
	@echo "Waiting for Postgres to become ready..."
	@until docker compose exec -T db pg_isready -U postgres -d postgres >/dev/null 2>&1; do sleep 1; done
	$(MAKE) migrate-up

infra-up-wo-migrate:
	docker compose --profile infra up -d

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
	CONFIG_COMMON=./config/common.yaml APP_ENV=local CONFIG_SERVICE=./config/deals.yaml go run ./cmd/deals & \
	CONFIG_COMMON=./config/common.yaml APP_ENV=local CONFIG_SERVICE=./config/users.yaml go run ./cmd/users & \
	CONFIG_COMMON=./config/common.yaml APP_ENV=local CONFIG_SERVICE=./config/chats.yaml go run ./cmd/chats & \
	wait

seed-demo:
	go run ./cmd/seed-demo

create-migration:
	migrate create -ext sql -dir migrations/$(db) -seq $(name)

db_url = postgres://postgres:$${DB_PASSWORD:-postgres}@localhost:5432/$(1)?sslmode=disable

define reset_and_migrate
	@echo "Resetting $(1)..."
	@sh -c 'set -e; \
	out=$$(migrate -path migrations/$(2) -database "$(call db_url,$(1))" down 2>&1) || { \
		case "$$out" in \
			*"no change"*) ;; \
			*) printf "%s\n" "$$out"; exit 1 ;; \
		esac; \
	}; \
	migrate -path migrations/$(2) -database "$(call db_url,$(1))" up'
endef

migrate-up:
	migrate -path migrations/auth -database "postgres://postgres:$${DB_PASSWORD:-postgres}@localhost:5432/auth_db?sslmode=disable" up
	migrate -path migrations/users -database "postgres://postgres:$${DB_PASSWORD:-postgres}@localhost:5432/users_db?sslmode=disable" up
	migrate -path migrations/deals -database "postgres://postgres:$${DB_PASSWORD:-postgres}@localhost:5432/deals_db?sslmode=disable" up
	migrate -path migrations/chats -database "postgres://postgres:$${DB_PASSWORD:-postgres}@localhost:5432/chats_db?sslmode=disable" up

migrate-reset-all:
	$(call reset_and_migrate,auth_db,auth)
	$(call reset_and_migrate,users_db,users)
	$(call reset_and_migrate,deals_db,deals)
	$(call reset_and_migrate,chats_db,chats)

reseed-demo: migrate-reset-all seed-demo



PROTO_FILES := $(shell find contracts/grpc -name "*.proto")

grpcgen:
	protoc \
 	--go_out=. \
 	--go_opt=paths=source_relative \
 	--go-grpc_out=. \
 	--go-grpc_opt=paths=source_relative \
 	$(PROTO_FILES)
