# Практическое занятие 9. Работа с базой данных

## Цель занятия
Научиться интегрировать Spring Boot приложение с реляционной базой данных, использовать JPA/Hibernate для объектно-реляционного отображения и реализовывать репозитории для работы с данными.

## Теоретическая часть

### Ключевые концепции

**Entity (Сущность)** - это обычный Java-класс, который:
- Помечается аннотацией `@Entity`
- Соответствует таблице в базе данных
- Каждый экземпляр представляет строку в таблице
- Поля класса соответствуют колонкам таблицы

**Основные JPA аннотации:**
- `@Entity` - помечает класс как сущность
- `@Table` - задает имя таблицы в БД
- `@Id` - указывает на первичный ключ
- `@GeneratedValue` - определяет стратегию генерации ID
- `@Column` - маппинг поля на колонку таблицы
- `@OneToOne`, `@OneToMany`, `@ManyToOne`, `@ManyToMany` - аннотации для связей

**Как добавить Entity:**
```java
@Entity
@Table(name = "table_name")
public class MyEntity {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(name = "name", nullable = false, length = 100)
    private String name;
    
    // конструкторы, геттеры, сеттеры
}
```

**Repository (Репозиторий)** - интерфейс для работы с данными:
- Наследуется от `JpaRepository<Entity, KeyType>`
- Spring автоматически создает реализации
- Предоставляет готовые методы: save, findById, findAll, delete

**Пример репозитория:**
```java

public interface CarRepository extends JpaRepository<Car, Integer> {
    // Spring Data генерирует реализацию автоматически
    List<Car> findByEngineType(String engineType);
    
    // Кастомный запрос с помощью @Query
    @Query("SELECT c FROM Car c WHERE c.vin > :minVin")
    List<Car> findCarsWithVinGreaterThan(@Param("minVin") Integer minVin);
}
```

**HQL (Hibernate Query Language)** - объектно-ориентированный язык запросов:
- Работает с сущностями и их свойствами, а не с таблицами и колонками
- Поддерживает полиморфные запросы
- Автоматически присоединяет связанные сущности

**Примеры HQL:**
```java
// Базовый запрос
@Query("SELECT c FROM Car c")

// Запрос с JOIN
@Query("SELECT c FROM Car c JOIN c.engine e")

// Запрос с параметрами
@Query("SELECT c FROM Car c WHERE c.vin > :minVin")

// Запрос с агрегацией
@Query("SELECT COUNT(c) FROM Car c WHERE c.engineType = :type")
```

## Практическое задание

### Задание 1: Настройка окружения

**Цель:** Поднять базу данных PostgreSQL в Docker и настроить подключение.

**Шаги выполнения:**

1. **Установите Docker** (если не установлен):
    - Windows/Mac: https://docs.docker.com/get-docker/
    - Linux: `sudo apt-get install docker.io docker-compose`

2. **Создайте docker-compose.yml** в корне проекта:
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    container_name: transport_db
    environment:
      POSTGRES_DB: transport_db
      POSTGRES_USER: student
      POSTGRES_PASSWORD: student123
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U student -d transport_db"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

3. **Запустите базу данных:**
```bash
docker-compose up -d
```

4. **Проверьте работу:**
```bash
docker ps  # Должен показать работающий контейнер postgres
```

### Задание 2: Добавление зависимостей

**Цель:** Подключить необходимые библиотеки для работы с JPA и PostgreSQL.

**Шаги выполнения:**

Добавьте в `build.gradle.kts`:
```kotlin
dependencies {
    implementation("org.springframework.boot:spring-boot-starter-data-jpa")
    runtimeOnly("org.postgresql:postgresql")
    // остальные зависимости...
}
```

### Задание 3: Создание сущности Engine

**Цель:** Преобразовать иерархию двигателей в JPA Entity.

**Интерфейс Engine:**
```java
public interface Engine {
    boolean isCompatible(Customer customer, ProductionTypes type);
}
```

**Абстрактная сущность AbstractEngine:**
```java
@Getter
@Setter
@Entity
@Inheritance(strategy = InheritanceType.SINGLE_TABLE)
@DiscriminatorColumn(name = "engine_type")
public abstract class AbstractEngine implements Engine {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(name = "engine_type", insertable = false, updatable = false)
    private String engineType;
    
    // Остальная логика остается в дочерних классах
}
```

**Задача:** Создайте сущности для конкретных типов двигателей:

```java
@Entity
@DiscriminatorValue("PEDAL")
public class PedalEngine extends AbstractEngine {
    @Column(name = "pedal_size")
    private Integer pedalSize;
    
    // Конструкторы
    public PedalEngine() {}
    
    public PedalEngine(Integer pedalSize) {
        this.pedalSize = pedalSize;
    }
    
    // Геттеры и сеттеры
    public Integer getPedalSize() { return pedalSize; }
    public void setPedalSize(Integer pedalSize) { this.pedalSize = pedalSize; }
    
    // Реализация метода isCompatible
    @Override
    public boolean isCompatible(Customer customer, ProductionTypes type) {
        // Логика совместимости TODO доработайте по прошлым практикам
        return true;
    }
}

@Entity
@DiscriminatorValue("HAND") 
public class HandEngine extends AbstractEngine {
    // Реализуйте аналогично PedalEngine
}
```

### Задание 4: Создание сущности Car

**Цель:** Преобразовать класс Car в JPA Entity с связью к Engine.

**Интерфейс Transport:**
```java
public interface Transport {
    boolean isCompatible(Customer customer);
}
```

**Сущность Car:**
```java
@Entity
@Table(name = "cars")
@Getter
@Setter
@NoArgsConstructor
public class Car implements Transport {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer vin;
    
    @OneToOne(cascade = CascadeType.ALL, fetch = FetchType.LAZY)
    @JoinColumn(name = "engine_id")
    private AbstractEngine engine;
    
    // Конструктор, принимающий Engine
    public Car(AbstractEngine engine) {
        this.engine = engine;
    }
    
    // Реализация метода isCompatible
    @Override
    public boolean isCompatible(Customer customer) {
        return engine != null && engine.isCompatible(customer, ProductionTypes.CAR);
    }
}
```

**Проблема для решения:** Почему при запуске возникает ошибка? Исправьте конструкторы.

### Задание 5: Создание репозитория

**Цель:** Создать Spring Data JPA репозиторий для работы с Car.

**Интерфейс CarRepository:**
```java
public interface CarRepository extends JpaRepository<Car, Integer> {
    
    // Метод для поиска по типу двигателя и VIN больше указанного
    @Query("SELECT c FROM Car c JOIN c.engine e WHERE e.engineType = :engineType AND c.vin > :minVin")
    List<Car> findByEngineTypeAndVinGreaterThan(@Param("engineType") String engineType, 
                                               @Param("minVin") Integer minVin);
    
    // Дополнительные методы
    List<Car> findByEngine_EngineType(String engineType);
    
    @Query("SELECT COUNT(c) FROM Car c WHERE c.engine.engineType = :engineType")
    Long countByEngineType(@Param("engineType") String engineType);
}
```

### Задание 6: Обновление сервиса

**Цель:** Заменить хранение в памяти на работу с базой данных.

**Интерфейс CarProvider:**
```java
public interface CarProvider {
    Car takeCar(Customer customer);
}
```

**Обновленный HseCarService:**
```java
@Component
@RequiredArgsConstructor
public class HseCarService implements CarProvider {

    private final List<SalesObserver> observers = new ArrayList<>();
    private final CustomerProvider customerProvider;
    private final CarRepository carRepository;

    public void addObserver(SalesObserver observer) {
        observers.add(observer);
    }

    private void notifyObserversForSale(Customer customer, ProductionTypes productType, int vin) {
        observers.forEach(obs -> obs.onSale(customer, productType, vin));
    }

    @Sales
    public void sellCars() {
        var customers = customerProvider.getCustomers();
        customers.stream()
                .filter(customer -> Objects.isNull(customer.getCar()))
                .forEach(customer -> {
                    var car = this.takeCar(customer);
                    if (Objects.nonNull(car)) {
                        customer.setCar(car);
                        notifyObserversForSale(customer, ProductionTypes.CAR, car.getVin());
                    }
                });
    }

    @Override
    public Car takeCar(Customer customer) {
        var filteredCars = carRepository.findAll().stream()
                .filter(car -> car.isCompatible(customer))
                .toList();

        var firstCar = filteredCars.stream().findFirst();
        firstCar.ifPresent(car -> carRepository.delete(car));

        return firstCar.orElse(null);
    }

    public <T> Car addCar(CarFactory<T> carFactory, T carParams) {
        return carRepository.save(carFactory.create(carParams));
    }

    public Car addExistingCar(Car car) {
        return carRepository.save(car);
    }

    public Optional<Car> findByVin(Integer vin) {
        return carRepository.findById(vin);
    }

    public void deleteByVin(Integer vin) {
        carRepository.deleteById(vin);
    }

    public List<Car> getCarsWithFiltration(String engineType, Integer vin) {
        if (engineType != null && vin != null) {
            return carRepository.findByEngineTypeAndVinGreaterThan(engineType, vin);
        } else if (engineType != null) {
            return carRepository.findByEngine_EngineType(engineType);
        } else {
            return carRepository.findAll();
        }
    }

    public List<Car> getCars() {
        return carRepository.findAll();
    }
}
```

### Задание 7: Настройка конфигурации

**Цель:** Настроить подключение к базе данных.

Создайте `application.yml`:
```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/transport_db
    username: student
    password: student123
    driver-class-name: org.postgresql.Driver
  
  jpa:
    hibernate:
      ddl-auto: update
    show-sql: true
    properties:
      hibernate:
        format_sql: true
        dialect: org.hibernate.dialect.PostgreSQLDialect

logging:
  level:
    org.hibernate.SQL: DEBUG
    org.hibernate.type.descriptor.sql.BasicBinder: TRACE
```

## Тестирование

### Тест 1: Проверка подключения к БД
1. Запустите приложение
2. Проверьте логи - не должно быть ошибок подключения
3. В логах должны отображаться SQL-запросы создания таблиц

### Тест 2: CRUD операции через API
Используйте Swagger UI или Postman:

**Создание машины:**
```http
POST /api/cars
Content-Type: application/json

{
  "engineType": "PEDAL",
  "pedalSize": 10
}
```

**Проверка создания:**
```http
GET /api/cars
```

**Поиск по VIN:**
```http
GET /api/cars/1
```

### Тест 3: Проверка фильтрации
```http
GET /api/cars?engineType=PEDAL&minVin=0
```

## Дополнительные задания

### Задание 8: Добавьте сущность Customer
```java
@Entity
@Table(name = "customers")
public class Customer {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(name = "name", nullable = false)
    private String name;
    
    @Column(name = "age")
    private Integer age;
    
    @OneToOne
    @JoinColumn(name = "car_vin")
    private Car car;
    
    // Конструкторы
    public Customer() {}
    
    public Customer(String name, Integer age) {
        this.name = name;
        this.age = age;
    }
    
    // Геттеры и сеттеры
    public Long getId() { return id; }
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public Integer getAge() { return age; }
    public void setAge(Integer age) { this.age = age; }
    public Car getCar() { return car; }
    public void setCar(Car car) { this.car = car; }
}
```

### Задание 9: Создайте репозиторий для Customer

### Задание 10: Добавьте миграции базы данных
Используйте Flyway для управления изменениями схемы БД:

1. Добавьте зависимость:
```kotlin
implementation("org.flywaydb:flyway-core")
```

2. Создайте миграции в `src/main/resources/db/migration`:
- `V1__Create_cars_table.sql`
- `V2__Create_customers_table.sql`

## Критерии успешного выполнения

- [ ] База данных запускается через Docker
- [ ] Приложение успешно подключается к БД
- [ ] Создаются все необходимые таблицы
- [ ] CRUD операции работают через API
- [ ] Фильтрация по типу двигателя и VIN работает
- [ ] Метод продажи машин корректно удаляет их из БД

## Подсказки

1. **Если Docker не запускается:** проверьте, что Docker Desktop запущен
2. **Если ошибка подключения:** проверьте логин/пароль в application.yml
3. **Если не создаются таблицы:** проверьте настройки Hibernate ddl-auto
4. **Для просмотра данных в БД:** используйте pgAdmin или DBeaver

## Полезные ресурсы

1. [Spring Data JPA Documentation](https://spring.io/projects/spring-data-jpa)
2. [Hibernate ORM Guide](https://hibernate.org/orm/documentation/)
3. [PostgreSQL Docker Image](https://hub.docker.com/_/postgres)
4. [JPA Entity Mapping](https://docs.jboss.org/hibernate/orm/current/userguide/html_single/Hibernate_User_Guide.html#entity)

Успехов в выполнении задания!