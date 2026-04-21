# Variables
APP_NAME=wa-chat-service
MAIN_PATH=./cmd/app/main.go
PORT ?= 8121
VER ?=
ARGS ?=

.PHONY: run build clean test-integration save meili-up

# 1. Build and Run the App
run:
	@cls || clear
	go run $(MAIN_PATH)

encrypt:
	go run ./cmd/encrypt/encrypt.go encrypt=$(ARGS)

decrypt:
	go run ./cmd/encrypt/encrypt.go decrypt=$(ARGS)

# 4. Build the final Windows Binary
build:
	go build -o $(APP_NAME).exe $(MAIN_PATH)

clean:
	@if exist $(APP_NAME).exe del $(APP_NAME).exe

proto:
	protoc --go_out=. --go-grpc_out=. --proto_path=./docs/proto/ ./docs/proto/**/*.proto

meili-up:
	docker compose -f docker-compose.meili.yml up -d
meili-down:
	docker compose -f docker-compose.meili.yml down

sync:
	go run ./cmd/sync/main.go $(ARGS)

save:
ifeq ($(VER),)
	docker build -t $(APP_NAME):latest .
	docker image save $(APP_NAME) -o $(APP_NAME).tar
else
	docker build -t $(APP_NAME):$(VER) -t $(APP_NAME):latest .
	docker image save $(APP_NAME):$(VER) -o $(APP_NAME).tar
endif

docker:
	docker build -t $(APP_NAME):latest .
	docker compose up
