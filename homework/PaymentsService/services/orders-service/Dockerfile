FROM golang:1.22 AS builder

WORKDIR /src

COPY go.work go.work.sum ./
COPY gen ./gen
COPY services/orders-service ./services/orders-service

WORKDIR /src/services/orders-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/orders-service ./cmd

FROM alpine:3.20
RUN adduser -D -g '' appuser
USER appuser
COPY --from=builder /bin/orders-service /usr/local/bin/orders-service

ENTRYPOINT ["/usr/local/bin/orders-service"]
