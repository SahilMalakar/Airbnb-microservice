include .env
export

# --- Migrations ---
migrate-up:
	goose -dir db/migration postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir db/migration postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir db/migration postgres "$(DATABASE_URL)" status

# --- Run ---
run:
	go run main.go

# --- Build ---
build:
	go build -o bin/api_gateway_service main.go

.PHONY: migrate-up migrate-down migrate-status run build
