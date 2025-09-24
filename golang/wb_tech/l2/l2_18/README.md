

# HTTP Calendar 

HTTP-сервер для управления событиями календаря. 

---

## Структура проекта

```
.
├── cmd
│   └── calendar
│       └── main.go
├── Dockerfile
├── internal
│   ├── application
│   │   ├── service.go
│   │   └── tests
│   │       └── service_test.go
│   ├── domain
│   │   ├── entity
│   │   │   └── event.go
│   │   └── interfaces
│   │       ├── repository.go
│   │       └── service.go
│   ├── infrastructure
│   │   └── repository
│   │       └── event.go
│   └── presentation
│       └── http
│           └── ginapp
│               ├── dto
│               │   └── event.go
│               ├── handlers
│               │   └── calendar.go
│               ├── middleware
│               │   └── middleware.go
│               └── routes.go
├── Makefile
└── README.md
```

---

## Особенности

* **CRUD операции**: `create`, `update`, `delete` событий.
* **Выборка событий**: на день, неделю, месяц.
* **Middleware логирования**: метод, URL, статус, время обработки.
* **Graceful shutdown**: корректное завершение сервера при SIGINT/SIGTERM.
* **Unit-тесты**: покрытие основных функций бизнес-логики.
* **Конфигурация порта**: через переменную окружения `PORT` или флаг `-port`.

---

## Endpoints

Все эндпоинты находятся под `/calendar`.

| Метод | URL                 | Параметры/Body                                  | Описание           |
| ----- | ------------------- | ----------------------------------------------- | ------------------ |
| POST  | `/create_event`     | JSON или form: `user_id`, `date`, `event`       | Создание события   |
| POST  | `/update_event`     | JSON или form: `id`, `user_id`, `date`, `event` | Обновление события |
| POST  | `/delete_event`     | JSON или form: `id`, `user_id`                  | Удаление события   |
| GET   | `/events_for_day`   | Query: `user_id`, `date`                        | События на день    |
| GET   | `/events_for_week`  | Query: `user_id`, `date`                        | События на неделю  |
| GET   | `/events_for_month` | Query: `user_id`, `date`                        | События на месяц   |

**Формат даты:** `YYYY-MM-DD`
**Формат ответа:**

* Успешно:

```json
{"result": {...}}
```

* Ошибка:

```json
{"error": "описание ошибки"}
```

---

