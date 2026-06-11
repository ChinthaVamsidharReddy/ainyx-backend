# Backend Task — User with DOB & Calculated Age

A RESTful API built with **Go + GoFiber + PostgreSQL + SQLC** that manages users and dynamically calculates their age on every fetch.

---

## Tech Stack

| Layer | Library |
|-------|---------|
| HTTP Framework | [GoFiber v2](https://gofiber.io/) |
| DB Access | [SQLC](https://sqlc.dev/) (type-safe generated code) |
| Database | PostgreSQL 16 |
| Logging | [Uber Zap](https://github.com/uber-go/zap) |
| Validation | [go-playground/validator v10](https://github.com/go-playground/validator) |
| Driver | [lib/pq](https://github.com/lib/pq) |

---

## Project Structure

```
.
├── cmd/server/main.go          # Entry point – wires all layers and starts Fiber
├── config/config.go            # Reads env vars, builds DSN
├── db/
│   ├── migrations/             # Raw SQL migrations (up + down)
│   └── sqlc/                   # SQLC-generated Go DB layer (models, querier, queries)
├── internal/
│   ├── handler/                # HTTP layer – parses request, delegates to service
│   ├── middleware/             # RequestID injection + request-duration logger
│   ├── models/                 # Request/response DTOs
│   ├── repository/             # Thin wrapper around SQLC Queries
│   ├── routes/                 # Route registration
│   └── service/                # Business logic + age calculation
├── docker-compose.yml
├── Dockerfile
├── sqlc.yaml
└── .env.example
```

---

## Running Locally (Docker – Recommended)

```bash
# 1. Clone the repo
git clone https://github.com/ChinthaVamsidharReddy/ainyx-backend
cd ainyx-backend

# 2. Start Postgres + API in one command
docker compose up --build

# Server is available at http://localhost:8080
```

Docker Compose automatically:
- Starts a Postgres 16 container
- Runs the migration (`db/migrations/000001_create_users_table.up.sql`) on first boot
- Builds and starts the Go API

---

## Running Locally (Without Docker)

### Prerequisites
- Go 1.22+
- PostgreSQL running locally

```bash
# 1. Create the database
psql -U postgres -c "CREATE DATABASE ainyx;"

# 2. Run the migration
psql -U postgres -d ainyx -f db/migrations/000001_create_users_table.up.sql

# 3. Configure environment
cp .env.example .env
# Edit .env and set your DB_PASSWORD

# 4. Run
go run ./cmd/server/main.go
```

---

## API Reference

All endpoints are under `/users`. Errors are returned as `{"error": "message"}`.

### POST /users — Create a user

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "dob": "1990-05-10"}'
```

Response `201 Created`:
```json
{"id": 1, "name": "Alice", "dob": "1990-05-10"}
```

---

### GET /users/:id — Get user by ID (includes `age`)

```bash
curl http://localhost:8080/users/1
```

Response `200 OK`:
```json
{"id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35}
```

---

### PUT /users/:id — Update a user

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Updated", "dob": "1991-03-15"}'
```

Response `200 OK`:
```json
{"id": 1, "name": "Alice Updated", "dob": "1991-03-15"}
```

---

### DELETE /users/:id — Delete a user

```bash
curl -X DELETE http://localhost:8080/users/1
```

Response `204 No Content`

---

### GET /users — List all users (with optional pagination)

```bash
# All users (no pagination wrapper)
curl http://localhost:8080/users

# Paginated
curl "http://localhost:8080/users?page=1&page_size=5"
```

Response (paginated):
```json
{
  "data": [{"id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35}],
  "total": 1,
  "page": 1,
  "page_size": 5,
  "total_pages": 1
}
```

---

## Running Unit Tests

```bash
go test ./internal/service/... -v
```

Tests cover the `CalculateAge` function with cases for:
- Birthday already passed this year
- Birthday is today
- Birthday not yet this year

---

## Design Decisions

### Age Calculation
Age is computed at request time in Go — **not stored in the database**. The `CalculateAge` function in `internal/service/user_service.go` subtracts birth year from current year, then decrements by 1 if today is before the birthday (month/day comparison). This handles the edge case where someone born on Dec 31 is not yet that age until their birthday.

### SQLC (no ORM)
SQLC generates type-safe Go code from raw SQL. This means zero reflection at runtime, compile-time guarantees that query params match, and easy auditability — you read actual SQL, not ORM magic.

### Repository Pattern
The `repository` package is a thin adapter between the service layer and SQLC. Keeping SQLC types out of the service/handler packages means the DB layer can be swapped (e.g., for a mock in tests) without touching business logic.

### Middleware
- **RequestID** — Reads `X-Request-ID` from the incoming request or mints a new UUID. Injects it into the response header and into `c.Locals` so the logger can include it.
- **Logger** — Uses Zap to emit structured JSON logs for every request: method, path, status, latency, request_id, client IP.

---

## Environment Variables

| Variable | Default | Required |
|----------|---------|----------|
| `DB_HOST` | `localhost` | No |
| `DB_PORT` | `5432` | No |
| `DB_USER` | `postgres` | No |
| `DB_PASSWORD` | — | Yes |
| `DB_NAME` | `ainyx` | No |
| `SERVER_PORT` | `8080` | No |
