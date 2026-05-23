# Chirpy Server (Go)

A production-style, JSON HTTP backend for a Twitter-like microblogging service called **Chirpy**. Users can sign up, log in, post short messages ("chirps"), browse and filter them, and upgrade to a premium "Chirpy Red" tier via a third-party payment webhook.

The project is intentionally built on top of the Go standard library (`net/http`) with a small, carefully chosen set of dependencies. It is designed to be a clear, readable reference implementation of a modern Go web service, with real-world concerns covered end-to-end: authentication, password hashing, refresh tokens, database migrations, type-safe SQL, structured error responses, request authorization, and a webhook integration.

---

## Table of Contents

- [Why This Project](#why-this-project)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [API Reference](#api-reference)
- [Authentication Model](#authentication-model)
- [Security Notes](#security-notes)
- [Roadmap](#roadmap)

---

## Why This Project

If you are learning Go, building a backend portfolio piece, or evaluating idiomatic patterns for a small Go service, Chirpy is a useful reference because it shows:

- **The standard library is enough.** Routing is done with Go 1.22+ `http.ServeMux` method-and-path patterns. No router framework is required.
- **Real authentication, not toy auth.** Passwords are hashed with Argon2id (memory-hard, modern), access is issued as short-lived JWTs (1 hour), and long-lived refresh tokens are stored opaquely in the database so they can be revoked.
- **Type-safe database access.** SQL is written by hand in `sql/queries/` and compiled into Go via `sqlc`, eliminating an entire class of ORM-related runtime bugs.
- **Versioned schema.** Migrations live in `sql/schema/` in the Goose format, so the database is reproducible from zero.
- **Honest separation of concerns.** HTTP handlers, business validation, the authentication package, and the generated database layer are all in distinct packages.
- **Operational hooks.** The service exposes a readiness endpoint, a hit-counter admin page, and a guarded dev-only reset endpoint for integration testing.
- **A real third-party integration.** A `Polka` webhook upgrades users to the `is_chirpy_red` premium tier, authenticated with an API key header — a realistic shape for payment provider callbacks.

In short: it is small enough to read in one sitting, but covers most of the concerns you would actually face shipping a Go HTTP service.

---

## Features

- User registration, login, and profile update (email + password).
- Stateless JWT access tokens (HS256, 1 hour TTL) for API authorization.
- Opaque, server-stored refresh tokens (60 day TTL) with explicit revoke endpoint.
- Create, list, filter, sort, fetch by ID, and delete chirps (owner-only delete).
- Chirp validation: 140 character limit and profanity masking.
- Webhook endpoint for promoting users to "Chirpy Red" premium status, gated by an API key.
- Static file server mounted under `/app/` with a request counter middleware.
- Admin metrics page and a `dev`-only full reset endpoint for testing.
- Structured JSON error responses with consistent shape.

---

## Tech Stack

- **Language:** Go 1.26+
- **HTTP:** Standard library `net/http` with Go 1.22+ pattern-based routing.
- **Database:** PostgreSQL 13+ (uses `gen_random_uuid()`).
- **DB Driver:** [`lib/pq`](https://github.com/lib/pq)
- **Query layer:** [`sqlc`](https://sqlc.dev) (code generation from SQL).
- **Migrations:** [`goose`](https://github.com/pressly/goose) (annotations already present in `sql/schema/`).
- **Password hashing:** [`alexedwards/argon2id`](https://github.com/alexedwards/argon2id)
- **JWT:** [`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt)
- **UUIDs:** [`google/uuid`](https://github.com/google/uuid)
- **Config:** [`joho/godotenv`](https://github.com/joho/godotenv) (`.env` loading)

---

## Architecture

```
                ┌──────────────────────────────────┐
   HTTP ─────►  │  net/http ServeMux (main.go)     │
                │   - method+path routing          │
                │   - middlewareMetricsInc         │
                └────────────────┬─────────────────┘
                                 │
            ┌────────────────────┼────────────────────────┐
            │                    │                        │
       ┌────▼─────┐         ┌────▼─────┐            ┌─────▼─────┐
       │ handlers │         │ internal │            │  static   │
       │ (root    │◄────────┤ /auth    │            │  /app/    │
       │ package) │         │  JWT,    │            │ index.html│
       └────┬─────┘         │  Argon2, │            └───────────┘
            │               │  tokens  │
            │               └──────────┘
            │
       ┌────▼─────────────────────┐
       │ internal/database (sqlc) │
       └────┬─────────────────────┘
            │
       ┌────▼─────────┐
       │ PostgreSQL   │
       └──────────────┘
```

- HTTP handlers live in the root `main` package (`chirp.go`, `user.go`, `auth.go`, `polka.go`, `metrics.go`, `reset.go`, `readiness.go`).
- Cross-cutting helpers are in `json.go` (`respondWithJSON`, `respondWithError`).
- `internal/auth` is the only package allowed to touch JWTs, Argon2id, and bearer/API-key header parsing.
- `internal/database` is generated by `sqlc` from `sql/queries/*.sql` against the schema in `sql/schema/*.sql`. **Do not hand-edit files inside `internal/database/`.**

---

## Project Structure

```
chirpy-server-go/
├── assets/                  # Static assets (logo, etc.)
├── internal/
│   ├── auth/                # JWT, Argon2id, bearer & API key parsing, refresh tokens
│   │   ├── api_key.go
│   │   ├── hash.go
│   │   ├── jwt.go
│   │   ├── token.go
│   │   └── *_test.go
│   └── database/            # sqlc-generated query layer (DO NOT EDIT)
│       ├── chirps.sql.go
│       ├── db.go
│       ├── models.go
│       ├── refresh_tokens.sql.go
│       └── users.sql.go
├── sql/
│   ├── queries/             # Hand-written SQL, input for sqlc
│   │   ├── chirps.sql
│   │   ├── refresh_tokens.sql
│   │   └── users.sql
│   └── schema/              # Goose migrations
│       ├── 001_users.sql
│       ├── 002_chirps.sql
│       ├── 003_users_hashed_password.sql
│       ├── 004_refresh_tokens.sql
│       └── 005_users_is_chirpy_red.sql
├── auth.go                  # /api/login, /api/refresh, /api/revoke
├── chirp.go                 # /api/chirps* handlers and validation
├── user.go                  # /api/users (create, update)
├── polka.go                 # /api/polka/webhooks
├── metrics.go               # /admin/metrics + hit counter middleware
├── readiness.go             # /api/healthz
├── reset.go                 # /admin/reset (dev only)
├── json.go                  # JSON response helpers
├── main.go                  # Wiring, env loading, server bootstrap
├── index.html               # Tiny landing page served under /app/
├── go.mod / go.sum
└── sqlc.yaml
```

---

## Getting Started

### Prerequisites

- **Go** 1.26 or newer (see `go.mod`).
- **PostgreSQL** 13 or newer (for `gen_random_uuid()`).
- **Goose** for running migrations: <https://github.com/pressly/goose>
- **sqlc** *(only required if you change SQL files)*: <https://sqlc.dev>

Install the CLI tools:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 1. Clone the Repository

```bash
git clone https://github.com/farulivan/chirpy-server-go.git
cd chirpy-server-go
```

### 2. Configure Environment Variables

Copy the template and fill in real values:

```bash
cp .env.example .env
```

The `.env` file must define:

| Variable    | Required | Description                                                                                  |
| ----------- | -------- | -------------------------------------------------------------------------------------------- |
| `DB_URL`    | yes      | PostgreSQL connection string in `lib/pq` format.                                             |
| `PLATFORM`  | yes      | Set to `dev` to enable `POST /admin/reset`. Use anything else (e.g. `prod`) in production.   |
| `SECRET`    | yes      | High-entropy secret used to sign HS256 JWT access tokens. Generate with `openssl rand -hex 64`. |
| `POLKA_KEY` | yes      | API key expected from the Polka webhook in the `Authorization: ApiKey <POLKA_KEY>` header.   |

The server refuses to start if any of them is missing. See [`.env.example`](./.env.example) for the full template.

### 3. Provision the Database

```bash
createdb chirpy
```

(Or use any equivalent tool you prefer.)

### 4. Run Database Migrations

From the project root:

```bash
goose -dir sql/schema postgres "$DB_URL" up
```

To roll back the latest migration:

```bash
goose -dir sql/schema postgres "$DB_URL" down
```

### 5. Install Go Dependencies

```bash
go mod download
```

### 6. Run the Server

```bash
go run .
```

The server listens on `http://localhost:8080`. You should see:

```
Serving files from . on port 8080
```

Quick sanity check:

```bash
curl -i http://localhost:8080/api/healthz
```

---

## Development Workflow

### Regenerating Database Code

If you modify any file under `sql/queries/` or `sql/schema/`, regenerate the typed query layer:

```bash
sqlc generate
```

This rewrites `internal/database/*.sql.go` and `internal/database/models.go` according to `sqlc.yaml`.

### Running Tests

The `internal/auth` package ships with unit tests covering password hashing, JWT issuance/validation, and bearer/API-key parsing:

```bash
go test ./...
```

### Building a Production Binary

```bash
go build -o chirpy-server-go .
./chirpy-server-go
```

---

## API Reference

All request and response bodies are JSON unless otherwise noted. Errors share a single shape:

```json
{ "error": "human readable message" }
```

### Health & Static

| Method | Path           | Auth | Description                                            |
| ------ | -------------- | ---- | ------------------------------------------------------ |
| GET    | `/api/healthz` | none | Returns `200 OK` with body `OK` if the service is up.  |
| GET    | `/app/*`       | none | Serves static files (including `index.html`).          |

### Users

| Method | Path         | Auth   | Description                                |
| ------ | ------------ | ------ | ------------------------------------------ |
| POST   | `/api/users` | none   | Create a new user.                         |
| PUT    | `/api/users` | Bearer | Update the caller's email and password.    |

`POST /api/users` request:

```json
{ "email": "user@example.com", "password": "s3cret!" }
```

Returns `201 Created` with the new user (the password hash is never exposed):

```json
{
  "id": "uuid",
  "created_at": "2026-05-23T03:27:00Z",
  "updated_at": "2026-05-23T03:27:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

### Authentication

| Method | Path           | Auth                   | Description                              |
| ------ | -------------- | ---------------------- | ---------------------------------------- |
| POST   | `/api/login`   | none                   | Exchange email + password for tokens.    |
| POST   | `/api/refresh` | Bearer (refresh token) | Issue a new access token.                |
| POST   | `/api/revoke`  | Bearer (refresh token) | Revoke a refresh token. Returns `204`.   |

`POST /api/login` response:

```json
{
  "id": "uuid",
  "created_at": "...",
  "updated_at": "...",
  "email": "user@example.com",
  "is_chirpy_red": false,
  "token": "<JWT access token, 1h>",
  "refresh_token": "<opaque hex string, 60d>"
}
```

### Chirps

| Method | Path                    | Auth   | Description                                                                          |
| ------ | ----------------------- | ------ | ------------------------------------------------------------------------------------ |
| POST   | `/api/chirps`           | Bearer | Create a chirp. Rejects bodies over 140 chars. Profanity is masked.                  |
| GET    | `/api/chirps`           | none   | List chirps. Supports `?author_id=<uuid>` and `?sort=asc\|desc` (default `asc`).      |
| GET    | `/api/chirps/{chirpID}` | none   | Fetch one chirp by ID. Returns `404` if missing.                                     |
| DELETE | `/api/chirps/{chirpID}` | Bearer | Delete a chirp. `403` if the caller is not the author. Returns `204` on success.     |

`POST /api/chirps` body:

```json
{ "body": "Hello world!" }
```

Profanity rule: any whitespace-separated token equal (case-insensitive) to `kerfuffle`, `sharbert`, or `fornax` is replaced with `****`.

### Webhooks

| Method | Path                  | Auth                                 | Description                                                |
| ------ | --------------------- | ------------------------------------ | ---------------------------------------------------------- |
| POST   | `/api/polka/webhooks` | `Authorization: ApiKey <POLKA_KEY>`  | Promotes a user to Chirpy Red on `user.upgraded` events.   |

Body:

```json
{
  "event": "user.upgraded",
  "data": { "user_id": "uuid" }
}
```

Responses:

- `204 No Content` on success or when the event is not `user.upgraded`.
- `401 Unauthorized` if the API key is missing or incorrect.
- `404 Not Found` if the user does not exist.

### Admin

| Method | Path              | Auth | Description                                                                                              |
| ------ | ----------------- | ---- | -------------------------------------------------------------------------------------------------------- |
| GET    | `/admin/metrics`  | none | HTML page showing how many times the `/app/*` static handler has been hit since startup.                 |
| POST   | `/admin/reset`    | none | **Destructive.** Deletes all users (chirps cascade) and resets the hit counter. Only enabled when `PLATFORM=dev`. |

---

## Authentication Model

1. The client calls `POST /api/login` with email and password.
2. The server verifies the password with Argon2id and returns:
   - A short-lived **access token** (HS256 JWT, issuer `chirpy-access`, 1 hour).
   - A long-lived **refresh token** (32 random bytes, hex-encoded, 60 days), persisted in the `refresh_tokens` table.
3. Authenticated endpoints expect `Authorization: Bearer <access token>`.
4. When the access token expires, the client calls `POST /api/refresh` with the refresh token in the `Authorization: Bearer` header to receive a new access token. The query rejects revoked or expired refresh tokens.
5. `POST /api/revoke` marks a refresh token as revoked so it can no longer be exchanged.
6. The webhook endpoint uses a different scheme: `Authorization: ApiKey <POLKA_KEY>`.

---

## Security Notes

- **Use a strong `SECRET`.** The JWT signing key must be high-entropy and kept out of source control. Treat `.env` as secret.
- **Never set `PLATFORM=dev` in production.** It exposes `POST /admin/reset`, which deletes every user.
- **Argon2id parameters** use `argon2id.DefaultParams`. Tune them for your hardware if needed.
- **Refresh tokens are stored as plaintext** in the `refresh_tokens` table today. They are random 256-bit values, but for defense-in-depth they could be hashed before storage (see Roadmap).
- **TLS termination** is expected to happen at a reverse proxy (Caddy, nginx, a load balancer, etc.). The Go server listens on plain HTTP on port 8080.
- **Rate limiting** is not implemented at the application layer. Put the service behind a rate-limiting proxy if it is exposed to the public internet.

---

## Roadmap

The following items are planned or considered to evolve Chirpy from a learning project into a more production-hardened service. Contributions and proposals are welcome.

### Short term

- **Add a `Makefile` / `Taskfile`** with targets for `run`, `test`, `migrate-up`, `migrate-down`, `sqlc`, and `lint` to standardize developer workflows.
- **Containerize the service** with a multi-stage `Dockerfile` and a `docker-compose.yml` that brings up Postgres + the API for one-command local development.
- **CI pipeline** (GitHub Actions) running `go vet`, `go test ./...`, `staticcheck`, and `sqlc diff` on every pull request.
- **Structured logging** with `log/slog` instead of the default `log` package, including request IDs and latency.
- **Graceful shutdown** on `SIGINT` / `SIGTERM` with `http.Server.Shutdown` and a context-aware `db.Close`.

### Medium term

- **Pagination** for `GET /api/chirps` (e.g., `?limit=` and `?cursor=`) so the endpoint scales beyond a few hundred rows.
- **Hash refresh tokens at rest** (e.g., SHA-256) so a database leak does not immediately compromise live sessions.
- **Email verification flow** for newly created users, plus password reset via signed, one-time tokens.
- **Per-user rate limiting and abuse protection** for chirp creation and login attempts.
- **OpenAPI / Swagger spec** generated from the handlers, plus a hosted docs UI.
- **Integration tests** that spin up Postgres in a container (e.g., `testcontainers-go`) and exercise every endpoint end-to-end.
- **Migrate from `lib/pq` to `pgx`** for better performance, prepared-statement caching, and richer type support.

### Long term

- **WebSocket or SSE feed** for real-time chirp delivery to followers.
- **Follow / timeline model** so users see only chirps from accounts they follow.
- **Media attachments** (images, short clips) backed by S3-compatible object storage.
- **Multi-tenant or organization support** with role-based access control.
- **Observability stack:** Prometheus metrics endpoint, OpenTelemetry traces, and a Grafana dashboard.
- **Replace the in-process counter** at `/admin/metrics` with a real metrics pipeline that survives restarts and works across replicas.
- **Horizontal scaling readiness:** stateless replicas behind a load balancer, with refresh-token storage and rate limiting backed by a shared store (e.g., Redis).

---

## License

This project is licensed under the **MIT License**. See the [`LICENSE`](./LICENSE) file for the full text.

In short: you may use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the software, provided that the copyright notice and the license notice are included in all copies or substantial portions of the Software. The software is provided "as is", without warranty of any kind.

