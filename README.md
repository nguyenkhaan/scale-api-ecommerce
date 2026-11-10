# SaleScale

## Overview

SaleScale is a web application that models a flash-sale system. The backend focuses on correct multi-SKU checkout under concurrent load: PostgreSQL is the authoritative data store, while the API must preserve inventory, quotas, idempotency, and order lifecycle rules.

The repository currently contains a Go/Gin backend connected to Neon PostgreSQL through GORM and a React/Vite frontend. The backend includes health checks, request ID propagation, request logs that omit query strings and headers, and the initial Phase 00 test suite.

## Tech Badges

![Go 1.27.1](https://img.shields.io/badge/Go-1.27.1-00ADD8?logo=go&logoColor=white)
![Gin 1.12.0](https://img.shields.io/badge/Gin-1.12.0-008ECF?logo=gin&logoColor=white)
![GORM 1.31.2](https://img.shields.io/badge/GORM-1.31.2-00ADD8)
![PostgreSQL on Neon](https://img.shields.io/badge/PostgreSQL-Neon-4169E1?logo=postgresql&logoColor=white)
![React 19.3.1](https://img.shields.io/badge/React-19.3.1-61DAFB?logo=react&logoColor=black)
![Vite 8.3.4](https://img.shields.io/badge/Vite-8.3.4-646CFF?logo=vite&logoColor=white)

## Backend

### How to Setup

Prerequisites:

- Go `1.27.1` or later in the `1.27` release line.
- A Neon database and its connection string.
- Air for hot reload. Air is installed globally in the current development environment.

From the repository root:

```bash
cd app/backend
cp .env.example .env
go mod download
```

Update `.env` with the connection string from Neon:

```dotenv
DATABASE_URL=<NEON_CONNECTION_STRING>
PORT=4000
ENV=dev
```

Do not commit `.env` or write the connection string to logs. Start the API with hot reload:

```bash
air
```

Or run it directly with Go:

```bash
go run ./cmd/api
```

The API listens on `http://localhost:4000` by default. Check its health endpoints:

```bash
curl http://localhost:4000/api/health/liveness
curl http://localhost:4000/api/health/ready
```

Run the backend quality checks:

```bash
gofmt -w .
go test ./...
go test ./test -race -coverpkg=./...
go vet ./...
go build ./...
```

Apply or inspect migrations on the database referenced by `DATABASE_URL`:

```bash
go run ./cmd/migrate status
go run ./cmd/migrate up
```

Seed the complete development fixture only after migrations, and only against a disposable development or test database. The command is repeatable, but deliberately refreshes the fixture-owned values.

```bash
ALLOW_SEED=true go run ./cmd/seed
```

The fixture has two accounts: `admin@gmail.com` (admin) and `customer@gmail.com` (customer). Its password is development-only and is stored in PostgreSQL as a bcrypt hash, never plaintext.

To run the seed-command integration check against the database in `app/backend/.env`, opt in explicitly:

```bash
RUN_SEED_INTEGRATION=true go test -tags=integration ./test -run TestSeedCommandLoadsDevelopmentFixture
```

Reset is destructive and is restricted to an explicitly configured test environment. The command refuses to run unless both guards are present:

```bash
ENV=test \
ALLOW_DATABASE_RESET=true \
go run ./cmd/migrate reset
```

This project does not run PostgreSQL locally and does not require Docker Compose for Phase 00.
