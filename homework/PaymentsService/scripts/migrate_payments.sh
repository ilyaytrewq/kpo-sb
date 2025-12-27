#!/usr/bin/env bash
set -euo pipefail

DB_URL="${PAYMENTS_DATABASE_URL:-postgres://postgres:postgres@localhost:5434/payments?sslmode=disable}"

psql "$DB_URL" -f services/payments-service/db/migrations/0001_init.up.sql
