# Запуск приложения

## Стандартный способ (с Docker)

```bash
# Клонировать репозиторий
git clone https://github.com/Tbits007/grpc-usdt-rate.git
cd grpc-usdt-rate

# Собрать и запустить
docker-compose up --build -d
```

## Makefile команды

- `make build`  
  Сборка приложения, создаёт бинарный файл `./bin/usdt-service`.

- `make run`  
  Запуск приложения локально через `go run`.

- `make test`  
  Запуск unit-тестов с подробным выводом и покрытием кода.

- `make docker-build`  
  Сборка Docker-образа с приложением под тегом `grpc-usdt-rate`.

- `make docker-up`  
  Сборка Docker-образа и запуск контейнеров в фоне через `docker-compose up -d`.

- `make docker-down`  
  Остановка и удаление запущенных контейнеров `docker-compose`.

- `make lint`  
  Запуск статического анализа кода с помощью `golangci-lint`.

# gRPC API

### 1. Получение курса USDT
**Endpoint**: `RateService/GetRates`  
**Тип**: Unary RPC  
**Описание**: Возвращает текущие курсы покупки (bid) и продажи (ask) USDT  
**Запрос**: Пустое сообщение (`{}`)  
**Ответ**: 
```json
{
    "ask": 78.76,
    "bid": 78.7,
    "timestamp": "1753304472"
}
```

## 2. Проверка работоспособности сервиса (HealthCheck)

**Endpoint**: `RateService/HealthCheck`  
**Тип**: Unary RPC (одиночный запрос-ответ)  
**Назначение**: Проверка статуса работы сервиса и зависимостей  

**Формат запроса**:
```json
{}
```
**Ответ**: 
```json
{
  "status": true,
}
```

## Конфигурация

Сервис можно настроить с использованием либо флагов командной строки, либо переменных окружения. Флаги имеют приоритет над переменными окружения.

### Параметры конфигурации

| Флаг            | Переменная окружения | Значение по умолчанию                              | Описание                               |
|-----------------|-----------------------|----------------------------------------------------|----------------------------------------|
| `-postgres-dsn` | `POSTGRES_DSN`        | `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable` | Строка подключения к PostgreSQL |
| `-grpc-port`    | `GRPC_PORT`           | `50051`                                            | Порт gRPC сервера                     |
| `-metrics-port` | `METRICS_PORT`        | `2112`                                             | Порт сервера метрик                   |
| `-otlp-endpoint`| `OTLP_ENDPOINT`       | `otel-collector:4317`                              | Endpoint OTLP коллектора               |
| `-service-name` | `SERVICE_NAME`        | `usdt-rate-service`                                | Имя сервиса для трейсинга              |

### Примеры использования

Запуск с флагами:
```bash
go build -o ./bin/usdt-service.exe ./cmd/app/main.go

./bin/usdt-service \
  -postgres-dsn="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" \
  -grpc-port=50051 \
  -metrics-port=2112 \
  -otlp-endpoint="otel:4317"
  -service-name="usdt-rate-service"
```

Запуск с переменными окружения:
```
$env:POSTGRES_DSN = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
$env:GRPC_PORT = "50051"
$env:METRICS_PORT = "2112"
$env:OTLP_ENDPOINT = "otel:4317"
$env:SERVICE_NAME = "usdt-rate-service"

go run ./cmd/app/main.go
```

## TODO:

- [x] Логирование с помощью `zap`
- [x] Использование миграций для создания схемы БД
- [x] Использование линтера `golangci-lint`
- [x] Комментарии в коде на английском языке
- [x] Трассировка запросов с помощью `OpenTelemetry`
- [x] Мониторинг с помощью `Prometheus`

---

- [ ] Интуитивно-понятное разбитие коммитов
