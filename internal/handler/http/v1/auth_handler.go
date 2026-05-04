package http_v1

import (
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
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	encrypedToken, serverError, err := h.authUsecase.Login(ctx.Context(), requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	// Set the encrypted token in an HTTP-only cookie with the appropriate expiration time
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    encrypedToken,
		Expires:  time.Now().Add(h.cfg.JOSE.AccessTokenExpiry),
		Secure:   h.cfg.App.SecureCookie,
		HTTPOnly: true,
	})
	code, response := api_response.NewApiResponse("Login successful", nil)
	return ctx.Status(code).JSON(response)
}
