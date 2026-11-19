# Database

Dhaba Orders uses PostgreSQL as its persistent relational database.

## Tables

- menu_items: available food items and their prices
- orders: customer orders and their lifecycle status
- order_items: individual items belonging to an order

## Order lifecycle

NEW -> COOKING -> READY

An order may also be marked CANCELLED.

## Initialization

The SQL initialization script is:
init/001_schema.sql

The PostgreSQL container will execute this script when initializing
a new, empty database data directory.
