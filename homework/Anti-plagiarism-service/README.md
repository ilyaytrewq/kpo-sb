# Anti-Plagiarism Service

Сервис антиплагиата для управления работами студентов, их отправками и отчетами проверки.

## 🛠 Tech Stack

- **Go** — backend
- **Chi** — HTTP router
- **OpenAPI 3.0** — API specification
- **oapi-codegen** — code generation

## 📁 Project Structure

```
.
├── openapi.yaml              # OpenAPI specification
├── internal/
│   └── api/
│       └── generated.go      # Generated server & types
├── go.mod
└── README.md
```

## 🚀 Getting Started

### Generate server code from OpenAPI

```bash
oapi-codegen -generate chi-server,types -package api -o internal/api/generated.go openapi.yaml
```

### Install dependencies

```bash
go mod tidy
```

## 📚 API Endpoints

| Method | Endpoint                        | Description                  |
|--------|---------------------------------|------------------------------|
| POST   | `/works`                        | Create a new work            |
| POST   | `/works/{workId}/submissions`   | Submit work for review       |
| GET    | `/works/{workId}/reports`       | Get analytics by workId      |
| GET    | `/works/{workId}/stats`         | Get statistics by workId     |
| GET    | `/submissions/{submissionId}`   | Get submission details       |

