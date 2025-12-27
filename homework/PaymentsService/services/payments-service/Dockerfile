FROM golang:1.22 AS builder

WORKDIR /src

COPY go.work go.work.sum ./
COPY gen ./gen
COPY services/payments-service ./services/payments-service

WORKDIR /src/services/payments-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/payments-service ./cmd

FROM alpine:3.20
RUN adduser -D -g '' appuser
USER appuser
COPY --from=builder /bin/payments-service /usr/local/bin/payments-service

ENTRYPOINT ["/usr/local/bin/payments-service"]
