# Anti-Plagiarism Service
Микросервисная система для проверки студенческих работ на плагиат с использованием векторных эмбеддингов.

## 🏗️ Архитектура

### 4 микросервиса:

1. **api-gateway** (`:8080`) — публичный API, маршрутизация запросов
2. **file-storing** (`:8082`) — хранение загруженных файлов (S3)
3. **file-analisys** (`:8081`) — оркестрация проверки на плагиат
4. **embedding-service** (`:8083`) — генерация векторных представлений текста

### Пайплайн обработки:

```
Client → API Gateway → File Storing (загрузка файла, получение fileId)
                    → File Analysis (скачивание файла, chunking, запуск проверки)
                    → Embedding Service (векторизация chunks)
                    → Qdrant (поиск похожих векторов)
                    → Postgres (сохранение результатов)
Client ← API Gateway ← File Analysis (получение отчета)
```

## ⚙️ Асинхронная обработка

- `AnalyzeFile` возвращает 202 и ставит задачу в очередь.
- Обработка выполняется воркер пулом в `file-analisys` (пакет `filequeue`).
- Статусы отчета: `QUEUED → PROCESSING → DONE/ERROR`.
- Настройки: `FILEQUEUE_WORKERS`, `FILEQUEUE_SIZE`.

## 🛠 Tech Stack

- **Go 1.25+** — backend
- **Chi** — HTTP router
- **OpenAPI 3.0** — API specification
- **oapi-codegen** — code generation from OpenAPI
- **PostgreSQL** — хранение работ/отчетов
- **Qdrant** — векторный поиск
- **Yandex Cloud Embeddings** — генерация эмбеддингов
- **S3-compatible storage** — хранение файлов (Yandex Object Storage / MinIO)
- **Docker / Docker Compose** — окружение и запуск
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

## 🔌 API Endpoints и примеры запросов

Ниже — все эндпоинты с примерами запуска. Для локальной проверки можно использовать файлы из `tests_files/`.

### API Gateway (`:8080`)

База: `http://localhost:8080/api/v1`  
Health: `GET http://localhost:8080/health`

```bash
GATEWAY_URL="http://localhost:8080/api/v1"

# Health
curl -s http://localhost:8080/health

# Создать работу
curl -X POST "$GATEWAY_URL/works" \
  -H "Content-Type: application/json" \
  -d '{"workId":"hw-kpo-3","name":"KPO HW-3","description":"Anti-plagiarism homework"}'

# Загрузить submission
curl -X POST "$GATEWAY_URL/works/hw-kpo-3/submissions" \
  -F "file=@tests_files/sample_short.txt"

# Получить отчеты по работе
curl "$GATEWAY_URL/works/hw-kpo-3/reports"

# Получить детали submission
curl "$GATEWAY_URL/submissions/sub-001"

# Статистика по работе
curl "$GATEWAY_URL/works/hw-kpo-3/stats"
```

### File Storing (`:8082`)

База: `http://localhost:8082/api/v1`  
Health: `GET http://localhost:8082/health`

```bash
STORING_URL="http://localhost:8082/api/v1"

# Health
curl -s http://localhost:8082/health

# Загрузить файл
curl -X POST "$STORING_URL/files/upload" \
  -F "file=@tests_files/sample_short.txt" \
  -F 'metadata={"workId":"hw-kpo-3","originalFileName":"sample_short.txt","contentType":"text/plain"};type=application/json'

# Скачать файл
curl -O -J "$STORING_URL/files/f47ac10b-58cc-4372-a567-0e02b2c3d479"

# Метаданные файла
curl "$STORING_URL/files/f47ac10b-58cc-4372-a567-0e02b2c3d479/info"
```

### File Analysis (`:8081`)

База: `http://localhost:8081/api/v1`  
Health: `GET http://localhost:8081/health`

```bash
ANALYSIS_URL="http://localhost:8081/api/v1"

# Health
curl -s http://localhost:8081/health

# Запустить анализ
curl -X POST "$ANALYSIS_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{"fileId":"f47ac10b-58cc-4372-a567-0e02b2c3d479","workId":"hw-kpo-3","submissionId":"sub-001"}'

# Получить отчет по submissionId
curl "$ANALYSIS_URL/reports/sub-001"

# Получить все отчеты по работе
curl "$ANALYSIS_URL/works/hw-kpo-3/reports"
```

### Embedding Service (`:8083`)

База: `http://localhost:8083/api/v1`  
Health: `GET http://localhost:8083/health`

```bash
EMBEDDING_URL="http://localhost:8083/api/v1"

# Health
curl -s http://localhost:8083/health

# Получить эмбеддинги для чанков
curl -X POST "$EMBEDDING_URL/embed" \
  -H "Content-Type: application/json" \
  -d '{"chunks":[{"chunkId":"chunk-001","text":"Hello world","chunkIndex":0},{"chunkId":"chunk-002","text":"Another chunk","chunkIndex":1}]}'
```

## 🧩 CI/CD

CI/CD добавлен.
