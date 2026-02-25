# Demonstration Client

This directory contains a simple Go-based demonstration client for the Inventory Reservation Service.

## Purpose

The client demonstrates the typical lifecycle of an order and inventory management:
1. **Admin Scope**: Initialize stock for products (e.g., Laptops and Mice).
2. **E-shop Scope**: A customer browses the shop, creates a reservation (cart hold), and later updates it.
3. **Checkout**: The customer submits an order, and the Order ID is attached to the reservation.
4. **CRM/Warehouse Scope**: An admin/warehouse worker receives the order and commits the shipment, which finalizes the inventory consumption.
5. **Edge Cases**: Demonstration of a failed reservation attempt due to insufficient stock.

## How to Run

1. Ensure the main service is running:
   ```bash
   make up
   ```

2. Start the demo server:
   ```bash
   make run-demo-ui
   ```

3. Open your browser at:
   [http://localhost:8081](http://localhost:8081)

## Key Concepts Demonstrated

- **Separation of Scopes**: Using Admin tokens for stock management and Public tokens for reservations.
- **Reservation Lifecycle**: `PENDING` -> `RESERVED` -> `COMMITTED`.
- **Atomic Operations**: Stock is reserved as soon as the customer adds to cart, preventing over-selling.
- **Order Attachment**: Linking an external Order ID to a reservation before final commitment.
