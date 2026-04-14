# WA Chat Service

Go service built with Fiber v3 for WhatsApp chat operations, media handling, and Firestore-backed data access.

## API Base URL

All routes are mounted under:

`/api`

## API Endpoints

### Health

- `GET /api/ping` - public health check
- `GET /api/ping-protected` - protected health check (see Authentication Notes)

### Auth (`/api/v1/auth`)

- `POST /login`
  - Body (JSON): auth login payload (handled by `dto.AuthLoginRequest`)
  - On success, sets encrypted token cookie `access_token` (HttpOnly)

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
- `GET /by-phone-number-id`
  - Query: `phone_number_id`
  - Optional query: `page`, `page_size`, `sort_by`, `sort_order`
- `GET /messages`
  - Query: `chat_id`
  - Optional query: `page`, `page_size`, `sort_by`, `sort_order`

### Template (`/api/v1/template`)

- `GET /get`
  - Query: `phone_number_id`
  - Optional query: `page`, `page_size`, `sort_by`, `sort_order`
- `POST /create`
- `POST /sync`
- `PUT /update`
- `DELETE /delete`

### Broadcast (`/api/v1/broadcast`)

- `POST /schedule`
- `POST /send`
  - Accepts encrypted bearer token in `Authorization` header (parsed via JWT middleware)
  - Endpoint returns `200` immediately and only continues async send flow when a valid `jwt_sub` is present

### Tenant Contact (`/api/v1/tenant/contact`)

- `POST /create`
- `GET /filter`
  - Supports filter/pagination query params
- `PUT /update`

### Storage Media (`/api/v1/storage-media`)

- `POST /upload`
  - Multipart form-data: `file` (single file), `phone_number_id`
- `GET /get`
  - Query: `id`
  - Returns streamed media bytes (`Content-Type` from stored object metadata)
- `DELETE /delete`
  - Query: `phone_number_id` and one of `id` or `media_id`
- `POST /save-media-id`
  - Body (JSON): `media_id`, `phone_number_id`
- `POST /resumable`
  - Body (JSON): resumable upload payload
- `POST /upload-meta`
  - Body (JSON): media metadata payload
- `GET /list`
  - Optional query: filter/pagination params

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
- `AccessToken` middleware is enabled globally in router wiring.
- The protected guard (`/api/ping-protected` and most `/api/v1/*` routes) checks:
  - `token_error_message` request header (if provided)
  - `token_sub` context value
- If `token_sub` is missing, protected routes return `401 Unauthorized`.

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
- `GCP_APP_BASE_URL`
- `GCP_TASK_BROADCAST_PARENT`
- `JOSE_RSA_PRIVATE_KEY` (path to RSA private key file)
- `JOSE_ACCESS_TOKEN_EXPIRY` (Go duration format, for example `24h`)
- `CORS_ENABLED`

Optional/common variables:

- `APP_SECURE_COOKIE`
- `CORS_ALLOW_ORIGINS`
- `CORS_ALLOW_METHODS`
- `CORS_ALLOW_HEADERS`
- `CORS_EXPOSE_HEADERS`
- `CORS_ALLOW_CREDENTIALS`
- `META_APP_ID`
- `META_GRAPH_API_VERSION`
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
- Google Cloud Storage client
- Google Cloud Tasks client

PostgreSQL connection bootstrap is currently not active in wiring (the connection code is present but commented out).

## Notes
### Template
-  When creating templates, if there's no parameter in the components, the `parameter_format` can be filled or left null, both are valid. If there's parameter(s) in the components, the `parameter_format` must be filled with either "NAMED" or "POSITIONAL".

### Broadcast
- Currently does not support Button for multi product message, one time password, voice call, single product message

### Media
- Resumable api is for creating template and profile picture.
