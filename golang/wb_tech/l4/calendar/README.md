# HTTP Calendar Service

`calendar` — это HTTP-сервис для управления событиями календаря.
Он умеет создавать, обновлять, удалять и выдавать события по дню, неделе и месяцу, а также содержит обязательные фоновые компоненты из задания:

- воркер напоминаний через канал
- отдельную горутину архивирования старых событий
- асинхронный логгер для HTTP-слоя

## Что реализовано

### CRUD API

Сервис поддерживает:

- создание события
- обновление события
- удаление события
- получение событий на день
- получение событий на неделю
- получение событий на месяц
- получение архивных событий

### Напоминания через канал

При создании или обновлении события можно передать `remind_at`.
Если поле задано, сервис кладёт задачу в канал напоминаний.

Фоновый воркер:

1. читает задачу из канала
2. ждёт нужного времени
3. отправляет напоминание через `ReminderSender`
4. помечает событие как `reminder_sent=true`

В текущей реализации напоминание отправляется через `LogNotifier`, то есть факт отправки фиксируется в логах.

### Архивирование старых событий

Отдельная горутина по тикеру проверяет активные события и переносит устаревшие в архив.

Событие считается устаревшим, если:

```text
event.date < now - ARCHIVE_AFTER
```

По умолчанию:

- `ARCHIVE_INTERVAL=1m`
- `ARCHIVE_AFTER=24h`

Архивные события больше не попадают в обычные выборки по дню, неделе и месяцу, но доступны через отдельный endpoint.

### Асинхронный логгер

HTTP-хендлеры и middleware не пишут в `stdout` напрямую.
Они складывают записи в буферизированный канал, а отдельная горутина логгера пишет JSON-строки в `stdout`.

Логгер используется для:

- логов HTTP-запросов
- ошибок валидации
- ошибок бизнес-логики
- логов напоминаний
- логов архивирования
- событий запуска и остановки сервера

## DI

Сборка зависимостей вынесена в контейнер [internal/di/container.go](/D:/computer-science/golang/wb_tech/l4/calendar/internal/di/container.go:1).

Контейнер создаёт и связывает:

- асинхронный логгер
- in-memory репозиторий
- сервис календаря
- notifier для напоминаний
- Gin router
- `http.Server`


## Архитектура

```text
calendar/
├── cmd/calendar/main.go
├── internal/
│   ├── application/
│   │   ├── service.go
│   │   └── tests/service_test.go
│   ├── di/
│   │   └── container.go
│   ├── domain/
│   │   ├── entity/event.go
│   │   └── interfaces/
│   ├── infrastructure/
│   │   ├── logging/async.go
│   │   ├── reminder/notifier.go
│   │   └── repository/event.go
│   └── presentation/http/ginapp/
│       ├── dto/event.go
│       ├── handlers/calendar.go
│       ├── middleware/middleware.go
│       ├── routes.go
│       └── routes_test.go
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## Запуск

### Локально

```powershell
go run ./cmd/calendar
```

По умолчанию сервис слушает:

```text
http://localhost:8080
```

### Через Makefile

```powershell
make run
```

## Переменные окружения

- `PORT` — порт HTTP-сервера, по умолчанию `8080`
- `GIN_MODE` — режим Gin, по умолчанию `release`
- `ARCHIVE_INTERVAL` — как часто запускать архиватор, по умолчанию `1m`
- `ARCHIVE_AFTER` — через сколько времени событие считается старым, по умолчанию `24h`
- `SHUTDOWN_TIMEOUT` — таймаут graceful shutdown, по умолчанию `10s`

Пример:

```powershell
$env:PORT="8081"
$env:GIN_MODE="debug"
$env:ARCHIVE_INTERVAL="10s"
$env:ARCHIVE_AFTER="30m"
$env:SHUTDOWN_TIMEOUT="5s"
go run ./cmd/calendar
```

## Форматы даты и времени

Для `date` и `remind_at` поддерживаются форматы:

- `RFC3339`, например `2026-04-28T10:00:00Z`
- `2006-01-02T15:04`, например `2026-04-28T10:00`
- `2006-01-02`, например `2026-04-28`

Для запросов `events_for_day`, `events_for_week`, `events_for_month` обычно удобно передавать `YYYY-MM-DD`.

## Endpoints

### Хелфчек

`GET /healthz`

```bash
curl http://localhost:8080/healthz
```

### Создать событие

`POST /calendar/create_event`

```bash
curl -X POST http://localhost:8080/calendar/create_event \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "date": "2026-04-28T18:00:00Z",
    "event": "Встреча с командой",
    "remind_at": "2026-04-28T17:45:00Z"
  }'
```

### Обновить событие

`POST /calendar/update_event`

```bash
curl -X POST http://localhost:8080/calendar/update_event \
  -H "Content-Type: application/json" \
  -d '{
    "id": "event-id",
    "user_id": 1,
    "date": "2026-04-28T19:00:00Z",
    "event": "Обновлённая встреча",
    "remind_at": "2026-04-28T18:40:00Z"
  }'
```

### Удалить событие

`POST /calendar/delete_event`

```bash
curl -X POST http://localhost:8080/calendar/delete_event \
  -H "Content-Type: application/json" \
  -d '{
    "id": "event-id",
    "user_id": 1
  }'
```

### События на день

`GET /calendar/events_for_day?user_id=1&date=2026-04-28`

```bash
curl "http://localhost:8080/calendar/events_for_day?user_id=1&date=2026-04-28"
```

### События на неделю

`GET /calendar/events_for_week?user_id=1&date=2026-04-28`

```bash
curl "http://localhost:8080/calendar/events_for_week?user_id=1&date=2026-04-28"
```

### События на месяц

`GET /calendar/events_for_month?user_id=1&date=2026-04-28`

```bash
curl "http://localhost:8080/calendar/events_for_month?user_id=1&date=2026-04-28"
```

### Архивные события

`GET /calendar/archived_events?user_id=1`

```bash
curl "http://localhost:8080/calendar/archived_events?user_id=1"
```

## Формат ответа

Успешный ответ:

```json
{
  "result": {}
}
```

Ответ с ошибкой:

```json
{
  "error": "описание ошибки"
}
```

## Пример жизненного цикла события

1. Клиент создаёт событие с `date` и, при необходимости, `remind_at`.
2. Сервис сохраняет событие в активное хранилище.
3. Если указан `remind_at`, задача отправляется в канал напоминаний.
4. В нужный момент воркер отправляет напоминание и помечает событие как обработанное.
5. Когда событие становится старым, архиватор переносит его в архив.
6. После этого событие исчезает из обычных выборок и доступно только через `archived_events`.

## Тесты

В проекте есть:

- юнит-тесты CRUD-логики
- тест на отправку напоминания
- тест на перенос старых событий в архив
- HTTP-интеграционный тест на создание и чтение события

Запуск:

```powershell
go test ./...
```

Через Makefile:

```powershell
make test
```

## Docker

Сборка контейнера:

```powershell
docker build -t calendar-service .
```

Запуск контейнера:

```powershell
docker run --rm -p 8080:8080 calendar-service
```
