# go-analyse

`go-analyse` — это небольшой HTTP-сервер на Go с `gin`, который отдает текущие данные о памяти и сборщике мусора в формате Prometheus через официальный клиент `prometheus/client_golang`, а также открывает стандартные `pprof`-endpoints для профилирования.

В реализации используются:

- `gin` для HTTP-роутинга
- `prometheus/client_golang` для экспорта `/metrics`
- `runtime.ReadMemStats` для чтения состояния памяти и GC
- `debug.SetGCPercent` для настройки агрессивности GC
- `net/http/pprof` для CPU/heap/goroutine-профилирования

## Что показывает сервер

На endpoint `GET /metrics` экспортируются две группы метрик:

- встроенные метрики официального Go collector:
  `go_memstats_alloc_bytes`, `go_memstats_mallocs_total`, `go_memstats_num_gc`, `go_memstats_last_gc_time_seconds`, `go_goroutines`
- пользовательские метрики этого приложения на основе `runtime.ReadMemStats`:
  `go_analyse_process_uptime_seconds`, `go_analyse_gc_configured_percent`, `go_analyse_gc_last_run_age_seconds`, `go_analyse_gc_last_run_timestamp_seconds`

## Структура проекта

```text
go-analyse/
├── cmd/go-analyse/main.go
├── internal/config/config.go
├── internal/metrics/collector.go
├── internal/server/http.go
├── Dockerfile
├── go.mod
└── README.md
```

## Запуск

### Локально

```powershell
go run ./cmd/go-analyse
```

По умолчанию сервер слушает:

```text
http://localhost:8080
```

### С переменными окружения

```powershell
$env:PORT="8090"
$env:GC_PERCENT="50"
$env:GIN_MODE="debug"
$env:SHUTDOWN_TIMEOUT="5s"
$env:READ_HEADER_TIMEOUT="2s"
go run ./cmd/go-analyse
```

## Переменные окружения

- `PORT` — HTTP-порт, по умолчанию `8080`
- `GC_PERCENT` — значение для `debug.SetGCPercent`, по умолчанию `100`
- `GIN_MODE` — режим `gin` (`debug`, `release`, `test`), по умолчанию `release`
- `SHUTDOWN_TIMEOUT` — таймаут graceful shutdown, по умолчанию `10s`
- `READ_HEADER_TIMEOUT` — `http.Server.ReadHeaderTimeout`, по умолчанию `5s`

## Endpoints

### Проверка работоспособности

```powershell
curl http://localhost:8080/healthz
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

### Метрики Prometheus

```powershell
curl http://localhost:8080/metrics
```

Пример фрагмента ответа:

```text
# HELP go_memstats_mallocs_total Cumulative count of heap objects allocated.
# TYPE go_memstats_mallocs_total counter
go_memstats_mallocs_total 1234
# HELP go_analyse_gc_last_run_age_seconds Seconds since the last completed GC cycle, calculated from runtime.ReadMemStats.
# TYPE go_analyse_gc_last_run_age_seconds gauge
go_analyse_gc_last_run_age_seconds 2.137
```

### Pprof index

```powershell
curl http://localhost:8080/debug/pprof/
```

## Docker

Сборка образа:

```powershell
docker build -t go-analyse .
```

Запуск контейнера:

```powershell
docker run --rm -p 8080:8080 go-analyse
```

Запуск контейнера с более агрессивным GC:

```powershell
docker run --rm -p 8080:8080 -e GC_PERCENT=50 -e GIN_MODE=release go-analyse
```

### Heap profile

```powershell
go tool pprof http://localhost:8080/debug/pprof/heap
```

### CPU profile на 5 секунд

```powershell
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=5
```

### Goroutine dump в текстовом виде

```powershell
curl "http://localhost:8080/debug/pprof/goroutine?debug=1"
```

## Проверка тестов

```powershell
go test ./...
```

## Как подключить к Prometheus

Пример job-конфигурации:

```yaml
scrape_configs:
  - job_name: go-analyse
    static_configs:
      - targets:
          - localhost:8080
```
