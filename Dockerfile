# Stage 1: Build
FROM golang:1.26.0 AS builder

WORKDIR /app
RUN chown 65532:65532 .
# the go mod download layer is cached if the go.mod and go.sum files are not changed
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o wa-chat-service ./cmd/app

# Stage 2: Run (minimal image)
FROM gcr.io/distroless/static-debian12 AS runner
COPY --from=builder --chown=nonroot:nonroot /app/wa-chat-service /app/wa-chat-service
WORKDIR /app
COPY --chown=nonroot:nonroot private.pem private.pem
ENV HTTP_PORT=8121
EXPOSE 8121
USER nonroot
ENTRYPOINT ["./wa-chat-service"]
