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

The default test run does not connect to Neon. To run the database smoke test, set `TEST_DATABASE_URL` to the connection string of a dedicated Neon test database or branch, then run:

```bash
go test ./test -run '^TestNeonDatabaseConnection$' -v
```

Database migration, seed, and reset tooling has not been implemented in the current Phase 00 state, so the repository does not yet provide valid commands for those tasks.
