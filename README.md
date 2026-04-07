# Auction Service (Go & Gin)

A RESTful auction service built with **Go** and the **Gin** framework, following clean architecture principles.

## Prerequisites

- **Go** 1.23+
- **PostgreSQL**
- **`air`** for hot reload (optional)
  ```bash
  go install github.com/air-verse/air@latest
  ```
- **`migrate`** for creating migration files (optional)
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- **`swag`** for regenerating swagger docs (optional)
  ```bash
  go install github.com/swaggo/swag/cmd/swag@v1.8.12
  ```

## Setup

**1. Install dependencies**
```bash
make tidy
```

**2. Configure the application**
```bash
cp conf.yml.example conf.yml
```
Edit `conf.yml` and set your database credentials and other settings.

**3. Run migrations**
```bash
make migrate
```

**4. Seed initial data** (creates the superadmin account)
```bash
make seed
```

**5. Start the server**
```bash
make run
```

The API is now available at `http://localhost:8080`.

Swagger UI: `http://localhost:8080/swagger/index.html`

## Commands

| Command | Description |
| :--- | :--- |
| `make run` | Start the HTTP server |
| `make run.dev` | Start with hot reload (requires `air`) |
| `make migrate` | Run database migrations |
| `make seed` | Seed superadmin account |
| `make migration name=<name>` | Create a new migration file |
| `make generate-swagger` | Regenerate swagger docs from annotations |
| `make test.unit` | Run unit tests |
| `make test.cover` | Run tests with coverage report |
| `make tidy` | Tidy Go modules |

## Project Structure

```
.
â”œâ”€â”€ main.go                        # Entry point & HTTP server
â”œâ”€â”€ cmd/
â”‚   â”œâ”€â”€ migrate/main.go            # Migration runner
â”‚   â””â”€â”€ seed/main.go               # Seeder runner
â”œâ”€â”€ conf.yml                       # Application config
â”œâ”€â”€ delivery/
â”‚   â”œâ”€â”€ api/                       # Route handlers & router setup
â”‚   â”œâ”€â”€ dto_request/               # Request DTOs
â”‚   â”œâ”€â”€ dto_response/              # Response DTOs
â”‚   â””â”€â”€ middleware/                # JWT, panic, translator, request ID
â”œâ”€â”€ use_case/                      # Business logic
â”œâ”€â”€ repository/                    # Database queries
â”œâ”€â”€ infrastructure/                # DB connection, migrations, logger
â”œâ”€â”€ manager/                       # Dependency injection container
â”œâ”€â”€ model/                         # Domain models & context helpers
â”œâ”€â”€ internal/
â”‚   â”œâ”€â”€ jwt/                       # JWT generation & parsing
â”‚   â”œâ”€â”€ i18n/                      # Internationalisation (en, id)
â”‚   â””â”€â”€ validator/                 # Request validation
â”œâ”€â”€ constant/                      # Sentinel errors & i18n keys
â”œâ”€â”€ global/                        # Config loader & getters
â”œâ”€â”€ util/                          # UUID, password helpers
â”œâ”€â”€ migration/                     # SQL migration files
â”œâ”€â”€ database/seeder/               # Database seeders
â”œâ”€â”€ storage/i18n/language/         # Translation YAML files (en, id)
â””â”€â”€ docs/                          # Generated swagger docs
```

## Authentication

Default superadmin credentials (set in `conf.yml`):
- **Email:** `superadmin@example.com`
- **Password:** `SuperAdmin@123`

All protected endpoints require a `Bearer` token in the `Authorization` header.
