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

- `POST /upsert`
  - Body (JSON): Broadcast configuration data
- `POST /schedule`
  - Body (JSON): Broadcast scheduling data
- `POST /send`
  - Expects JWT token in `Authorization` header (JWT middleware validation)
  - Endpoint returns `200` immediately and continues async send flow with valid `jwt_sub`
- `PUT /cancel`
  - Body (JSON): Broadcast ID to cancel
- `GET /get-filtered`
  - Supports filter/pagination query params
- `GET /get-recipients-filtered`
  - Supports filter/pagination query params for broadcast recipients

### Tenant Contact (`/api/v1/tenant/contact`)

- `POST /create`
- `GET /filter`
  - Supports filter/pagination query params
- `PUT /update`

### Storage Media (`/api/v1/storage-media`)

- `POST /upload`
  - Multipart form-data: `file` (single file), `phone_number_id`
- `GET /get`
  - Query: `media` (encrypted media token)
  - Optional header: `Range: bytes=...` for browser/video playback and seek support
  - Returns streamed media bytes (`Content-Type` from stored object metadata)
  - Supports `206 Partial Content` when a valid range is requested
- `POST /encrypt-link`
  - Body (JSON): Media link/URL to encrypt
- `DELETE /delete`
  - Query: `phone_number_id` and `id`
- `GET /list`
  - Optional query: filter/pagination params

## Authentication Notes

- Login endpoint (`POST /api/v1/auth/login`) returns an AES-encrypted JWT token in a cookie `access_token` (HttpOnly).
- For protected routes, clients can send the encrypted token in the `Authorization: Bearer <encrypted_token>` header.
- Token parsing middleware exists globally and expects an AES-encrypted token.
- The protected guard (most `/api/v1/*` routes) checks for `token_sub` context value.
- If `token_sub` is missing, protected routes return `401 Unauthorized`.
- Broadcast `/send` endpoint uses JWT-based authentication:
  - Extracts encrypted JWT from `Authorization` header
  - Decrypts and validates the token
  - Extracts `jwt_sub` from the token and sets it in context
  - Has a "pass-through" mode where it continues with `ctx.Next()` even if validation fails, but only triggers async send if `jwt_sub` is present

## Global Middleware

- Request logger
- Panic recovery
- CORS (config-driven)
- OPTIONS preflight short-circuit
- Request body/file-size guard at 16 MB (`413` when exceeded)
- Access token parsing and validation
- Protected route authorization

## gRPC Infrastructure

The service exposes a gRPC server on the same port as the HTTP (Fiber) server with the following services:

### gRPC Services

#### Message Service
- `SendMessage` - Send WhatsApp messages via gRPC
- Proto definition: `docs/proto/v1/message.proto`

#### Storage Media Service
- Media asset management via gRPC
- Proto definition: `docs/proto/v1/storage-media.proto`

### HMAC gRPC Interceptor

The gRPC interceptor implements custom HMAC-based authentication:
- Expects clients to include an `x-signature` metadata containing an HMAC signature of the request payload
- Body is JSON-marshaled with 1 space indentation
- Checks for a timestamp in `x-timestamp` metadata to prevent replay attacks ([replay prevention reference](https://webhooks.fyi/security/replay-prevention))
- Uses SHA256 as the hashing algorithm for good balance between security and performance
- Signature generated from URI + content ([reference](https://learn.microsoft.com/en-us/azure/communication-services/tutorials/hmac-header-tutorial?pivots=programming-language-csharp))
- Uses shared secret configured via `HMAC_SECRET`

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



## Configuration

The app loads `.env` via `godotenv` in development and then builds config from environment variables.

Required variables:

- `APP_NAME`
- `APP_VERSION`
- `APP_ENVIRONMENT` (`development|production|staging`)
- `APP_PORT`
- `APP_URL`
- `DATABASE_URL`
- `AES_ENCRYPTION_KEY` - Token encryption key
- `GCP_PROJECT_ID`
- `GCP_APP_BASE_URL`
- `GCP_TASK_BROADCAST_PARENT`
- `JOSE_RSA_PRIVATE_KEY` - Path to RSA private key file
- `JOSE_ACCESS_TOKEN_EXPIRY` - Go duration format (e.g., `24h`)
- `CORS_ENABLED`
- `HMAC_SECRET` - Shared secret for gRPC HMAC authentication

Optional/common variables:

- `APP_SECURE_COOKIE`
- `CORS_ALLOW_ORIGINS`
- `CORS_ALLOW_METHODS`
- `CORS_ALLOW_HEADERS`
- `CORS_EXPOSE_HEADERS`
- `CORS_ALLOW_CREDENTIALS`
- `META_APP_ID`
- `META_GRAPH_API_VERSION`
- `GOOGLE_APPLICATION_CREDENTIALS` - Recommended for local/dev if using Application Default Credentials
- `MEILISEARCH_HOST` - Meilisearch server URL
- `MEILISEARCH_API_KEY` - Meilisearch API key

## Proto Code Generation

To regenerate proto files, install the required tools:

```bash
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

This project uses `protoc` version 25.9.

To regenerate:

```bash
make proto
```

Proto files are located in `docs/proto/v1/` with generated code in the same directory.


## Runtime Behavior

At startup, the service initializes:

- Firebase app
- Firestore client
- Google Cloud Storage client
- Google Cloud Tasks client
- PostgreSQL database connection
- Meilisearch search index

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

## Utilities & Libraries

### `pkg/api_response/`
- Standardized JSON envelope (code, message, data, errors)
- HTTP status code mapping
- gRPC response builder for error conversion

### `pkg/errs/`
- Authentication-specific errors (invalid credentials, expired tokens)
- Database-specific errors
- Generic errors (forbidden, not found, already exists)
- Validation error handling
- Error classifier for auth detection

### `pkg/filter_request/`
- Operator-based query filtering (eq, ne, lt, gt, etc.)
- Type-specific filters: string, number, email, UUID
- Pagination and sorting helpers

### `pkg/meta/`
- WhatsApp Business API metadata structures
- Webhook payload parsing and marshaling
- Media metadata handling

### `pkg/server/`
- HTTP server setup (Fiber) with configurable body limits and timeouts
- gRPC server setup with HMAC authentication interceptor
- Graceful shutdown support
- Middleware initialization and composition

### `pkg/service-accounts/`
- Firebase service account key storage
- GCP credentials management

### `pkg/utils/`
- **Logger** - Dual output (stdout + `app.log`), environment-aware formatting (ANSI colors in development, plain text in production), structured logging with Zap
- **Validator** - Custom validators for file extensions and file counts, filter option validation, tag-based validation with custom error formatting
- **Transaction Manager** - Unified interface for mixed-database transactions: `Do()` (combined Firestore + GORM), `DoFirestore()`, `DoGorm()`
- **Web Utilities** - HTTP HEAD requests for header fetching, file download functionality, URL filename extraction with query parameter cleanup
- **JSON/Map Utilities** - Bidirectional struct ↔ map conversion, null value omission, JSON string serialization
- **Formatter** - Data formatting utilities



## Make Targets

- `make run` - Start development server
- `make build` - Build `wa-chat-service.exe`
- `make clean` - Remove built executable
- `make proto` - Generate gRPC code from `.proto` files
- `make meili-up` - Start Meilisearch container
- `make meili-down` - Stop Meilisearch container
- `make docker` - Build image and run docker compose

## Running with Docker

```bash
docker compose up --build
```

Or:

```bash
make docker
```

Default compose services:

- `wa-chat-service` on `8120`
- `Meilisearch` (optional search backend)

## Project Structure

```text
.
|-- cmd/
|   `-- app/          # Main HTTP/gRPC service
|-- config/           # Configuration loading
|-- docs/
|   `-- proto/v1/     # gRPC proto definitions and generated code
|-- internal/
|   |-- app/          # Application wiring and dependency injection
|   |-- dto/          # Data transfer objects with validation
|   |-- handler/
|   |   |-- grpc/     # gRPC service handlers
|   |   `-- http/     # HTTP handlers (Fiber)
|   |       |-- middleware/  # HTTP middleware (access_token, jwt, etc.)
|   |       `-- v1/   # API v1 handlers
|   |-- model/        # Domain models (Chat, Message, Broadcast, etc.)
|   |-- repository/   # Data access layer
|   |   |-- firestore/ # Firestore-specific implementation
|   |   |-- meili/     # Meilisearch implementation
|   |   `-- types.go   # Repository interfaces
|   |-- service/      # Business logic services
|   |   |-- access_token/  # Token management
|   |   |-- encrypt/       # Encryption/decryption
|   |   |-- google/        # Google Cloud services
|   |   |-- jose/          # JWT signing
|   |   `-- whatsapp/      # WhatsApp API integration
|   `-- usecase/      # Use case orchestration
|-- pkg/              # Shared utilities
|   |-- api_response/ # Standard response envelopes
|   |-- errs/         # Error handling and types
|   |-- filter_request/ # Query filtering and pagination
|   |-- meta/         # WhatsApp metadata structures
|   |-- server/       # HTTP/gRPC server setup
|   `-- utils/        # Helper functions (logger, validator, etc.)
|-- Dockerfile
|-- docker-compose.yml
|-- docker-compose.meili.yml
`-- Makefile
```

## Important Notes

### Template
- When creating templates, if there's no parameter in the components, the `parameter_format` can be filled or left null, both are valid.
- If there's parameter(s) in the components, the `parameter_format` must be filled with either "NAMED" or "POSITIONAL".

### Broadcast
- Currently does not support Button for multi-product message, one-time password, voice call, or single product message.

### Media
- Resumable API is for creating template and profile picture.
- `GET /api/v1/storage-media/get` streams responses in chunks. For non-range requests, the response is streamed without a fixed `Content-Length`; for valid ranges, the server returns `Content-Length` and `Content-Range`.
