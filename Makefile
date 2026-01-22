# === CONFIG ===

ifneq (,$(wildcard .env))
	include .env
	export
endif

MIGRATIONS_DIR = migrations
DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable
TEST_DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)_test?sslmode=disable


# === TEST MIGRATIONS ===
migrate-test-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(TEST_DB_URL)" up

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

down:
	docker compose down

restart:
	docker compose down && docker compose up --build

psql:
	docker compose exec db psql -U user -d timekeeper

update-sum:
	go mod tidy