# Licence-Approval System

Данный проект представляет собой **полноценную демонстрацию** приложения по учёту и проверке лицензий с использованием следующих компонентов:

1. **Client** — клиентское приложение, которое проверяет состояние лицензии и при необходимости отправляет заявку на её получение.
2. **License Server** — сервер лицензирования, хранящий заявки и лицензии в базе данных (PostgreSQL).
3. **Mock OAuth2 Server** — фейковый сервер аутентификации/авторизации по OAuth2 (для тестовых целей), эмулирующий выдачу токенов.

Вся система может запускаться как в Docker (через `docker-compose`), так и локально (через `Makefile`).

---

### Главные каталоги

- **`client/`**: Исходники клиентского приложения, которое обращается к серверу лицензий (`server`) и читает/сохраняет конфиг (`config.json`).  
- **`mock-oauth-server/`**: Исходники и конфиги фейкового сервера OAuth2 (выдача токенов, авторизация).  
- **`server/`**: Исходники сервера лицензирования, работающего с базой PostgreSQL, а также папка `config/` с `.env`, ключами TLS и т.д.  
- **`docker-compose.yml`**: Описание контейнеров (PostgreSQL, mock-oauth, server, client) для запуска в Docker.  
- **`Makefile`**: Локальные команды для сборки и запуска без Docker.

---

## 2. Основной функционал

### 2.1. **Client**

- Хранит конфигурацию в `config.json` (поля `LICENSE_SERVER_URL`, `LICENSE_KEY`).  
- При старте проверяет существующий ключ. Если отсутствует, генерирует новый (на базе MachineID) и сохраняет в `config.json`.  
- Обращается к `server` по API (`/api/check-license`, `/api/create-license-request`) с помощью HTTPS.  
- Поддерживает проверку сертификата сервера (через RootCAs), но с `InsecureSkipVerify: true` (для теста).  
- Логика:
  1. Проверяет лицензии (если активна — завершается).
  2. При отсутствии лицензии — отправляет заявку на `/api/create-license-request`.
  3. Регулярно проверяет статус до одобрения/отклонения.

### 2.2. **License Server**

- Поднимается на порту `:8443`, держит свои сертификаты (`server.crt`, `server.key`).
- Реализует следующие эндпоинты:
  - `GET /api/check-license?license_key=...` — проверяет, есть ли лицензия, её статус или активная заявка.
  - `POST /api/create-license-request` — создаёт заявку на лицензию (статус `pending`).
- Хранит данные в PostgreSQL (таблицы `licenses`, `license_requests`).
- Имеет административный интерфейс:
  - `GET /admin/license-requests` — выводит все активные/неодобренные заявки (шаблон `admin_requests.html`).
  - `POST /admin/approve-license` — утверждает заявку (и создаёт лицензию).
  - `POST /admin/reject-license` — отклоняет заявку.
- Для аутентификации админ-маршрутов сервер использует **OAuth2** (через `mock-oauth`) и сессии (cookie) на базе `gorilla/sessions`.
- Файл `.env` хранит переменные окружения (см. ниже).

### 2.3. **Mock OAuth2 Server**

- Поднимается на порту `:8081`, держит свои сертификаты (`mock-oauth.crt`, `mock-oauth.key`).
- Предоставляет эндпоинты:
  - `/authorize` (GET/POST) — имитирует страницу логина, выдаёт `authorization_code`.
  - `/token` — обменивает `code` на `access_token`.
  - Дополнительные методы refresh token, introspection.
- Хранит данные в памяти (в модулях `internal/repository/inmem`).
- Локальная учётная запись: `username=admin`, `password=password`.

---

## 3. Запуск через Docker Compose

### 3.1. Предварительные действия

- Установить **Docker** и **docker-compose**.  
- В `server/.env` и `client/config.json` задать необходимые параметры, представленные в пункте 5 (см. ниже).


### 3.2. Команды

Из корня проекта:

```bash
docker-compose build
docker-compose up
```

**Docker поднимет**:
1. **db (PostgreSQL)** на порту `5432`.
2. **mock-oauth-server** на порту `8081` (проброшено наружу — [https://localhost:8081](https://localhost:8081)).
3. **license_server** на порту `8443` (проброшено наружу — [https://localhost:8443](https://localhost:8443)).
4. **client** — запускается фоново, внешний порт не нужен.

### 3.3. Подключение к базе данных

1. **`-U license_user`**  
   Указывает имя пользователя базы данных.  

2. **`-h db`**  
   Указывает хост, где запущен PostgreSQL (в Docker-среде это контейнер с именем `db`).  

3. **`-d license_db`**  
   Указывает название базы данных.  

Если подключаетесь к PostgreSQL **локально** (вне Docker-контейнера), то хост может быть `localhost`, а имя пользователя и базы данных должны соответствовать вашим настройкам (см. параметры в `.env`).  

**Пример команды**:  
```bash
psql -U license_user -h db -d license_db
```

### 3.4. Проверка работы

- **Client**: смотрите логи в консоли — он проверит лицензию, если надо, создаст заявку.  
- **Server**: доступен по адресу [https://localhost:8443](https://localhost:8443).  
- Откройте [https://localhost:8443/admin/license-requests](https://localhost:8443/admin/license-requests) в браузере. Проверьте отказ в доступе к административным ресурсам при отсутствии авторизации. 
- Необходимо авторизоваться через Mock OAuth. По адресу [https://localhost:8443/auth/login](https://localhost:8443/auth/login) введите `username=admin`, `password=password`
- После входа увидите список заявок на лицензию (если клиент её создал). Можно одобрить или отклонить.  
- **Mock OAuth**: [https://localhost:8081/authorize](https://localhost:8081/authorize) откроется страница авторизации.

### 3.5. Остановка

```bash
docker-compose down
```

---

# 4. Запуск локально (Makefile)

## 4.1. Настройки

- Необходим **Go (1.20+)**.
- Нужен локальный PostgreSQL (или меняем `DB_HOST=localhost` в `.env` и создаём пользователя/БД вручную).
- Установить `make` для использования Makefile.

## 4.2. Команды

```bash
make build            # Сборка всех сервисов (client, mock-oauth, server)
make run              # Сборка и одновременный запуск client, server, mock-oauth

make build-client
make run-client

make build-server
make run-server

make build-mock-oauth
make run-mock-oauth

make clean            # Удаление build-артефактов
```

При `make run` все три сервиса стартуют в одном терминале.  
Для остановки — **Ctrl+C**.  
Убедитесь, что в `.env` прописано `DB_HOST=localhost` и у вас локально запущен PostgreSQL.

---

# 5. Важные файлы конфигурации

## 5.1. `.env` (в директории `server/`)

```bash
# ===== База данных =====
DB_USER=license_user
DB_PASS=yourpassword
DB_NAME=license_db
DB_HOST=db
DB_PORT=5432
DB_SSLMODE=disable

# ===== OAuth2.0 =====
OAUTH_CLIENT_ID=000000
OAUTH_CLIENT_SECRET=999999
OAUTH_AUTH_URL=https://localhost:8081/authorize
OAUTH_TOKEN_URL=https://mock-oauth:8081/token # или https://localhost:8443 при локальном запуске
OAUTH_REDIRECT_URL=https://localhost:8443/oauth-cb

# ===== Сессия =====
SESSION_SECRET=1124394994 # Любая случайная строка 

# ===== Пути к ключам RSA =====
PRIVATE_KEY_PATH=config/keys/private_key.pem
PUBLIC_KEY_PATH=config/keys/public_key.pem

# ===== Пути к сертификатам для HTTPS =====
CERT_FILE=config/certs/server.crt
KEY_FILE=config/certs/server.key
```

> - `OAUTH_TOKEN_URL`: внутри Docker → `https://mock-oauth:8081/token`; снаружи → `https://localhost:8081/token`.  
> - `CERT_FILE` / `KEY_FILE`: самоподписанные TLS-сертификаты для лиценз-сервера.

---

## 5.2. `client/config.json`

```json
{
  "LICENSE_SERVER_URL": "https://server:8443", # или https://localhost:8443 при локальном запуске
  "LICENSE_KEY": ""
}
```

# 6. Основные сценарии использования

1. **Первый запуск**  
   - Поднимаете PostgreSQL + сервер лицензий + клиент.  
   - Клиент генерирует лицензионный ключ, отправляет запрос на `/api/check-license`.  
   - Сервер отвечает «нет лицензии», клиент создаёт заявку.  
   - Администратор заходит на `/admin/license-requests`, видит «pending»-заявку, нажимает «Одобрить», вводит **TAG**, и лицензия становится «active».  
   - Клиент периодически проверяет `/api/check-license`, видит «License is active. TAG: <число>» — завершается.

2. **Mock OAuth**  
   - Для защиты административной панели (`/admin/...`) сервер использует OAuth2.  
   - При переходе в браузере вас редиректит на `mock-oauth:8081/authorize` (или `localhost:8081/authorize`).  
   - Вводите `admin/password`, получаете `code`, затем происходит обмен → `access_token`, сервер создаёт cookie-сессию.  
   - Теперь можно управлять заявками.

3. **Обновление/модификация**  
   - Можно вносить изменения в код, сертификаты, `.env` и пересобирать.  
   - Включить или отключить `InsecureSkipVerify`.  
   - Перенастроить пути к ключам RSA (используются для подписей лицензий).
