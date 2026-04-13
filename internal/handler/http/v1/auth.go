package http_v1

import (
	"log"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authUsecase usecase.Auth
	cfg         *config.Config
}

func NewAuthHandler(authUsecase usecase.Auth, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		cfg:         cfg,
	}
}

func (h *AuthHandler) RegisterRoutes(api fiber.Router) {
	authGroup := api.Group("/auth")
	{
		authGroup.Post("/login", h.login)
	}
}

func (h *AuthHandler) login(ctx fiber.Ctx) error {
	var requestData dto.AuthLoginRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		log.Println("[ERROR][internal/handler/http/v1/auth.go][login] ctx.Bind().Body error:", err)
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	encrypedToken, serverError, err := h.authUsecase.Login(ctx.Context(), requestData)
	if err != nil {
		code, response := api_response.NewApiResponse(serverError, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	log.Println("[INFO][internal/handler/http/v1/auth.go][login] Login successful for tenantID:", requestData.TenantID)
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    encrypedToken,
		Expires:  time.Now().Add(h.cfg.JOSE.AccessTokenExpiry),
		HTTPOnly: true,
	})
	code, response := api_response.NewApiResponse(false, nil, "Login successful", nil)
	return ctx.Status(code).JSON(response)
}
