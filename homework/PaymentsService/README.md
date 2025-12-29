# PaymentsService (HW-4)
Микросервисная система **Orders + Payments**: создание заказов и управление счетом пользователя.
Оплата запускается **асинхронно через Kafka**, а консистентность обеспечивается через **Transactional Outbox/Inbox** и обязательный **Idempotency-Key** для безопасных повторов запросов.

Базовый URL: `http://158.160.219.201:8080/api/v1`  \
Swagger UI: `http://158.160.219.201:8088`  \
Kafka UI: `http://158.160.219.201:8085`  \
Frontend (демо): `http://158.160.219.201:3000`

---

## Архитектура

### Компоненты

**Бизнес-сервисы:**

1. **api-gateway** (`:8080`) — публичный HTTP API (OpenAPI), проксирует запросы в **Orders** и **Payments** по gRPC.
2. **orders-service** (`:9001`) — хранит заказы в Postgres, публикует событие `PaymentRequested` через outbox, читает `PaymentResult` и обновляет статус заказа.
3. **payments-service** (`:9002`) — хранит счета в Postgres, читает `PaymentRequested`, выполняет списание атомарно и пишет `PaymentResult` через outbox.
4. **frontend** (`:3000`) — небольшой UI для ручного прогона сценария.

**Инфраструктура:** Kafka брокер + Kafka UI, Redis (read-cache), два Postgres (orders/payments), Swagger UI.

### Пайплайн обработки заказа

```
Client
  → API Gateway (HTTP)
    → Orders Service (gRPC)
      → Orders DB (Order=NEW + Outbox: PaymentRequested)
        → Kafka topic: payments.payment_requested.v1
          → Payments consumer
            → Payments DB (Inbox + атомарное списание + Outbox: PaymentResult)
              → Kafka topic: payments.payment_result.v1
                → Orders consumer
                  → Orders DB (Inbox + Order status: FINISHED/CANCELLED)
```

---

## Запуск

### 1) Поднять всё окружение

```bash
docker compose up --build -d
```

### 2) (Опционально) Создать Kafka-топики

Если топики не создались автоматически, можно выполнить:

```bash
bash scripts/create_topics.sh
```

После запуска:
- API Gateway: `http://158.160.219.201:8080/api/v1`
- Swagger UI: `http://158.160.219.201:8088`
- Kafka UI: `http://158.160.219.201:8085`
- Frontend: `http://158.160.219.201:3000`

---

## ⚙️ Асинхронная обработка и consistency

- `POST /orders` создаёт заказ со статусом **NEW** и **не ждёт** результата оплаты.
- Итоговый статус заказа становится **FINISHED** или **CANCELLED** после обработки цепочки событий.

### Kafka

Топики:
- `payments.payment_requested.v1` — запрос на оплату (key = `order_id`)
- `payments.payment_result.v1` — результат оплаты (key = `order_id`)

Группы потребителей:
- `payments-service` читает `payments.payment_requested.v1`
- `orders-service` читает `payments.payment_result.v1`

Offsets коммитятся **только после** успешного завершения DB-транзакции (ручной commit).

## 🛠 Tech Stack

- **Go 1.25+** — backend
- **gRPC** — синхронные вызовы между gateway ↔ services
- **Chi + OpenAPI 3.0** — HTTP слой API Gateway
- **PostgreSQL** — две БД (orders/payments)
- **Apache Kafka** — асинхронные события
- **Redis** — read-cache (баланс/заказы)
- **sqlc** — типобезопасный слой запросов к БД
- **buf + Protobuf** — контракты gRPC / события
- **oapi-codegen** — генерация сервера/типов для API Gateway
- **Docker / Docker Compose** — запуск окружения
- **React + Vite** — frontend

---

## 🔌 API Endpoints

Base path: `/api/v1`

### Payments
- `POST /payments/account` — создать счёт (макс. 1 на пользователя)
- `POST /payments/account/topup` — пополнить счёт
- `GET /payments/account/balance` — получить баланс (**требует `X-User-Id`**)

### Orders
- `POST /orders` — создать заказ (оплата стартует асинхронно)
- `GET /orders` — список заказов пользователя
- `GET /orders/{orderId}` — детали / статус заказа

### Важные заголовки
- `Idempotency-Key: <string>` — **обязателен для всех POST**
- `X-User-Id: <string>` — опционален (gateway может сгенерировать), **обязателен** для `GET /payments/account/balance`

---

## 📁 Project Structure

```
.
├── api-files/
│   └── openapi/
│       └── api-gateway.yaml          # OpenAPI спецификация HTTP API
├── proto/                            # Protobuf контракты (gRPC + events)
├── gen/                              # Сгенерированный код (buf + oapi-codegen)
├── services/
│   ├── api-gateway/                  # HTTP API + gRPC clients
│   ├── orders-service/               # Orders (Postgres + Kafka outbox/inbox)
│   ├── payments-service/             # Payments (Postgres + Kafka outbox/inbox)
│   └── frontend/                     # React/Vite UI
├── scripts/                          # generate_code.sh, generate_sql.sh, create_topics.sh, lint
└── docker-compose.yaml
```

---

## Кодогенерация кода

### Protobuf + OpenAPI

```bash
bash scripts/generate_code.sh
```

Внутри скрипта:
- `buf generate` (protobuf)
- `oapi-codegen ... api-files/openapi/api-gateway.yaml` (HTTP API Gateway)

### sqlc (Postgres queries)

```bash
bash scripts/generate_sql.sh
```

### Линт спецификаций

```bash
bash scripts/check_api-files.sh
```
