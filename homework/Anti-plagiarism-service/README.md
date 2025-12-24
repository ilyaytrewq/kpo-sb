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

## 🔌 API Endpoints

Базовый URL: `http://158.160.186.61:8080/api/v1`

- `POST /works` — создать работу (assignment)
- `POST /works/{workId}/submissions` — загрузить файл работы и запустить проверку
- `GET /works/{workId}/reports` — получить отчеты по всем сабмитам работы
- `GET /submissions/{submissionId}` — получить детали сабмита и отчет
- `GET /works/{workId}/stats` — агрегированная статистика по работе
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
|
└── README.md
```

## 🚀 Quick Start

### Запуск сервисов 

```bash
docker compose up --build -d
cd embedding-service && docker compose up --build -d
cd ../file-storing && docker compose up --build -d
cd ../file-analisys && docker compose up --build -d
cd ../api-gateway && docker compose up --build -d
```

### Кодогенерация кода
#### использовалась кодогенерация на основе open api файлов
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
ilyatikhonov@MacBook-Pro-Ilya api-gateway % sqlc generate
```


## 🧩 CI/CD

### CI (GitHub Actions)

- Триггеры: `push` и `pull_request` в ветки `main` и `dev/hw3`.
- Детектит Go-модули проекта и папку с Python-тестами (если есть).
- Линт OpenAPI-спецификаций через Redocly.
- Для каждого Go-модуля: `gofmt` (проверка), `go vet`, `go test`.
- Для Python: `unittest` discovery (если найдена папка тестов).
- Для `push` после тестов запускает CD и интеграционные E2E-тесты.

### CD (Deploy на Yandex VPS)

- Вызывается из CI на `push` и деплоит конкретный коммит.
- Поиск базовой директории проекта по `api-gateway/docker-compose.yaml`.
- Подготовка SSH, синхронизация кода через `rsync`:
  - `main` → `/opt/anti-plagiarism/prod`
  - другие ветки → `/opt/anti-plagiarism/dev`
- На сервере: запись env-файлов из секретов, валидация `docker-compose.yaml`,
  `docker compose up -d --build` для всех сервисов, prune старых образов.
