FROM golang:1.22 AS builder

WORKDIR /src

COPY go.work go.work.sum ./
COPY gen ./gen
COPY services/api-gateway ./services/api-gateway

WORKDIR /src/services/api-gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/api-gateway ./cmd

FROM alpine:3.20
RUN adduser -D -g '' appuser
USER appuser
COPY --from=builder /bin/api-gateway /usr/local/bin/api-gateway

ENTRYPOINT ["/usr/local/bin/api-gateway"]
