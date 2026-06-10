# dummy-be

Intentionally vulnerable Go backend for security classifier training. Exposes `/register`, `/login`, and `/home` endpoints with known vulnerabilities (SQL injection, broken auth, weak JWT).

## Intentional Vulnerabilities

| Vulnerability | Location | Description |
|---|---|---|
| SQL Injection | `POST /login` | Username/password interpolated directly into SQL query |
| Hardcoded admin | `POST /login` | `admin` / `admin123` always works as fallback |
| Plaintext password | DB + response | Passwords stored and returned in plaintext |
| Weak JWT secret | `POST /login` | JWT signed with configurable but weak secret |
| Broken auth | `GET /home` | Token not properly validated server-side |
| Error exposure | All endpoints | Raw DB errors returned to client |

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
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=myapp_db
DB_USER=myapp_user
DB_PASS=your_password
JWT_SECRET=weak-secret
PORT=8080
```

### 2. Initialize the database

```bash
PGPASSWORD=your_password psql -h 127.0.0.1 -U myapp_user -d myapp_db -f db/schema.sql
```

This creates the `users` table (with `blocked` column) and the `fraud_verdicts` table.

### 3. Build and run

```bash
sudo docker build -t nexth-dummy-be:latest .

sudo docker run -d \
  --name dummy-be \
  --network host \
  --env-file ~/nexth-dummy-be/.env \
  nexth-dummy-be:latest
```

`--network host` lets the container reach postgres on `127.0.0.1:5432` directly.

### 4. Verify

```bash
# Check logs
sudo docker logs dummy-be

# Clean login
curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# SQLi attempt
curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"'\'' OR 1=1--","password":"x"}'
```

## Endpoints

| Method | Path        | Description                        |
|--------|-------------|------------------------------------|
| POST   | `/register` | Register a new user                |
| POST   | `/login`    | Login — returns JWT (SQLi vuln)    |
| GET    | `/home`     | Protected home — broken auth vuln  |

## Logs

Every request is logged as a JSON line to **both stdout and `/root/logs/requests.jsonl`** inside the container. The log format is what [agent-classifier](https://github.com/satryacode/agent-classifier) expects:

```json
{
  "level": "info",
  "timestamp": "2026-06-09T10:00:00.000Z",
  "caller": "middleware/logger.go:49",
  "msg": "request",
  "method": "POST",
  "path": "/login",
  "status": 200,
  "ip": "127.0.0.1",
  "body": "{\"username\":\"admin\",\"password\":\"admin123\"}",
  "response_time_ms": 12,
  "user_agent": "curl/7.88.1"
}
```

Read live logs:

```bash
# Stream stdout (all logs)
sudo docker logs -f dummy-be

# Read the log file inside the container
sudo docker exec dummy-be cat /root/logs/requests.jsonl
```

> The agent-classifier bridges container stdout → a host-side file automatically via `start.sh`.

## DB Schema

```sql
-- Users table (with blocked flag set by agent-analyzer)
users (id, username, email, password, created_at, blocked)

-- Fraud verdicts written by agent-classifier, reviewed by agent-analyzer
fraud_verdicts (
    id, source_ip, user_identity, method, path,
    confidence_score, reason, original_log_entry_reference,
    detected_at, remediated
)
```

Blocked users get `403 Forbidden` on login.

## Stopping

```bash
sudo docker stop dummy-be && sudo docker rm dummy-be
```
