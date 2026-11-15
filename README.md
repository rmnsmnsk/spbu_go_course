# Movie Service

The service synchronizes missing movies from an external movie stand to PostgreSQL at startup. HTTP and gRPC requests read movie data only from PostgreSQL.

## Configuration

The service reads configuration from environment variables:

- `DATABASE_URL` — PostgreSQL DSN; required.
- `MOVIE_BASE_URL` — external movie stand URL; defaults to `https://homeworksite.site`.
- `HTTP_ADDR` — HTTP listen address; defaults to `:8080`.

The gRPC server listens on `:9090`.

## Run

```bash
DATABASE_URL='postgres://movie_service:movie_service@127.0.0.1:5432/movies?sslmode=disable' \
go run ./cmd/movie-service
```

At startup the service:

1. connects to PostgreSQL;
2. applies embedded SQL migrations;
3. finds the last movie available on the external stand;
4. downloads and saves only IDs missing from PostgreSQL;
5. starts the HTTP and gRPC servers.

Repeated synchronization does not create duplicate rows because `movies.id` is a primary key and writes use an upsert.

## HTTP API

- `GET /load-movies?from=1&to=10`
- `GET /last-movie`
- `GET /swagger/`

## Checks

```bash
go fmt ./...
go test ./...
go vet ./...
```
