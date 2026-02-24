# === CONFIG ===

ifneq (,$(wildcard .env))
	include .env
	export
endif

MIGRATIONS_DIR = migrations
DB_HOST_LOCAL = localhost
DB_SSL_MODE ?= disable
TEST_DB_PORT ?= 5433
TEST_DB_NAME ?= $(DB_NAME)_test

DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST_LOCAL):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)
TEST_DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST_LOCAL):$(TEST_DB_PORT)/$(TEST_DB_NAME)?sslmode=$(DB_SSL_MODE)


# === TEST MIGRATIONS ===
migrate-test-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(TEST_DB_URL)" up

migrate-test-down:
	-migrate -path $(MIGRATIONS_DIR) -database "$(TEST_DB_URL)" down -all

migrate-test-reset: migrate-test-down migrate-test-up

# === MIGRATIONS ===
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down

migrate-drop:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" drop -f

migrate-force:
	@echo "⚠️ Forcing to version: $(VERSION)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(VERSION)

migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

migrate-new:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

# === BUILD / RUN ===
build:
	docker compose build

up:
	docker compose up

test-db-up:
	docker compose up -d db_test

test-db-down:
	docker compose stop db_test

test-db-logs:
	docker compose logs db_test

down:
	docker compose down

restart:
	docker compose down && docker compose up --build

psql:
	docker compose exec db psql -U user -d timekeeper

update-sum:
	go mod tidy

test-unit:
	go test ./...

test-unit-no-cache:
	go test -count=1 ./...

test-coverage:
	go test ./... -cover

test-coverage-profile:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

test-coverage-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

test-integration: test-db-up
	@echo "Waiting for test database to be ready..."
	@until docker compose exec -T db_test pg_isready -U $(DB_USER) -d $(TEST_DB_NAME) > /dev/null 2>&1; do \
		sleep 1; \
	done
	$(MAKE) migrate-test-reset
	DB_HOST=$(DB_HOST_LOCAL) DB_PORT=$(TEST_DB_PORT) DB_NAME=$(TEST_DB_NAME) DB_SSL_MODE=$(DB_SSL_MODE) go test -tags=integration ./...

test-integration-no-cache: test-db-up
	@echo "Waiting for test database to be ready..."
	@until docker compose exec -T db_test pg_isready -U $(DB_USER) -d $(TEST_DB_NAME) > /dev/null 2>&1; do \
		sleep 1; \
	done
	$(MAKE) migrate-test-reset
	DB_HOST=$(DB_HOST_LOCAL) DB_PORT=$(TEST_DB_PORT) DB_NAME=$(TEST_DB_NAME) DB_SSL_MODE=$(DB_SSL_MODE) go test -count=1 -tags=integration ./...

test: test-db-up
	@echo "Waiting for test database to be ready..."
	@until docker compose exec -T db_test pg_isready -U $(DB_USER) -d $(TEST_DB_NAME) > /dev/null 2>&1; do \
		sleep 1; \
	done
	$(MAKE) migrate-test-reset
	DB_HOST=$(DB_HOST_LOCAL) DB_PORT=$(TEST_DB_PORT) DB_NAME=$(TEST_DB_NAME) DB_SSL_MODE=$(DB_SSL_MODE) go test ./...
	DB_HOST=$(DB_HOST_LOCAL) DB_PORT=$(TEST_DB_PORT) DB_NAME=$(TEST_DB_NAME) DB_SSL_MODE=$(DB_SSL_MODE) go test -tags=integration ./...
