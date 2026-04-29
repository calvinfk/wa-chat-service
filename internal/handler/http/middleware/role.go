package middleware

import (
	"slices"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/errs"

	"github.com/gofiber/fiber/v3"
)

func Role(allowedRoles ...model.UserRole) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userRole := ctx.Locals("token_sub").(dto.AuthData).Role
		if slices.Contains(allowedRoles, userRole) {
			return ctx.Next()
		}
		httpCode, response := api_response.NewErrorApiResponse(false, errs.ErrGenericForbidden)
		return ctx.Status(httpCode).JSON(response)
	}
}
