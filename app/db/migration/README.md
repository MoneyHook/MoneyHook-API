# Migration tests

The regular test suite validates schema and master-data definitions without a database:

```sh
go test ./...
```

Required master data is always reconciled. Development/demo sample data is inserted only when `ENABLE_SEED_DATA` is set to a value accepted as true by Go's `strconv.ParseBool` (for example `true`, `TRUE`, `True`, or `1`). An unset, blank, or false value disables it; an invalid value stops API startup.

When the API is started by Compose, startup initializes Firebase Auth first, provisions the fixed development Google user when `ENABLE_DEVELOPMENT_USER=true`, and then runs the database migration and seed. The development user and sample data use the same fixed Firebase UID (`a77a6e94-6aa2-47ea-87dd-129f580fb669`). The Auth user is provisioned only when both `ENABLE_DEVELOPMENT_USER` and `ENABLE_SEED_DATA` are true; both flags are disabled by default.

The integration suite creates and removes an isolated PostgreSQL schema and/or MySQL database on the configured test servers:

```sh
MIGRATION_TEST_POSTGRES_DSN='postgres://moneyhook:password@localhost:5432/moneyhook?sslmode=disable' \
MIGRATION_TEST_MYSQL_DSN='moneyhook:password@tcp(localhost:3306)/?parseTime=true' \
go test -tags=integration ./db/migration -v
```

Either DSN may be omitted to test only one dialect. The configured users must be allowed to create and drop schemas or databases. Never point these variables at credentials that should not have test DDL privileges.
