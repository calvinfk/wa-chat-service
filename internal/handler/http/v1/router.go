package http_v1

import (
	"wa_chat_service/config"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type RouterHandlerV1 struct {
	ActivityLogUsecase   usecase.ActivityLog
	MessageUsecase       usecase.Message
	StorageMediaUsecase  usecase.StorageMedia
	ChatUsecase          usecase.Chat
	TemplateUsecase      usecase.Template
	AccessTokenService   service.AccessToken
	EncryptService       service.Encrypt
	GoogleStorageService service.GoogleStorage
}

type HandlerV1 interface {
	RegisterRoute(api fiber.Router)
}

func NewApiV1Routes(api fiber.Router, routerHandler RouterHandlerV1, config *config.Config) {
	chatHandler := NewChatHandler(routerHandler.MessageUsecase, routerHandler.ChatUsecase)
	chatHandler.RegisterRoute(api)
	storageMediaHandler := NewStorageMediaHandler(routerHandler.StorageMediaUsecase)
	storageMediaHandler.RegisterRoutes(api)
	templateHandler := NewTemplateHandler(routerHandler.TemplateUsecase)
	templateHandler.RegisterRoute(api)
}
