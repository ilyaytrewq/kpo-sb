set -euo pipefail

rm -rf gen/go
buf generate

mkdir -p gen/openapi/gateway
oapi-codegen \
  -package gateway \
  -generate types,chi-server,spec \
  -o gen/openapi/gateway/gateway.gen.go \
  api-files/openapi/api-gateway.yaml
