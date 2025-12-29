FROM golang:1.25.4 AS builder

WORKDIR /src

COPY gen ./gen

COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
WORKDIR /src/services/api-gateway
RUN go mod download

WORKDIR /src
COPY services/api-gateway ./services/api-gateway

WORKDIR /src/services/api-gateway
RUN go build -o /out/app ./cmd/main.go


FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
      poppler-utils \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/app /app/app

CMD ["/app/app"]

# я люблю илюшууу мяумяу лялялялял 