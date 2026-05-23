.DEFAULT_GOAL := help

SHELL := /bin/bash

GO_API_DIR := services/api
GO_ADMIN_DIR := apps/admin
MINIAPP_DIR := apps/miniapp

GOOSE := goose
SQLC := sqlc
# Dùng `go run` để không bắt buộc cài CLI riêng (cài 1 lần thì local hơn).
TEMPL := go run github.com/a-h/templ/cmd/templ@latest

.PHONY: help
help: ## Hiển thị danh sách target
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ─── Lint & Test ─────────────────────────────────────────────────────────────

.PHONY: lint
lint: lint-go lint-ts ## Lint tất cả (Go + TS)

.PHONY: lint-go
lint-go: ## Lint Go (golangci-lint)
	cd $(GO_API_DIR) && golangci-lint run ./... || true
	cd $(GO_ADMIN_DIR) && golangci-lint run ./... || true

.PHONY: lint-ts
lint-ts: ## Lint TypeScript (eslint)
	npm run lint --workspaces --if-present

.PHONY: test
test: test-go test-ts ## Test tất cả

.PHONY: test-go
test-go: ## Test Go (services/api + apps/admin)
	cd $(GO_API_DIR) && go test ./... -race -coverprofile=coverage.out
	cd $(GO_ADMIN_DIR) && go test ./... -race -coverprofile=coverage.out

.PHONY: test-ts
test-ts: ## Test TypeScript
	npm run test --workspaces --if-present

# ─── Build ───────────────────────────────────────────────────────────────────

.PHONY: build
build: build-go build-ts ## Build tất cả

.PHONY: build-go
build-go: templ ## Build Go binaries (auto-regen templ trước)
	mkdir -p $(GO_API_DIR)/bin $(GO_ADMIN_DIR)/bin
	cd $(GO_API_DIR) && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o bin/api ./cmd/api
	cd $(GO_API_DIR) && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o bin/worker ./cmd/worker
	cd $(GO_ADMIN_DIR) && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o bin/admin ./cmd/admin

.PHONY: build-ts
build-ts: ## Build miniapp dist/
	npm run build --workspaces --if-present

# ─── Run dev ─────────────────────────────────────────────────────────────────

.PHONY: api
api: ## Chạy public API (port 8080)
	cd $(GO_API_DIR) && go run ./cmd/api

.PHONY: worker
worker: ## Chạy background worker
	cd $(GO_API_DIR) && go run ./cmd/worker

.PHONY: admin
admin: templ ## Chạy admin web (port 8081)
	cd $(GO_ADMIN_DIR) && go run ./cmd/admin

.PHONY: miniapp
miniapp: ## Chạy ZMP miniapp dev (port 5173)
	cd $(MINIAPP_DIR) && npm run dev

# ─── DB / codegen ────────────────────────────────────────────────────────────

.PHONY: db-up
db-up: ## Chạy migration lên latest
	cd $(GO_API_DIR) && $(GOOSE) -dir migrations postgres "$$DATABASE_URL" up

.PHONY: db-down
db-down: ## Rollback 1 migration
	cd $(GO_API_DIR) && $(GOOSE) -dir migrations postgres "$$DATABASE_URL" down

.PHONY: db-status
db-status: ## Trạng thái migration
	cd $(GO_API_DIR) && $(GOOSE) -dir migrations postgres "$$DATABASE_URL" status

.PHONY: sqlc
sqlc: ## Sinh Go code từ SQL queries
	cd $(GO_API_DIR) && $(SQLC) generate

.PHONY: templ
templ: ## Sinh Go code từ admin templates (chạy trước go build/run)
	cd $(GO_ADMIN_DIR) && $(TEMPL) generate -path internal/web

.PHONY: openapi
openapi: ## Sinh TS SDK từ openapi.yaml
	npx --yes @openapitools/openapi-generator-cli generate \
		-i $(GO_API_DIR)/openapi.yaml \
		-g typescript-fetch \
		-o packages/api-client/src/generated

# ─── Infra local ─────────────────────────────────────────────────────────────

.PHONY: infra-up
infra-up: ## Chạy Postgres + Redis + MinIO local
	docker compose -f deploy/docker-compose.yml up -d

.PHONY: infra-down
infra-down: ## Tắt infra local
	docker compose -f deploy/docker-compose.yml down

.PHONY: infra-logs
infra-logs: ## Xem log infra
	docker compose -f deploy/docker-compose.yml logs -f

# ─── Vệ sinh ─────────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Xóa build artifacts
	rm -rf $(GO_API_DIR)/bin $(GO_ADMIN_DIR)/bin
	rm -rf $(MINIAPP_DIR)/dist
	rm -rf node_modules
	find . -name 'coverage.out' -delete

.PHONY: fmt
fmt: ## Format code (Go + TS)
	cd $(GO_API_DIR) && gofmt -w .
	cd $(GO_ADMIN_DIR) && gofmt -w .
	npm run format
