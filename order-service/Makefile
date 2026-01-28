include .env
export

.PHONY: tools
tools:
	@echo "Syncing tool dependencies..."
	@go mod download
	@echo "✓ Tools synced (run via 'go run')"

.PHONY: setup
setup: tools
	@echo "Setting up project..."
	@go mod tidy
	@git config core.hooksPath .githooks
	@echo "✓ Setup complete!"

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go run golang.org/x/tools/cmd/goimports -local api-refactor -w .
	@go fmt ./...
	@echo "✓ Formatted"

.PHONY: lint
lint:
	@echo "Running linter..."
	@go run -modfile=go.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint run ./internal/... ./cmd/...
	@echo "✓ Lint complete"

.PHONY: test
test:
	@echo "Running all tests..."
	@go run gotest.tools/gotestsum@latest --format testname -- ./...

.PHONY: run
run:
	@go run cmd/api/main.go

.PHONY: build
build:
	@echo "Building..."
	@go build -o bin/api-refactor cmd/api/main.go
	@echo "✓ Built to bin/api-refactor"

.PHONY: dev
dev:
	@echo "Starting dev server with Air..."
	@direnv exec . sh -c 'go run ./cmd/api'

# Compose DATABASE_URL from separate variables
DATABASE_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=require

.PHONY: migrate-up
migrate-up:
	@if [ -z "$(DB_HOST)" ] || [ -z "$(DB_USER)" ] || [ -z "$(DB_NAME)" ]; then \
		echo "Error: Database variables not set. Make sure direnv is loaded or set DB_HOST, DB_USER, DB_PASSWORD, DB_NAME in .env"; \
		exit 1; \
	fi
	@echo "Running migrations up..."
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path db/migrations -database "$(DATABASE_URL)" up
	@echo "✓ Migrations applied"

.PHONY: migrate-down
migrate-down:
	@if [ -z "$(DB_HOST)" ] || [ -z "$(DB_USER)" ] || [ -z "$(DB_NAME)" ]; then \
		echo "Error: Database variables not set. Make sure direnv is loaded or set DB_HOST, DB_USER, DB_PASSWORD, DB_NAME in .env"; \
		exit 1; \
	fi
	@echo "Rolling back migrations..."
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path db/migrations -database "$(DATABASE_URL)" down 1
	@echo "✓ Rolled back 1 migration"

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)..."
	@go run github.com/golang-migrate/migrate/v4/cmd/migrate create -ext sql -dir db/migrations -seq $(NAME)
	@echo "✓ Migration files created in db/migrations/"

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(DB_HOST)" ] || [ -z "$(DB_USER)" ] || [ -z "$(DB_NAME)" ]; then \
		echo "Error: Database variables not set. Make sure direnv is loaded or set DB_HOST, DB_USER, DB_PASSWORD, DB_NAME in .env"; \
		exit 1; \
	fi
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make migrate-force VERSION=123"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(VERSION)..."
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path db/migrations -database "$(DATABASE_URL)" force $(VERSION)
	@echo "✓ Migration version forced to $(VERSION)"

.PHONY: sqlc
sqlc:
	@echo "Generating sqlc code..."
	@cd db && go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
	@echo "✓ sqlc generated"
