module github.com/ilyaytrewq/payments-service/api-gateway

go 1.22.0

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/google/uuid v1.6.0
	github.com/ilyaytrewq/payments-service/gen v0.0.0
	google.golang.org/grpc v1.78.0
)

replace github.com/ilyaytrewq/payments-service/gen => ../../gen
