Отлично — раз OpenAPI-файлы сгенерированы и Swagger больше не падает, дальше ты переходишь от **контрактов** к **запуску инфраструктуры + реализации сервисов**. Ниже — порядок шагов, который обычно гарантированно доводит до “всё поднимается в docker-compose и работает по пайплайну”.

## 1) Проверь, что спеки валидные (чтобы дальше не ловить “failed to load spec”)

В корне репозитория:

```bash
npx @redocly/cli lint ./api-files/openapi.yaml
npx @redocly/cli lint ./api-files/file-storing.yaml
npx @redocly/cli lint ./api-files/file-analisys.yaml
npx @redocly/cli lint ./api-files/embedding-service.yaml
```

Redocly lint — стандартный быстрый способ найти битые `$ref`, несовпадения схем и т.п. ([redocly.com][1])

(Если ты пользуешься openapi-generator — у него тоже есть CLI/контейнер для генерации клиентов/серверов.) ([hub.docker.com][2])

## 2) Подними Swagger UI локально, чтобы глазами проверить “как будет смотреться”

Самый простой способ — docker swagger-ui с переменной `SWAGGER_JSON` (поддерживается официально):

```bash
docker run --rm -p 8080:8080 \
  -e SWAGGER_JSON=/spec/file-storing.yaml \
  -v "$(pwd)/api-files:/spec" \
  docker.swagger.io/swaggerapi/swagger-ui
```

Так Swagger UI точно подхватит твой файл. ([swagger.io][3])

> Если у тебя снова `zsh: command not found: #` — включи комментарии:
> `setopt interactivecomments` ([Unix & Linux Stack Exchange][4])

## 3) Подними инфраструктуру (docker-compose): Postgres + MinIO(S3) + Qdrant

**Почему так:**

* File Storing удобно делать через **S3-совместимое хранилище** (локально — MinIO).
* Вектора и similarity search — в **Qdrant** (REST 6333 / gRPC 6334, есть Web UI). ([qdrant.tech][5])

Минимум, который должен быть доступен после `docker compose up -d`:

* MinIO API (9000) + console (обычно 9001) — креды задаются через `MINIO_ROOT_USER/MINIO_ROOT_PASSWORD` ([Chainguard Academy][6])
* Qdrant: REST `:6333`, gRPC `:6334`, dashboard на `:6333/dashboard` ([qdrant.tech][5])
* Postgres для gateway и analysis

## 4) Реализуй сервисы по очереди (так проще отлаживать)

### A) File Storing (самый быстрый “закрыть”)

**Что сделать:**

* `POST /files/upload` → кладёшь объект в MinIO/S3, возвращаешь `fileId`
* `GET /files/{fileId}` → отдаёшь байты
* `GET /files/{fileId}/info` → метаданные

**Go + S3/MinIO:**

* AWS SDK for Go v2 умеет кастомные endpoint’ы (нужно для MinIO). ([AWS Documentation][7])
* Для MinIO чаще нужен path-style режим (`UsePathStyle/force_path_style` идея ровно про это). ([GitHub][8])

### B) Embedding Service (пока можно mock)

Для ДЗ обычно достаточно:

* принять chunks
* вернуть детерминированные “вектора” (например, из хэша текста), чтобы Qdrant-поиск работал воспроизводимо

Если будешь реально дергать OpenAI — учти, что размерность зависит от модели (small vs large). ([GitHub][9])

### C) File Analysis (сердце пайплайна)

**Минимальный рабочий путь:**

1. `POST /analyze` → сразу `202 QUEUED`, сохраняешь report со статусом `QUEUED`
2. фоновый воркер:

   * скачал файл из File Storing
   * извлёк текст → чанки
   * эмбеддинги (embedding-service)
   * upsert в Qdrant + search похожих (внутри workId, исключая текущий submissionId)
   * посчитал similarityPercent + plagiarismDetected
   * сохранил report как `DONE`

Qdrant можно дергать по gRPC — у них есть официальный Go-клиент. ([GitHub][10])

### D) API Gateway (последним)

Потому что он склеивает всё:

* `POST /works/{workId}/submissions`:

  1. грузит файл в file-storing
  2. запускает analysis `/analyze`
  3. пишет submission в свою БД
* `GET /works/{workId}/reports` проксирует из analysis

## 5) Сборка end-to-end проверки (обязательно перед сдачей)

1. `docker compose up -d`
2. Swagger UI открывается и показывает все спеки (без ошибок)
3. `curl` сценарий:

   * создать work
   * отправить submission
   * поллить report, пока не DONE
4. (бонус) wordcloud эндпоинт отдаёт PNG

---

Если хочешь, я дам **конкретный “Definition of Done” чек-лист** под каждый сервис (что логировать, какие ошибки возвращать, какие статусы менять) и **готовый набор curl-команд** под твои endpoint’ы, чтобы ты просто прогнал и сделал скриншоты для отчёта.

[1]: https://redocly.com/docs/cli/commands/lint?utm_source=chatgpt.com "lint"
[2]: https://hub.docker.com/r/openapitools/openapi-generator-cli?utm_source=chatgpt.com "openapitools/openapi-generator-cli - Docker Image"
[3]: https://swagger.io/docs/open-source-tools/swagger-ui/usage/installation/?utm_source=chatgpt.com "Installation | Swagger Docs"
[4]: https://unix.stackexchange.com/questions/557486/allowing-comments-in-interactive-zsh-commands?utm_source=chatgpt.com "Allowing comments in interactive zsh commands"
[5]: https://qdrant.tech/documentation/quickstart/?utm_source=chatgpt.com "Local Quickstart"
[6]: https://edu.chainguard.dev/chainguard/chainguard-images/getting-started/minio/?utm_source=chatgpt.com "Getting Started with the MinIO Chainguard Container"
[7]: https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-endpoints.html?utm_source=chatgpt.com "Configure Client Endpoints - AWS SDK for Go v2"
[8]: https://github.com/hashicorp/terraform-provider-aws/issues/23026?utm_source=chatgpt.com "provider: S3 use path style configuration · Issue #23026"
[9]: https://github.com/swagger-api/swagger-ui/issues/8591?utm_source=chatgpt.com "When using Swagger on Docker the parameter ..."
[10]: https://github.com/qdrant/go-client?utm_source=chatgpt.com "Go client for Qdrant vector search engine"
