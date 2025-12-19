# Anti-Plagiarism Service

Микросервисная система для проверки студенческих работ на плагиат с использованием векторных эмбеддингов.

## ✅ Статус проекта

**OpenAPI спецификации:** ✅ Обновлены (11 декабря 2025)  
**Генерация кода:** ✅ Завершена  
**Документация:** ✅ Создана  
**Реализация:** 🚧 В процессе

📋 **См. [CHECKLIST.md](./CHECKLIST.md)** для детального прогресса

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

**Полная документация:** [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)

## 🛠 Tech Stack

- **Go 1.25+** — backend
- **Chi** — HTTP router
- **OpenAPI 3.0** — API specification
- **oapi-codegen** — code generation from OpenAPI
- **PostgreSQL + pgvector** — vector database
- **OpenAI** — text-embedding-3-small (1536 dimensions)
- **Docker** — containerization

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

### 1. Регенерация кода (если изменили OpenAPI)

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

### 2. Установка зависимостей

```bash
# Для каждого сервиса
cd api-gateway && go mod tidy
cd ../file-storing && go mod tidy
cd ../file-analisys && go mod tidy
cd ../embedding-service && go mod tidy
```

### 3. Настройка PostgreSQL с pgvector

```bash
docker run -d --name anti-plagiarism-db \
  -e POSTGRES_DB=anti_plagiarism \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  pgvector/pgvector:pg16
```

### 4. Запуск сервисов (когда реализованы handlers)

```bash
# В отдельных терминалах
cd api-gateway && go run cmd/main.go
cd file-storing && go run cmd/main.go
cd file-analisys && go run cmd/main.go
cd embedding-service && go run cmd/main.go
```

## 📖 Документация

| Документ | Описание |
|----------|----------|
| [📋 CHECKLIST.md](./CHECKLIST.md) | Чеклист задач и прогресс |
| [✅ OPENAPI_UPDATE_COMPLETE.md](./OPENAPI_UPDATE_COMPLETE.md) | Отчет об обновлении |
| [🏗️ docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) | Полная архитектура системы |
| [📝 docs/API_UPDATE_SUMMARY.md](./docs/API_UPDATE_SUMMARY.md) | Изменения в API |
| [💻 docs/CLIENTS_USAGE.md](./docs/CLIENTS_USAGE.md) | Примеры использования клиентов |

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

### Ответ при загрузке submission:

```json
{
  "submissionId": "sub-001",
  "workId": "hw-kpo-3",
  "studentId": "student-123",
  "fileId": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "status": "QUEUED",
  "uploadedAt": "2025-12-11T12:30:00Z",
  "message": "Submission accepted. Plagiarism check is queued."
}
```

### Ответ с результатами проверки:

```json
{
  "submissionId": "sub-001",
  "workId": "hw-kpo-3",
  "studentId": "student-123",
  "fileId": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "status": "DONE",
  "uploadedAt": "2025-12-11T12:30:00Z",
  "report": {
    "reportId": "rep-001",
    "status": "DONE",
    "plagiarismDetected": true,
    "similarityPercent": 78.5,
    "createdAt": "2025-12-11T12:30:00Z",
    "completedAt": "2025-12-11T12:35:00Z",
    "matchedSubmissions": [
      {
        "submissionId": "sub-042",
        "studentId": "student-789",
        "similarityPercent": 78.5,
        "matchedChunks": 15
      }
    ]
  }
}
```

## 🎯 Ключевые особенности

- **Изоляция по workId** — каждая работа имеет свою таблицу embeddings
- **Асинхронная обработка** — клиент получает 202 Accepted и проверяет статус позже
- **Chunking** — документы разбиваются на части для точности и обхода лимитов
- **Векторный поиск** — cosine similarity для определения похожести
- **Threshold 50%** — порог для определения плагиата

## 🔧 Следующие шаги

1. 📝 Реализовать handlers (см. [CHECKLIST.md](./CHECKLIST.md))
2. 🗄️ Создать миграции баз данных
3. 🧪 Написать тесты
4. 🐳 Создать Docker образы
5. 📊 Добавить мониторинг

## 🤝 Contributing

1. Изучите [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)
2. Посмотрите [CHECKLIST.md](./CHECKLIST.md) для списка задач
3. Выберите задачу и создайте PR

## 📄 License

MIT

---

**Дата последнего обновления:** 11 декабря 2025  
**Статус:** OpenAPI спецификации готовы, начинаем реализацию
```

## 📚 API Endpoints

| Method | Endpoint                        | Description                  |
|--------|---------------------------------|------------------------------|
| POST   | `/works`                        | Create a new work            |
| POST   | `/works/{workId}/submissions`   | Submit work for review       |
| GET    | `/works/{workId}/reports`       | Get analytics by workId      |
| GET    | `/works/{workId}/stats`         | Get statistics by workId     |
| GET    | `/submissions/{submissionId}`   | Get submission details       |

