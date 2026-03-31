package http_v1

import (
	"wa_chat_service/config"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type RouterHandlerV1 struct {
	ActivityLogUsecase    usecase.ActivityLog
	AccessTokenService    service.AccessToken
	MessageUsecase        usecase.Message
	EncryptService        service.Encrypt
	GoogleStorageService  service.GoogleStorage
	GoogleFirebaseService service.GoogleFirebase
}

type HandlerV1 interface {
	RegisterRoute(api fiber.Router)
}

func NewApiV1Routes(api fiber.Router, routerHandler RouterHandlerV1, cfg *config.Config) {
	chatHandler := NewChatHandler(routerHandler.MessageUsecase)
	chatHandler.RegisterRoute(api)
}
