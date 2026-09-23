# course go autism ЛР1

сервис создаёт поездку, отдаёт её по id и завершает. создание поездки атомарно пишет в `trips` и `trip_status_history`.

## требования

- го 1.24+
- утилита [`tripgoctl`](https://github.com/course-go-autumn-2026/course-infra)
- docker

## запуск

```bash
tripgoctl cluster start
tripgoctl environment start
tripgoctl connect

make migrate
make run
```

`environment start` создаёт `.env` с адресами окружения.  
запуск через `Makefile` инклюдит в окружение токены из `.env`

# что взял не из стека. сёд-пати

uuid от гугла взял

# ручки

префикс v1 — `/api/v1`
префикс домена сервиса — `trips`

утилитарные:

- GET `/health` просто пинг
- GET `/ready` пинг с чеком доступности бд

бизнесовые:

- POST `trips` — создает поездку
- GET `trips/{tripId}` инфа по поездке
- POST `trips/{tripId}/finish` завершение поездки

# решения

уровень изоляции — `ReadCommited`. тбх не вижу зачем тут другой уровень изоляции и какие хаки вообще тут с этим можно сделать.  
гонка запросов на поездку разрешается на уровне БД — индекс с unique констрейнтом.
гонка на finish тоже проработана на уровне БД — `update WHERE`

менеджер транзакций  
моя реализация открывает транзакцию, кладет в контекст, вызывает функцию. остальное делает либа из стека

# идемпотентность

ключ и хеш тела хранятся в `idempotency_keys` с `expires_at`. TTL — `IDEMPOTENCY_TTL`, по умолчанию 24 часа

## докер

```bash
docker build -f deploy/Dockerfile -t trip-service:lr1 .
```

запуск рядом с постгрей из `tripgoctl`:

```bash
docker run --rm --env-file .env --network host trip-service:lr1
```
