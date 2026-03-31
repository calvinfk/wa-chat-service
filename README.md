# WA Chat Webhook

Go service for API processing with Fiber v3, Firestore activity logging, PostgreSQL connectivity utilities, and Google Cloud integrations (Storage + Firebase).

## Current API Surface

Base path: `/api`

- `GET /api/ping` - public health check.
- `GET /api/ping-protected` - protected health check.

Response format is standardized as:

```json
{
  "code": 200,
  "message": "...",
  "data": null,
  "errors": null
}
```

## Authentication Notes

- Access token middleware currently reads the token from cookie `access_token`.
- The cookie value is expected to be AES-encrypted (configured by `AES_ENCRYPTION_KEY`).
- Protected endpoints use middleware that checks `userID` and token parsing result from context.

## Project Structure

```text
.
|-- cmd/
|   `-- app/
|-- config/
|-- internal/
|   |-- app/
|   |-- dto/
|   |-- handler/
|   |   `-- http/
|   |       |-- middleware/
|   |       `-- v1/
|   |-- model/
|   |-- repository/
|   |   |-- firestore/
|   |   `-- postgres/
|   |-- service/
|   `-- usecase/
|-- pkg/
```

## Configuration

The application reads environment variables from `.env` (loaded by `godotenv` in development).

Required/important variables from `.env.example`:

- `APP_NAME`
- `APP_VERSION`
- `APP_ENVIRONMENT` (`development|production|staging`)
- `APP_PORT`
- `APP_URL`
- `DATABASE_URL`
- `AES_ENCRYPTION_KEY`
- `GCP_PROJECT_ID`
- `GCP_TASK_API_KEY`
- `GCP_ATTACHMENT_BUCKET`
- `GOOGLE_APPLICATION_CREDENTIALS`

Optional but commonly used:

- `CORS_ENABLED`, `CORS_ALLOW_*`
- `GCP_ATTACHMENT_LINK_EXPIRY`
- `GCP_ATTACHMENT_MAX_SIZE`
- `FIREBASE_CLIENT_EMAIL`, `FIREBASE_PRIVATE_KEY`

## Running Locally

1. Copy env file and fill values.

```bash
cp .env.example .env
```

Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

2. Run the app.

```bash
make run
```

Alternative:

```bash
go run ./cmd/app/main.go
```

## Running with Docker

Build and run with compose:

```bash
docker compose up --build
```

Or use Make target:

```bash
make docker
```

## Development Notes

- The app initializes PostgreSQL connection, Firebase app/client, and GCP storage client during bootstrap.
- Activity logs are persisted to Firestore via middleware on incoming requests.
- `internal/handler/http/v1/router.go` is currently a placeholder for future v1 endpoints.
