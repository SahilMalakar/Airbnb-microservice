include .env
export
# Strip surrounding quotes from values loaded via `include .env`,
# since Make (unlike godotenv) exports them literally, quotes and all.
# Without this, DATABASE_URL leaks into the Go process as a malformed,
# quote-wrapped string, and pgx silently falls back to port 5432.
DATABASE_URL := $(shell echo $(DATABASE_URL) | sed -e 's/^"//' -e 's/"$$//')
PORT := $(shell echo $(PORT) | sed -e 's/^"//' -e 's/"$$//')

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

# --- Docker ---
docker-up:
	docker compose -f docker/docker-compose.dev.yml up -d

docker-down:
	docker compose -f docker/docker-compose.dev.yml down

docker-logs:
	docker compose -f docker/docker-compose.dev.yml logs -f

.PHONY: migrate-up migrate-down migrate-status run build docker-up docker-down docker-logs