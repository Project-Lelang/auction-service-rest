# Auction Service

REST API service for auction, bidding, payment, shipment, notification, and withdrawal flows. The service is written in Go with Gin and uses a clean architecture style.

## Requirements

- Go 1.23+
- MySQL 8+
- Redis 6+ for async jobs and notification queues
- SMTP account for production email delivery
- Optional: Firebase service account for FCM delivery
- Optional: Midtrans server key for Snap payment
- Optional: Biteship API key for shipping order and tracking
- Optional tools:
  - `air` for hot reload
  - `swag` for Swagger generation
  - `migrate` for creating migration files

## Quick Start

1. Install dependencies:

```bash
make tidy
```

2. Create local config:

```bash
cp conf.yml.example conf.yml
```

3. Edit `conf.yml`:

- Set `mysql` credentials.
- Set `redis` host, port, password, and database.
- Set `jwt.secret_key` to a strong secret.
- Set `super_admin.email` and `super_admin.password`.
- Set `fe_uri` and `cors_allowed_origins` to the frontend URL.
- Configure `email` for SMTP before production.
- Configure `midtrans`, `biteship`, and `firebase` when those integrations are enabled.

4. Run migrations:

```bash
make migrate
```

5. Seed the super admin:

```bash
make seed
```

6. Start the service:

```bash
make run
```

The API runs at `http://localhost:8080` by default.

Swagger UI is available at `http://localhost:8080/swagger/index.html`.

## Main Commands

| Command | Description |
| :--- | :--- |
| `make run` | Start the HTTP server |
| `make run.dev` | Start with hot reload using `air` |
| `make build` | Build binary to `bin/auction-service` |
| `make migrate` | Run all pending migrations |
| `make migrate-rollback steps=1` | Roll back migrations |
| `make seed` | Seed the super admin account |
| `make migration name=create_table_name` | Create a new migration pair |
| `make generate-swagger` | Regenerate Swagger docs |
| `make test.unit` | Run all tests with race detector |
| `make test.cover` | Run tests and generate coverage report |
| `make tidy` | Tidy Go modules |

## Email Setup

Email is required in production for registration OTP and forgot password OTP.

```yaml
email:
  enabled: true
  host: smtp.example.com
  port: 587
  username: smtp-user
  password: smtp-password
  from_address: no-reply@example.com
  from_name: Auction Service
```

In non-production, if email is disabled, OTP values are logged for local development. In production, missing email configuration returns an error.

## Auth Flows

Registration:

1. `POST /auth/request-otp` with `{ "email": "user@example.com" }`.
2. User receives a 6-digit OTP by email.
3. `POST /auth/register` with profile data, password, and OTP.

Forgot password:

1. `POST /auth/forgot-password` with `{ "email": "user@example.com" }`.
2. If the account exists, user receives a 6-digit reset code by email.
3. `POST /auth/reset-password` with `{ "email": "user@example.com", "otp": "123456", "password": "newPassword123" }`.

OTP codes expire after 5 minutes. Forgot password request always returns success to avoid account enumeration.

## Production Checklist

- Use `environment: production`.
- Set `debug: false`.
- Replace `jwt.secret_key` and `filesystem.presigned_url_secret_key`.
- Use production MySQL and Redis credentials.
- Set `cors_allowed_origins` to trusted frontend origins only.
- Enable and verify SMTP email delivery.
- Configure Firebase only with a valid service account path.
- Configure Midtrans with the correct production or sandbox server key.
- Configure Biteship when shipment order creation and tracking are enabled.
- Keep `conf.yml` out of source control.
- Run migrations before deploying application code.
- Run `make generate-swagger` after API contract changes.
- Run tests before release. On Windows, use a workspace-local Go cache if needed:

```powershell
$env:GOCACHE='D:\go\src\auction-service\.gocache'
go test ./...
```

## Project Structure

```text
.
|-- main.go
|-- cmd/
|   |-- migrate/
|   `-- seed/
|-- constant/
|-- database/seeder/
|-- delivery/
|   |-- api/
|   |-- dto_request/
|   |-- dto_response/
|   `-- middleware/
|-- docs/
|-- global/
|-- infrastructure/
|-- internal/
|   |-- i18n/
|   |-- jwt/
|   |-- notification/
|   `-- validator/
|-- manager/
|-- migration/
|-- model/
|-- repository/
|-- storage/i18n/language/
|-- use_case/
`-- util/
```
