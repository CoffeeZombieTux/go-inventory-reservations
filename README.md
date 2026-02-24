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
- Server: `SERVER_HOST`, `SERVER_PORT`
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

### `.env` example for local development

```env
# App server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Main DB (used by app + migrations)
DB_HOST=localhost
DB_PORT=5432
DB_USER=root
DB_PASSWORD=RooT!@123
DB_NAME=inventory_reservations
DB_SSL_MODE=disable

# API auth tokens
ADMIN_API_TOKEN=admin-dev-token
PUBLIC_API_TOKEN=public-dev-token

# Logging
LOG_LEVEL=info
LOG_FORMAT=text

# Quote expiration flow
QUOTE_EXPIRATION_MINUTES=15
QUOTE_EXPIRATION_CRON_SPEC=@every 1m
QUOTE_EXPIRATION_COUNT_LIMIT=1000
QUOTE_EXPIRATION_NOTIFY_URL=
QUOTE_EXPIRATION_NOTIFY_TIMEOUT_SECONDS=5

# Archive flow
ARCHIVE_RESERVATIONS_AFTER_DAYS=30
ARCHIVE_RESERVATIONS_CRON_SPEC=0 4 * * *
ARCHIVE_RESERVATIONS_COUNT_LIMIT=10000

# Reservation item constraints
RESERVATION_ITEM_MAX_QUANTITY=20
```

### Integration test DB notes

`make test` / `make test-integration` use a dedicated DB on port `5433` and database `${DB_NAME}_test`.
With the example above this resolves to:
- `inventory_reservations_test`
- `localhost:5433`

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
