# WA Chat Service

Go service built with Fiber v3 for WhatsApp chat operations, media handling, and Firestore-backed data access.

## API Base URL

All routes are mounted under:

`/api`

## API Endpoints

### Health

- `GET /api/ping` - public health check
- `GET /api/ping-protected` - protected health check (see Authentication Notes)

### Chat (`/api/v1/chat`)

- `POST /send`
  - Body (JSON): `phone_number_id`, `recipient_id`, `recipient_name`, `sender_name`, `type`, and payload keyed by `type`
  - Example for text message payload:
    ```json
    {
      "phone_number_id": "123",
      "recipient_id": "628123456789",
      "recipient_name": "Recipient",
      "sender_name": "Agent",
      "type": "text",
      "text": {
        "body": "Hello"
      }
    }
    ```
- `GET /template-list`
  - Query: `phone_number_id`
- `GET /by-phone-number-id`
  - Query: `phone_number_id`
  - Optional query: `page`, `page_size`, `sort_by`, `sort_order`
- `GET /messages`
  - Query: `chat_id`
  - Optional query: `page`, `page_size`, `sort_by`, `sort_order`

### Storage Media (`/api/v1/storage-media`)

- `POST /upload`
  - Multipart form-data: `file` (single file), `phone_number_id`
- `GET /get`
  - Query: `id`
  - Returns streamed media bytes (`Content-Type` from stored object metadata)
- `DELETE /delete`
  - Query: `phone_number_id` and one of `id` or `media_id`
- `POST /upload-media-id`
  - Body (JSON): `media_id`, `phone_number_id`

## Standard JSON Response Shape

Most JSON endpoints return:

```json
{
  "code": 200,
  "message": "...",
  "data": null,
  "errors": null
}
```

Notes:

- Validation and business errors are normalized into `errors` and `message`.
- `GET /api/v1/storage-media/get` is a file stream endpoint and does not return this JSON envelope on success.

## Authentication Notes

- Token parsing middleware exists and expects an AES-encrypted token in cookie `access_token`.
- The protected guard (`/api/ping-protected`) checks `jwt_error_message` and `userID` from request context.
- In current router wiring, `AccessToken` middleware is commented out, so `/api/ping-protected` will return `401 Unauthorized` until token middleware is enabled for incoming requests.

## Global Middleware

- Request logger
- Panic recovery
- CORS (config-driven)
- OPTIONS preflight short-circuit
- Request body/file-size guard at 16 MB (`413` when exceeded)

## Configuration

The app loads `.env` via `godotenv` in development and then builds config from environment variables.

Required variables:

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
- `CORS_ENABLED`

Optional/common variables:

- `APP_SECURE_COOKIE`
- `CORS_ALLOW_ORIGINS`
- `CORS_ALLOW_METHODS`
- `CORS_ALLOW_HEADERS`
- `CORS_EXPOSE_HEADERS`
- `CORS_ALLOW_CREDENTIALS`
- `GCP_ATTACHMENT_LINK_EXPIRY` (seconds, `-1` to disable)
- `GCP_ATTACHMENT_MAX_SIZE` (bytes)
- `GOOGLE_APPLICATION_CREDENTIALS` (recommended for local/dev if using ADC)

## Running Locally

1. Create `.env` from `.env.example` and fill required values.

```powershell
Copy-Item .env.example .env
```

2. Start the app.

```bash
make run
```

Alternative:

```bash
go run ./cmd/app/main.go
```

## Running with Docker

```bash
docker compose up --build
```

Or:

```bash
make docker
```

## Bootstrap Behavior

At startup, the service initializes:

- Firebase app
- Firestore client
- Firebase Messaging client
- Firebase Storage client
- Google Cloud Storage client

PostgreSQL connection bootstrap is currently not active in `internal/app/app.go` (the connection code is present but commented out).
