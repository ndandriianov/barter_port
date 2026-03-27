swagger-auth:
	bash ./scripts/generate-swagger-auth.sh

infra-up:
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
	CONFIG_COMMON=./config/common.yaml APP_ENV=local CONFIG_SERVICE=./config/items.yaml go run ./cmd/items & \
	CONFIG_COMMON=./config/common.yaml APP_ENV=local CONFIG_SERVICE=./config/users.yaml go run ./cmd/users & \
	wait

create-migration:
	migrate create -ext sql -dir migrations/$(db) -seq $(name)
