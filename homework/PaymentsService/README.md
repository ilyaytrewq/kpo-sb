# PaymentsService (HW-4)

## Prerequisites
```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
npm install -g @redocly/cli
export PATH="$(go env GOPATH)/bin:$PATH"
```

## Code generation & lint
```bash
chmod +x scripts/check_api-files.sh
chmod +x scripts/generate_code.sh
chmod +x scripts/generate_sql.sh

./scripts/check_api-files.sh
./scripts/generate_code.sh
./scripts/generate_sql.sh
```

## Local infrastructure (Kafka + Postgres + Redis)
```bash
export HOST=localhost
docker compose up -d broker kafka-init kafka-ui orders-postgres payments-postgres redis
```

## Migrations
```bash
chmod +x scripts/migrate_orders.sh
chmod +x scripts/migrate_payments.sh

./scripts/migrate_orders.sh
./scripts/migrate_payments.sh
```

## Run services
```bash
docker compose up -d api-gateway orders-service payments-service
```

## Quick start cheat-sheet
```bash
./scripts/generate_code.sh
./scripts/generate_sql.sh
export HOST=localhost && docker compose up -d broker kafka-init kafka-ui orders-postgres payments-postgres
./scripts/migrate_orders.sh
./scripts/migrate_payments.sh
docker compose up -d api-gateway orders-service payments-service
```

## Kafka topics & groups
- Topics:
  - `payments.payment_requested.v1` (key = `order_id`)
  - `payments.payment_result.v1` (key = `order_id`)
- Consumer groups:
  - `payments-service` reads `payments.payment_requested.v1`
  - `orders-service` reads `payments.payment_result.v1`

Offsets are committed **only after** DB transaction commit. We use `FetchMessage` + `CommitMessages` (manual commit).

## Flow diagram
```
CreateOrder (API Gateway)
  → Orders gRPC
    → Orders outbox (PaymentRequested)
      → Kafka (payment_requested)
        → Payments consumer
          → Payments tx (inbox + account_ops + outbox PaymentResult)
            → Kafka (payment_result)
              → Orders consumer
                → Orders tx (inbox + status update)
```
