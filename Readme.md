# Gallery

Simple Go & Bootstrap web gallery for ArtesaniaSory.

## Usage:

Examples for compilation or execution.

### Compilation

```sh
go build ./cmd/web/
```

### Execution

```sh
./web

```

### Flags

- Port: Specifie the port to listen

```sh
./web -port="4000"
```

- DSN: Specifie the PostgreSQL Database URL

```sh
./web -db-dsn="postgres://user:password@pg_server/db_name"
# by default this take the ENV Var DATABASE_URL
```

You can pass an OS ENV Var in the `-db-dns` flag like:

```sh
./web -db-dsn=$DATABASE_URL
```

### Features

- Custom middlewares for basic security, request logging and recovery from Panic!.
- Http router (httprouter).
- File Server for static files.
- Embed files to the binary (static and templates).
- Template cache to memory.
- PostgreSQL Database (pgx).
- SessionManager with session storage to PostgreSQL (scs).
- Flags usage for custom setup
- Middleware manager (alice).
- Custom loggers for info and error
- Database Migrations (migrate)
- Graceful Shutdown
- JSON logs
- Query Timeout Context on each DB Request
- DB Connection pool configuration

### Migrations

- Create a migration:

```sh
migrate create -seq -ext=.sql -dir=./migrations create_categories_table
```

- Execute UP migrations

NOTE: in both cases, the `sslmode` querystring is optional.

```sh
migrate -path=./migrations -database="postgres://user:pass@db_server/db_name?sslmode=disable" up
```

- Execute down migrations

```sh
migrate -path=./migrations -database="postgres://user:pass@db_server/db_name?sslmode=disable" down
```

- See migration version

```sh
migrate -path=./migrations -database="postgres://user:pass@db_server/db_name?sslmode=disable" version
```

- Go to specific version (migration 1 in this case)

```sh
migrate -path=./migrations -database="postgres://user:pass@db_server/db_name?sslmode=disable" goto 1
```
