# dummy-be

Intentionally vulnerable Go backend for security classifier training. Exposes `/register`, `/login`, and `/home` endpoints with known vulnerabilities (SQL injection, broken auth, weak JWT).

## Prerequisites

- Docker
- PostgreSQL running on the host (port 5432)

## Setup

### 1. Configure environment

```bash
cp .env.example .env
```

Edit `.env` with your values:

```env
DB_HOST=
DB_PORT=
DB_NAME=
DB_USER=
DB_PASS=
JWT_SECRET=
PORT=8080
```

### 2. Initialize the database

```bash
psql -h 127.0.0.1 -U myapp_user -d myapp_db -f db/schema.sql
```

Or if postgres is running in Docker:

```bash
sudo docker exec -i postgres psql -U myapp_user -d myapp_db < db/schema.sql
```

### 3. Build the image

```bash
sudo docker build -t nexth-dummy-be:latest .
```

### 4. Run

```bash
sudo docker run -d \
  --name dummy-be \
  --network host \
  --env-file ~/nexth-dummy-be/.env \
  nexth-dummy-be:latest
```

`--network host` lets the container reach postgres on `127.0.0.1:5432` directly.

### 5. Verify

```bash
sudo docker logs dummy-be
curl http://localhost:8080/login \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

## Endpoints

| Method | Path      | Description                       |
| ------ | --------- | --------------------------------- |
| POST   | /register | Register a new user               |
| POST   | /login    | Login — returns JWT (SQLi vuln)   |
| GET    | /home     | Protected home — broken auth vuln |

## Logs

Every request is logged as a JSON line to `logs/requests.jsonl` (and stdout). The log format matches what the [agent-classifier](https://github.com/satryacode/agent-classifier) expects:

```json
{
  "level": "info",
  "timestamp": "2026-06-09T10:00:00.000Z",
  "msg": "request",
  "method": "POST",
  "path": "/login",
  "status": 200,
  "ip": "1.2.3.4",
  "body": "{\"username\":\"admin\"}",
  "response_time_ms": 12,
  "user_agent": "Mozilla/5.0"
}
```

To read logs from the container:

```bash
sudo docker exec dummy-be cat /root/logs/requests.jsonl
```

## Stopping

```bash
sudo docker stop dummy-be && sudo docker rm dummy-be
```
