# Go Inventory Reservation Service

Inventory reservation demo service built with Go, Gin, and PostgreSQL.

The service manages stock and reservation lifecycle for a typical checkout/cart flow:
- reserve items for a quote
- attach order
- commit shipment
- release or revert reservations
- auto-expire stale quotes
- archive old final-state reservations

## Why This Project Exists

This repository is a demonstration project focused on:
- clean layered architecture (`handler -> application -> service -> repository`)
- transactional consistency for stock movements
- optimistic locking for reservation updates
- pragmatic API error model and structured logging
- test automation with dedicated integration database

## Core Domain

### Stock
- `sku` (PK)
- `on_hand` (physical inventory)
- `reserved` (currently held by active reservations)
- `available = on_hand - reserved`

### Reservation
- `PENDING`: cart hold created, not yet submitted
- `RESERVED`: order attached
- `COMMITTED`: shipment finalized (inventory consumed)
- `RELEASED`: hold canceled, stock returned
- `REVERTED`: committed reservation reversed (refund/return)
- `EXPIRED`: pending hold timed out

### Reservation Item
- `(reservation_id, sku)`
- `qty`
- `is_active`

## Architecture

- `cmd/api`: application entrypoint
- `internal/kernel`: wiring dependencies, HTTP server, cron startup
- `internal/router`: route registration and middleware
- `internal/handler`: HTTP handlers, request validation, response mapping
- `internal/application`: orchestration / transaction scripts (use cases)
- `internal/service`: domain rules
- `internal/repository`: SQL access layer
- `internal/uow`: transaction boundaries (`WithUnitOfWork`)
- `internal/notifier`: quote expiration webhook client
- `migrations`: SQL schema migrations

## API Overview

### Health
- `GET /ping`

### Public (Bearer `PUBLIC_API_TOKEN`)
- `GET /stock/:sku`
- `GET /stock?limit=&offset=`
- `POST /reservation`
- `PUT /reservation`
- `GET /reservation/:id`
- `GET /reservation/quote/:quote_id`
- `GET /reservation/order/:order_id`
- `POST /reservation/attach-order`
- `POST /reservation/commit`
- `PATCH /reservation/:id/release`
- `POST /reservation/revert`

### Admin (Bearer `ADMIN_API_TOKEN`)
- `POST /admin/stock`
- `PUT /admin/stock`
- `DELETE /admin/stock/:sku`
- `GET /admin/stock/:sku/reservation-items?limit=&offset=`

Sample requests are in:
- `http/stock.http`
- `http/reservation.http`
- `http/health_check.http`

## Configuration

Main env vars (see `.env`):
- DB: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`
- Server: `APP_PORT`
- Auth: `ADMIN_API_TOKEN`, `PUBLIC_API_TOKEN`
- Expiration cron:
  - `QUOTE_EXPIRATION_MINUTES`
  - `QUOTE_EXPIRATION_CRON_SPEC`
  - `QUOTE_EXPIRATION_COUNT_LIMIT`
  - `QUOTE_EXPIRATION_NOTIFY_URL`
  - `QUOTE_EXPIRATION_NOTIFY_TIMEOUT_SECONDS`
- Archive cron:
  - `ARCHIVE_RESERVATIONS_AFTER_DAYS`
  - `ARCHIVE_RESERVATIONS_CRON_SPEC`
  - `ARCHIVE_RESERVATIONS_COUNT_LIMIT`
- Reservation item limit:
  - `RESERVATION_ITEM_MAX_QUANTITY` (defaults to 20 if unset)

## Local Run

### Start app and DB
```bash
make up
```

### Stop
```bash
make down
```

### Migrations
```bash
make migrate-up
make migrate-down
```

## Testing

### Unit tests
```bash
make test-unit
```

### Integration + full-flow tests (requires Docker)
Integration tests use a dedicated DB service:
- service: `db_test`
- port: `5433`
- db: `inventory_reservations_test`

Run everything:
```bash
make test
```

Or only integration-tagged tests:
```bash
make test-integration
```

### About integration tests
Integration/full-flow tests are tagged with `integration` build tag:
- `internal/repository/stock_repository_integration_test.go`
- `internal/application/flows_integration_test.go`
- `internal/application/process_expired_reservations_test.go`

This keeps `go test ./...` fast for unit-only runs.

## Implemented Test Coverage

- Unit tests:
  - hash function behavior (`internal/service/items_hash_test.go`)
  - pagination parameter normalization (`internal/model/api/pagination_test.go`)
  - stock service validation (`internal/service/stock_service_test.go`)
  - health route (`internal/router/app_health_routes_test.go`)
- Integration tests:
  - reset DB + migrate + create/read stock (`internal/repository/stock_repository_integration_test.go`)
  - full reservation lifecycle flows (`internal/application/flows_integration_test.go`)
  - optimistic-lock conflict handling on expiration (`internal/application/process_expired_reservations_test.go`)

Detailed testcase matrix: `docs/TEST_CASES.md`.

## Technical Notes

- Reservations use optimistic locking (`version`) to prevent lost updates.
- All multi-entity state transitions run inside DB transactions (`WithUnitOfWork`).
- Stock invariants are enforced both at DB level and service validation.
- Structured API error envelope includes optional request id.

## Code Review Notes

A review and remediation summary is available in:
- `docs/CODE_REVIEW.md`

## Demonstration Scope

This is intentionally a demonstration service. Production hardening that can be added:
- OpenAPI/Swagger contract
- idempotency keys for write endpoints
- stricter auth/token management (rotation, expiry, scopes)
- metrics and tracing
- larger test matrix for concurrent races and retry policies
