# Anti-Plagiarism Service
Микросервисная система для проверки студенческих работ на плагиат с использованием векторных эмбеддингов.

## 🏗️ Архитектура

### 4 микросервиса:

1. **api-gateway** (`:8080`) — публичный API, маршрутизация запросов
2. **file-storing** (`:8081`) — хранение файлов и векторных эмбеддингов
3. **file-analisys** (`:8082`) — оркестрация проверки на плагиат
4. **embedding-service** (`:8083`) — генерация векторных представлений текста

### Пайплайн обработки:

```
Client → API Gateway → File Storing (загрузка файла, получение fileId)
                    → File Analysis (запуск проверки с fileId)
                    → Embedding Service (векторизация chunks)
                    → File Storing (сохранение/поиск embeddings)
                    → БД (сохранение результатов)
Client ← API Gateway ← File Analysis (получение отчета)
```

## 🛠 Tech Stack

- **Go 1.25+** - backend
- **Chi** - HTTP router
- **OpenAPI 3.0** - API specification
- **oapi-codegen** - code generation from OpenAPI
- **PostgreSQL + pgvector** - vector database
- **Yandex llm** - embedddings
- **Yandex S3** - file storing
## 📁 Project Structure

```
.
├── api-files/                    # OpenAPI specifications
│   ├── openapi.yaml             # API Gateway spec
│   ├── file-storing.yaml        # File Storing spec
│   ├── file-analisys.yaml       # File Analysis spec
│   └── embedding-service.yaml   # Embedding Service spec
├── api-gateway/
│   ├── cmd/main.go
│   └── internal/
│       ├── api/generated.go     # Generated server code
│       ├── handlers/            # Handler implementations
│       └── clients/             # Generated clients
│           ├── filestoring/
│           └── fileanalisys/
├── file-storing/
│   └── internal/
│       ├── api/generated.go
│       └── handlers/
├── file-analisys/
│   └── internal/
│       ├── api/generated.go
│       ├── handlers/
│       └── clients/
│           ├── embedding/
│           └── filestoring/
├── embedding-service/
│   └── internal/
│       ├── api/generated.go
│       └── handlers/
├── docs/                         # Documentation
│   ├── ARCHITECTURE.md          # Full architecture
│   ├── API_UPDATE_SUMMARY.md    # API changes summary
│   └── CLIENTS_USAGE.md         # Client usage examples
├── CHECKLIST.md                  # Development checklist
├── OPENAPI_UPDATE_COMPLETE.md    # Update report
└── README.md
```

## 🚀 Quick Start

### 1. Кодогенерация кода

```bash
# Серверный код
oapi-codegen -generate chi-server,types -package api \
  -o ./api-gateway/internal/api/generated/generated.go \
  ./api-files/openapi.yaml

oapi-codegen -generate chi-server,types -package api \
  -o ./file-analisys/internal/api/generated/generated.go \
  ./api-files/file-analisys.yaml

oapi-codegen -generate chi-server,types -package api \
  -o ./file-storing/internal/api/generated/generated.go \
  ./api-files/file-storing.yaml

oapi-codegen -generate chi-server,types -package api \
  -o ./embedding-service/internal/api/generated/generated.go \
  ./api-files/embedding-service.yaml

# Клиентский код API Gateway
oapi-codegen -generate client,types -package filestoring \
  -o ./api-gateway/internal/clients/filestoring/client.go \
  ./api-files/file-storing.yaml

oapi-codegen -generate client,types -package fileanalysis \
  -o ./api-gateway/internal/clients/fileanalysis/client.go \
  ./api-files/file-analisys.yaml

oapi-codegen -generate client,types -package embedding \
  -o ./file-analisys/internal/clients/embedding/client.go \
  ./api-files/embedding-service.yaml

oapi-codegen -generate client,types -package filestoring \
  -o ./file-analisys/internal/clients/filestoring/client.go \
  ./api-files/file-storing.yaml

```

```bash
npx @redocly/cli lint ./api-files/openapi.yaml
npx @redocly/cli lint ./api-files/file-storing.yaml      
npx @redocly/cli lint ./api-files/file-analisys.yaml          
npx @redocly/cli lint ./api-files/embedding-service.yaml
  
```

```bash
ilyatikhonov@MacBook-Pro-Ilya api-gateway % sqlc generate
```

### 2. Запуск сервисов (когда реализованы handlers)


```bash
# из корня репозитория
bash ./run.sh

# или по отдельности
cd embedding-service && docker compose up -d
cd ../file-storing && docker compose up -d
cd ../file-analisys && docker compose up -d
cd ../api-gateway && docker compose up -d
```


## 🔌 API Endpoints

### API Gateway (`:8080`)

```bash
# Создать работу
POST /works
{
  "workId": "hw-kpo-3",
  "name": "KPO Homework 3",
  "description": "Anti-plagiarism homework"
}

# Загрузить submission
POST /works/{workId}/submissions
Content-Type: multipart/form-data
- file: <binary>
- studentId: "student-123"

# Получить отчеты по работе
GET /works/{workId}/reports

# Получить детали submission
GET /submissions/{submissionId}

# Статистика по работе
GET /works/{workId}/stats
```

### Примеры запросов (curl)

```bash
BASE_URL="http://localhost:8080/api/v1"

# Проверка доступности API Gateway
curl -s http://localhost:8080/health

# Создать работу
curl -X POST "$BASE_URL/works" \
  -H "Content-Type: application/json" \
  -d '{"workId":"hw-kpo-3","name":"KPO HW-3","description":"Anti-plagiarism homework"}'

# Загрузить submission (multipart/form-data)
curl -X POST "$BASE_URL/works/hw-kpo-3/submissions" \
  -F "file=@/path/to/hw3.pdf"

# Получить отчеты по работе
curl "$BASE_URL/works/hw-kpo-3/reports"

# Получить детали submission
curl "$BASE_URL/submissions/sub-001"

# Статистика по работе
curl "$BASE_URL/works/hw-kpo-3/stats"
```

## 📚 API Endpoints

| Method | Endpoint                        | Description                  |
|--------|---------------------------------|------------------------------|
| POST   | `/works`                        | Create a new work            |
| POST   | `/works/{workId}/submissions`   | Submit work for review       |
| GET    | `/works/{workId}/reports`       | Get analytics by workId      |
| GET    | `/works/{workId}/stats`         | Get statistics by workId     |
| GET    | `/submissions/{submissionId}`   | Get submission details       |
