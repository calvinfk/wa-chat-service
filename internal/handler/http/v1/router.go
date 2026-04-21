package http_v1

import (
	"wa_chat_service/config"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type HandlerHTTPV1 struct {
	MessageUsecase      usecase.Message
	StorageMediaUsecase usecase.StorageMedia
	ChatUsecase         usecase.Chat
	TemplateUsecase     usecase.Template
	BroadcastUsecase    usecase.Broadcast
	TenantUsecase       usecase.Tenant
	AuthUsecase         usecase.Auth
	EncryptService      service.Encrypt
	JWTService          service.JWT
	AccessTokenService  service.AccessToken
	ZSLog               *zap.SugaredLogger
}

type HandlerV1 interface {
	RegisterRoute(api fiber.Router)
}

func New(api fiber.Router, routerHandler HandlerHTTPV1, config *config.Config) {
	chatHandler := NewChatHandler(routerHandler.MessageUsecase, routerHandler.ChatUsecase)
	chatHandler.RegisterRoute(api)
	storageMediaHandler := NewStorageMediaHandler(routerHandler.StorageMediaUsecase, routerHandler.ZSLog)
	storageMediaHandler.RegisterRoutes(api)
	templateHandler := NewTemplateHandler(routerHandler.TemplateUsecase)
	templateHandler.RegisterRoute(api)
	broadcastHandler := NewBroadcastHandler(routerHandler.BroadcastUsecase, routerHandler.EncryptService, routerHandler.JWTService, routerHandler.ZSLog)
	broadcastHandler.RegisterRoute(api)
	tenantHandler := NewTenantHandler(routerHandler.TenantUsecase)
	tenantHandler.RegisterRoute(api)
	authHandler := NewAuthHandler(routerHandler.AuthUsecase, config)
	authHandler.RegisterRoutes(api)
}
