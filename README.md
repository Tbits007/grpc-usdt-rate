# Запуск приложения

## Стандартный способ (с Docker)

```bash
# Клонировать репозиторий
git clone https://github.com/Tbits007/grpc-usdt-rate.git
cd grpc-usdt-rate

# Собрать и запустить
docker-compose up --build -d
```

# gRPC API Документация

## Доступные методы

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
