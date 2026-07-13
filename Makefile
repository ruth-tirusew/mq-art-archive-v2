.PHONY: help up down logs postgres migrate api web admin env test test-integration arch-check web-install admin-install

# Use local Go when available; Docker image is fallback only (no GOTOOLCHAIN download).
GO := $(shell command -v go 2>/dev/null)
GO_DOCKER_IMAGE ?= golang:1.25-bookworm

help:
	@echo "mq — Ethiopian artists platform"
	@echo ""
	@echo "  make env       Copy .env.example to .env (if missing)"
	@echo "  make up        Start PostgreSQL"
	@echo "  make up-full   Start PostgreSQL + migrate + API (requires backend scaffold)"
	@echo "  make down      Stop all services"
	@echo "  make logs      Tail postgres logs"
	@echo "  make migrate   Run database migrations locally"
	@echo "  make test      Run unit tests"
	@echo "  make test-integration  Run integration tests (requires Docker)"
	@echo "  make arch-check  Verify hexagonal layer and bounded-context import rules"
	@echo "  make api       Run Go API locally (port 8080)"
	@echo "  make web       Run public SvelteKit app (port 5173)"
	@echo "  make admin     Run admin SvelteKit app (port 5174)"
ifneq ($(GO),)
	@echo ""
	@echo "  Go: $(GO) (local)"
else
	@echo ""
	@echo "  Go: not installed — test/migrate/arch-check use $(GO_DOCKER_IMAGE)"
endif

env:
	@test -f .env || cp .env.example .env
	@test -f apps/web/.env || cp apps/web/.env.example apps/web/.env
	@test -f apps/admin/.env || cp apps/admin/.env.example apps/admin/.env

web-install:
	cd apps/web && npm install

admin-install:
	cd apps/admin && npm install

web: env web-install
	cd apps/web && npm run dev -- --port 5173

admin: env admin-install
	cd apps/admin && npm run dev -- --port 5174

up: env
	docker compose up -d postgres

up-full: env
	docker compose --profile full up -d --build

down:
	docker compose --profile full down

logs:
	docker compose logs -f postgres

migrate: env
	@set -a && . ./.env && set +a && \
	if [ -n "$(GO)" ]; then \
		cd backend/api && GOTOOLCHAIN=local $(GO) run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$$DATABASE_URL" up; \
	else \
		docker run --rm --network host \
			-e "DATABASE_URL=$$DATABASE_URL" \
			-v "$(CURDIR)/backend/api:/app" -w /app \
			$(GO_DOCKER_IMAGE) sh -c 'go install github.com/pressly/goose/v3/cmd/goose@latest && goose -dir migrations postgres "$$DATABASE_URL" up'; \
	fi

api: env
	@if [ -n "$(GO)" ]; then \
		cd backend/api && GOTOOLCHAIN=local $(GO) run ./cmd/api; \
	else \
		echo "Go is not installed. Install Go or use: docker compose --profile full up"; \
		exit 1; \
	fi

test:
	@if [ -n "$(GO)" ]; then \
		cd backend/api && GOTOOLCHAIN=local $(GO) test ./...; \
	else \
		docker run --rm \
			-v "$(CURDIR)/backend/api:/app" -w /app \
			$(GO_DOCKER_IMAGE) go test ./...; \
	fi

test-integration:
	@if [ -n "$(GO)" ]; then \
		cd backend/api && GOTOOLCHAIN=local $(GO) test -tags=integration -count=1 ./...; \
	else \
		docker run --rm \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v "$(CURDIR)/backend/api:/app" -w /app \
			$(GO_DOCKER_IMAGE) go test -tags=integration -count=1 ./...; \
	fi

arch-check:
	@if [ -n "$(GO)" ]; then \
		cd backend/api && GOTOOLCHAIN=local $(GO) test ./internal/arch/...; \
	else \
		docker run --rm \
			-v "$(CURDIR)/backend/api:/app" -w /app \
			$(GO_DOCKER_IMAGE) go test ./internal/arch/...; \
	fi
