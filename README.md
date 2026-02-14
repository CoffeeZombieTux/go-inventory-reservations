## Reservation Flow

### Create Reservation (Active Cart)
**Endpoint:**
`POST http://localhost:8080/reservation`

Creates a new reservation with these properties:
- Status is set to `PENDING`.
- `expired_at` is set to the current time plus `.ENV.QUOTE_EXPIRATION_MINUTES` (in minutes).
- Products are reserved in stock for the specified duration.

---

### Update Reservation (Active Cart)
**Endpoint:**
`PUT http://localhost:8080/reservation`

Updates an existing reservation:
- `quote_id` and items can be modified.
- Resets the `expired_at` value.
- Status is set to `PENDING` (useful if the reservation was previously expired).

**Allowed statuses:** `PENDING`, `EXPIRED`

*Note: Updating always resets the expiration time.*

---

### Attach Order (Submit Order)
**Endpoint:**
`POST http://localhost:8080/reservation/attach-order`

Links an `order_id` to a reservation:
- Removes the `expired_at` field.
- Status is updated to `RESERVED`.

*Use this when a customer submits an order and stock is reserved for shipment.*

**Allowed status:** `PENDING`

---

### Commit Reservation
**Endpoint:**
`POST http://localhost:8080/reservation/commit`

Marks the reservation as completed, typically after products are shipped:
- Status is updated to `COMMITTED`.

**Allowed status:** `RESERVED`

---

### Release Reservation
**Endpoint:**
`GET http://localhost:8080/reservation/:id/release`

Cancels the reservation:
- Status is set to `RELEASED`.
- Removes the `expired_at` field.

**Allowed statuses:** `RESERVED`, `PENDING`, `EXPIRED`

---

### Revert Reservation
**Endpoint:**
`POST http://localhost:8080/reservation/revert`

Reverts a committed reservation, usually for refunds:
- Status is set to `REVERTED`.

**Allowed status:** `COMMITTED`

---

### Reservation Expiration Cron
Periodically checks reservations where `expired_at` is less than or equal to the current time.
- Sets status to `EXPIRED`.

**Configuration:**
- `.ENV.QUOTE_EXPIRATION_CRON_SPEC`: Cron schedule.
- `.ENV.QUOTE_EXPIRATION_COUNT_LIMIT`: Maximum records processed per run.

---

### Delete Old Reservations Cron
Removes all reservations in final statuses that have an `updated_at` older than the threshold (default: one month).

**Configuration:**
- `.ENV.ARCHIVE_RESERVATIONS_AFTER_DAYS`: Number of days before archiving.
- `.ENV.ARCHIVE_RESERVATIONS_CRON_SPEC`: Cron schedule.
- `.ENV.ARCHIVE_RESERVATIONS_COUNT_LIMIT`: Maximum records processed per run.

---

### Final Statuses
Reservations are considered final if marked as one of the following:
- `COMMITTED`
- `RELEASED`
- `REVERTED`
- `EXPIRED`

---

**Notes:**
- All time calculations use the server’s system time.
- Set environment variable names according to your deployment configuration.