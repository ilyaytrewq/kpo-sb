# PaymentsService (HW-4)

## Overview
Payments and Orders microservices with an API Gateway, Kafka-based async processing,
and exactly-once guarantees via transactional outbox/inbox and idempotency keys.

## Stack
- Go (gRPC, chi)
- PostgreSQL (orders/payments databases)
- Apache Kafka
- Redis (read cache)
- OpenAPI/Swagger (API Gateway + Swagger UI)
- Docker / Docker Compose
- sqlc, oapi-codegen, Protobuf
- React + Vite (frontend)

## Public endpoints (host 158.160.211.103)
- API Gateway (HTTP): http://158.160.211.103:8080/api/v1
- Swagger UI: http://158.160.211.103:8088
- Kafka UI: http://158.160.211.103:8085
- Frontend: http://158.160.211.103:3000
- Kafka broker (external): 158.160.211.103:29092

## HTTP API (short)
Base path: `/api/v1`

- `POST /payments/account` - create account (idempotent)
- `POST /payments/account/topup` - top up account (idempotent)
- `GET /payments/account/balance` - get balance (requires `X-User-Id`)
- `POST /orders` - create order (starts async payment)
- `GET /orders` - list orders
- `GET /orders/{orderId}` - get order details

Headers:
- `Idempotency-Key` is required for all POST requests.
- `X-User-Id` is optional for most endpoints (gateway generates it if missing).

## Kafka
Topics:
- `payments.payment_requested.v1` (key = `order_id`)
- `payments.payment_result.v1` (key = `order_id`)

Consumer groups:
- `payments-service` reads `payments.payment_requested.v1`
- `orders-service` reads `payments.payment_result.v1`

Offsets are committed only after DB transaction commit (`FetchMessage` + `CommitMessages`).

## Exactly-once and consistency
- Transactional outbox/inbox in Orders and Payments.
- Atomic balance update + `account_ops` prevents double charge per order.
- Idempotency keys on HTTP POST and top-up operations.

## Flow diagram
```
CreateOrder (API Gateway)
  -> Orders gRPC
    -> Orders outbox (PaymentRequested)
      -> Kafka (payment_requested)
        -> Payments consumer
          -> Payments tx (inbox + account_ops + outbox PaymentResult)
            -> Kafka (payment_result)
              -> Orders consumer
                -> Orders tx (inbox + status update)
```

## Local development

### Prerequisites
```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
npm install -g @redocly/cli
export PATH="$(go env GOPATH)/bin:$PATH"
```

### Code generation & lint
```bash
chmod +x scripts/check_api-files.sh
chmod +x scripts/generate_code.sh
chmod +x scripts/generate_sql.sh

./scripts/check_api-files.sh
./scripts/generate_code.sh
./scripts/generate_sql.sh
```

### Local infrastructure (Kafka + Postgres + Redis)
```bash
export HOST=localhost
docker compose up -d broker kafka-init kafka-ui orders-postgres payments-postgres redis
```

### Migrations
```bash
chmod +x scripts/migrate_orders.sh
chmod +x scripts/migrate_payments.sh

./scripts/migrate_orders.sh
./scripts/migrate_payments.sh
```

### Run services
```bash
docker compose up -d api-gateway orders-service payments-service swagger-ui frontend
```

### Quick start cheat-sheet
```bash
./scripts/generate_code.sh
./scripts/generate_sql.sh
export HOST=localhost && docker compose up -d broker kafka-init kafka-ui orders-postgres payments-postgres redis
./scripts/migrate_orders.sh
./scripts/migrate_payments.sh
docker compose up -d api-gateway orders-service payments-service swagger-ui frontend
```
