# Занятие 9. Controllers

## Цель занятия
Научиться создавать REST API для взаимодействия с внешними системами через Spring контроллеры. Понимать принципы валидации и документирования API.

## Требования к реализации
1. Реализовать CRUD + операцию продажи для автомобилей через REST
2. Добавить валидацию входных данных
3. Настроить документацию API через Swagger

## Тестирование
Протестировать все endpoints через Swagger UI после реализации

## Задание на доработку
- Реализовать аналогичные контроллеры для катамаранов, покупателей, отчетов

---

## Пояснения к реализации

### 1. Зачем нужны зависимости
**Обязательные зависимости в `build.gradle`:**
```gradle
dependencies {
    // Spring Web - основа для REST контроллеров
    implementation("org.springframework.boot:spring-boot-starter-web")
    // Swagger - документация и тестирование API
    implementation("org.springdoc:springdoc-openapi-starter-webmvc-ui:2.5.0")
}
```

**Почему именно эти зависимости:**
- `spring-boot-starter-web` включает Tomcat сервер и все необходимое для REST
- Swagger позволяет автоматически генерировать документацию и предоставляет UI для тестирования

### 2. Конфигурация Swagger
**Обязательная конфигурация:**
```java
@Configuration
public class SwaggerConfig {
    @Bean
    public OpenAPI customOpenAPI() {
        return new OpenAPI()
                .info(new Info()
                        .title("HSE Car Service API")
                        .version("1.0")
                        .description("API для управления автомобилями"));
    }
}
```

**Зачем это нужно:** Эта конфигурация создает красивую документацию с описанием нашего API, которая будет доступна по адресу `http://localhost:8080/swagger-ui/index.html`

### 3. DTO и валидация
**Record для запроса:**
```java
public record CarRequest(
        @Schema(description = "Тип двигателя", example = "PEDAL")
        @Pattern(regexp = "PEDAL|HAND|LEVITATION", message = "Допустимые значения: PEDAL, HAND, LEVITATION")
        String engineType,

        @Schema(description = "Размер педалей (1-15)", example = "6")
        @Min(value = 1, message = "Минимальный размер педалей - 1")
        @Max(value = 15, message = "Максимальный размер педалей - 15")
        @Nullable
        Integer pedalSize
) {}
```

**Почему именно так:**
- DTO (Data Transfer Object) отделяет API от внутренней модели
- Валидация на уровне контроллера предотвращает некорректные данные
- Аннотации `@Schema` улучшают документацию в Swagger

### 4. Enum для типов двигателей
**Зачем нужен:** Безопасная обработка строковых значений, защита от опечаток

### 5. Структура контроллера
**Ключевые моменты реализации:**

**Аннотации класса:**
```java
@RestController
@RequestMapping("/api/cars")
@RequiredArgsConstructor
@Tag(name = "Автомобили", description = "Управление транспортными средствами")
```
- `@RestController` - маркирует класс как REST контроллер
- `@RequestMapping` - задает базовый путь
- `@Tag` - группирует endpoints в Swagger

**Пример GET endpoint:**
```java
@GetMapping("/{vin}")
@Operation(summary = "Получить автомобиль по VIN")
public ResponseEntity<Car> getCarByVin(@PathVariable int vin) {
    // Логика поиска по VIN
    return ResponseEntity.ok(car); // 200 OK
    // или return ResponseEntity.notFound().build(); // 404
}
```

**Пример POST с валидацией:**
```java
@PostMapping
@Operation(summary = "Создать автомобиль")
public ResponseEntity<Car> createCar(
        @Valid @RequestBody CarRequest request,
        BindingResult bindingResult) {
    
    if (bindingResult.hasErrors()) {
        throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "Сообщение об ошибке");
    }
    // Логика создания
    return ResponseEntity.status(HttpStatus.CREATED).body(car); // 201 Created
}
```

**Почему ResponseEntity:** Позволяет тонко настраивать HTTP статусы и headers

### 6. Паттерны проектирования
- **Facade**: `HseFacade` скрывает сложность создания автомобилей
- **DTO**: Отделение API модели от доменной модели
- **Dependency Injection**: Контроллер получает зависимости через конструктор

### 7. HTTP статусы
- `200 OK` - успешный запрос
- `201 Created` - ресурс создан
- `400 Bad Request` - ошибка валидации
- `404 Not Found` - ресурс не найден
- `204 No Content` - успешное удаление

---

## Практические шаги

1. **Добавьте зависимости** в `build.gradle`
2. **Создайте SwaggerConfig** в package `config`
3. **Реализуйте CarRequest** с валидацией
4. **Создайте CarController** с основными операциями:
    - GET /api/cars - список всех
    - GET /api/cars/{vin} - по VIN
    - POST /api/cars - создание
    - PUT /api/cars/{vin} - обновление
    - DELETE /api/cars/{vin} - удаление
    - POST /api/cars/sell/{vin} - продажа конкретного
    - POST /api/cars/sell - продажа всех

5. **Протестируйте** через Swagger UI

---

## Ссылки для изучения
1. [REST API и CRUD - основы](https://proglib.io/p/chto-takoe-api-i-crud-prostymi-slovami)
2. [Spring REST контроллеры](https://javarush.com/groups/posts/3160-spring---ehto-ne-strashno-kontroliruem-svoy-rest)
3. [Лучшие практики REST API](https://habr.com/ru/articles/541592/)
4. [Postman для тестирования API](https://www.postman.com/)
5. [Swagger UI](http://localhost:8080/swagger-ui/index.html#/)
