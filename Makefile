# Variables
APP_NAME=wa-chat-service
MAIN_PATH=./cmd/app/main.go
PORT ?= 8121
VER ?=
ARGS ?=

.PHONY: run build clean proto meili-up meili-down docker

run:
	@cls || clear
	go run $(MAIN_PATH)

build:
	go build -o $(APP_NAME).exe $(MAIN_PATH)

clean:
	@if exist $(APP_NAME).exe del $(APP_NAME).exe

# --proto_path=./docs/proto/ for base path,  ./docs/proto/**/*.proto for all proto files in subdirectories
proto:
	protoc --go_out=. --go-grpc_out=. --proto_path=./docs/proto/ ./docs/proto/**/*.proto

meili-up:
	docker compose -f docker-compose.meili.yml up -d
meili-down:
	docker compose -f docker-compose.meili.yml down

docker:
	docker build -t $(APP_NAME):latest .
	docker compose up
