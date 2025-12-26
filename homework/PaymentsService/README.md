```bash 
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
export PATH="$(go env GOPATH)/bin:$PATH"

chmod +x scripts/check_api-files.sh 
chmod +x scripts/generate_code.sh
chmod +x scripts/generate_sql.sh
 
./scripts/check_api-files.sh
./scripts/generate_code.sh
./scripts/generate_sql.sh

```