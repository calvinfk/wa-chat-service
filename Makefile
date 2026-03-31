# Variables
APP_NAME=wa-chat-service
MAIN_PATH=./cmd/app/main.go
PORT ?= 8121
VER ?=
ARGS ?=

.PHONY: run build clean test-integration save

# 1. Build and Run the App
run:
	@cls || clear
	go run $(MAIN_PATH)

migrate:
	go run ./cmd/migrate/migrate.go $(ARGS)

# 4. Build the final Windows Binary
build:
	go build -o $(APP_NAME).exe $(MAIN_PATH)

clean:
	@if exist $(APP_NAME).exe del $(APP_NAME).exe

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
