# DroneFleet

DroneFleet is a contract-first Go API for simulated drone fleet operations.

The project is organized around these decisions:

- OpenAPI spec first for the HTTP contract.
- `oapi-codegen` for generated API models and Chi route wiring.
- PostgreSQL for the production backend.
- `sqlc` for generated, type-safe database queries.
- In-memory store only for local demos and tests when `DATABASE_URL` is not set.

## Layout

```text
api/openapi.yaml        OpenAPI contract
db/schema.sql           PostgreSQL schema
db/queries/*.sql        sqlc query definitions
db/seed.sql             Local seed data
internal/api            Generated oapi-codegen package
internal/db             Generated sqlc package
internal/service        HTTP behavior implementing the OpenAPI server interface
internal/postgres       PostgreSQL repository implementation
internal/store          In-memory repository implementation for tests/local demos
internal/repository     Repository interface shared by both stores
cmd/server              Application entry point
```

## Generate

```sh
make generate
```

This regenerates:

- `internal/api/dronefleet.gen.go` from `api/openapi.yaml`
- `internal/db/*.go` from `db/schema.sql` and `db/queries/*.sql`

## Test

```sh
make test
```

## Run

Without `DATABASE_URL`, the app uses the in-memory store:

```sh
go run ./cmd/server
```

With PostgreSQL:

```sh
DATABASE_URL='postgres://user:pass@localhost:5432/dronefleet?sslmode=disable' go run ./cmd/server
```

## Docker Compose

Build and run the app plus PostgreSQL:

```sh
docker compose up --build
```

The API is available on `http://localhost:8080`, and PostgreSQL is available on `localhost:5432`.

The Postgres image runs `db/schema.sql` and `db/seed.sql` the first time the named volume is created.
