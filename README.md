# database
Database utilities for Nagomi.

## Prequesties
- PostgreSQL 18+
- Go 1.26.0+
- [sqlc](https://github.com/sqlc-dev/sqlc)
- [goose](https://github.com/pressly/goose)

## Development
### Internal Methods
sqlc is used to generate internal methods used for the database.

### Migrations
goose is used to generate and run migrations for the database. Please refer to [goose's documentation](https://pressly.github.io/goose/documentation/annotations) for more information on how it works.
#### Creating New Migration
```bash
goose -dir ./sql/migrations create init sql
```
#### Running `up` Migrations
> [!NOTE]
> Both `GOOSE_DRIVER` and `GOOSE_DBSTRING` must be set when running this command.
> ```env
> GOOSE_DRIVER=postgres
> GOOSE_DBSTRING=postgres://admin:admin@localhost:5432/admin_db
> ```
```bash
goose -dir ./sql/migrations up
```
