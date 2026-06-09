# Dummy Backend Design — nexth-dummy-be

**Date:** 2026-06-09
**Purpose:** Deliberately vulnerable Go backend serving as a pentest target for attacker/defender agents.

---

## Overview

A realistic-looking REST API with intentional security vulnerabilities across three flows: register, login, and home. Logs all requests as structured JSON to AWS CloudWatch for downstream classification by a defender agent.

---

## Stack

| Component | Choice |
|-----------|--------|
| Language | Go |
| Framework | Gin |
| DB Driver | pgx (raw queries, no ORM) |
| Database | Aurora PostgreSQL 17 (AWS RDS) |
| Logging | zap → stdout → CloudWatch Logs |
| Deployment | EC2 (binary or Docker) |

---

## Infrastructure

| Resource | Value |
|----------|-------|
| RDS Host | `database-1.cluster-c0zimei407hh.us-east-1.rds.amazonaws.com` |
| DB Name | `postgres` |
| DB User | `postgres` |
| DB Password | `nexthack-2026` |
| Region | `us-east-1` |
| SSL | `sslmode=require` |
| CloudWatch Log Group | `/dummy-be/app` |
| App Port | `8080` |

---

## Project Structure

```
nexth-dummy-be/
├── main.go
├── config/
│   └── config.go          # Env var loading
├── db/
│   └── db.go              # pgx pool init (sslmode=require)
├── handlers/
│   ├── auth.go            # Register + Login
│   └── home.go            # Home
├── middleware/
│   └── logger.go          # Structured JSON request logger
├── models/
│   └── user.go            # User struct
├── .env.example
└── Dockerfile
```

---

## Database Schema

```sql
CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(255),
    email      VARCHAR(255),
    password   VARCHAR(255),   -- plaintext (intentional vuln)
    created_at TIMESTAMP DEFAULT NOW()
);

-- Seed default admin account
INSERT INTO users (username, email, password)
VALUES ('admin', 'admin@dummy.com', 'admin123');
```

---

## API Endpoints

### POST /register
- No input validation
- Stores password as plaintext
- Exposes raw DB error on conflict

**Request:**
```json
{ "username": "alice", "email": "alice@test.com", "password": "pass123" }
```
**Response 200:**
```json
{ "message": "user registered", "user_id": 1 }
```
**Response 500 (verbose DB error exposed):**
```json
{ "error": "pq: duplicate key value violates unique constraint \"users_email_key\"" }
```

### POST /login
- SQL query built via `fmt.Sprintf` (SQLi vuln)
- Returns full user object including plaintext password
- Hardcoded fallback: `admin` / `admin123`
- JWT signed with hardcoded secret `"secret"` (HS256)

**Request:**
```json
{ "username": "admin", "password": "admin123" }
```
**Response 200:**
```json
{
  "token": "eyJ...",
  "user": { "id": 1, "username": "admin", "email": "admin@dummy.com", "password": "admin123", "created_at": "..." }
}
```

### GET /home
- Accepts any Bearer token (signature not verified)
- Returns user data based on username claim in token (unverified)
- No rate limiting

**Header:** `Authorization: Bearer <anything>`
**Response 200:**
```json
{ "message": "welcome", "user": { "id": 1, "username": "alice", "email": "alice@test.com" } }
```

---

## Intentional Vulnerabilities

| Vuln | Location | Type |
|------|----------|------|
| SQL Injection | `POST /login` | `fmt.Sprintf` in raw query |
| Plaintext password storage | `POST /register`, DB | Broken crypto |
| Full user object in response | `POST /login` | Sensitive data exposure |
| JWT not verified | `GET /home` | Broken auth |
| Hardcoded JWT secret `"secret"` | `config/config.go` | Hardcoded credential |
| Default admin account | DB seed | Default credentials |
| Verbose DB errors | `POST /register` | Information disclosure |
| CORS `*` | Gin middleware | Misconfiguration |
| No rate limiting | All endpoints | Missing control |

---

## CloudWatch Log Format

Every request logged as structured JSON to stdout (CloudWatch agent picks up):

```json
{
  "timestamp": "2026-06-09T10:00:00Z",
  "method": "POST",
  "path": "/login",
  "status": 200,
  "ip": "1.2.3.4",
  "body": "{\"username\":\"admin\",\"password\":\"' OR 1=1--\"}",
  "response_time_ms": 12,
  "user_agent": "curl/7.88"
}
```

---

## EC2 IAM Role Requirements

- `logs:CreateLogGroup`
- `logs:CreateLogStream`
- `logs:PutLogEvents`
- `rds-db:connect` (if IAM auth ever needed)
