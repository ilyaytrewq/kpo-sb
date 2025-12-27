#!/usr/bin/env bash
set -euo pipefail

DB_URL="${ORDERS_DATABASE_URL:-postgres://postgres:postgres@localhost:5433/orders?sslmode=disable}"

psql "$DB_URL" -f services/orders-service/db/migrations/0001_init.up.sql
