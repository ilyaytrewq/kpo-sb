FROM golang:1.25.4 AS builder

WORKDIR /src

COPY gen ./gen

COPY services/payments-service/go.mod services/payments-service/go.sum ./services/payments-service/
WORKDIR /src/services/payments-service
RUN go mod download

WORKDIR /src
COPY services/payments-service ./services/payments-service

WORKDIR /src/services/payments-service
RUN go build -o /out/app ./cmd/main.go


FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
      poppler-utils \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/app /app/app

CMD ["/app/app"]
#k