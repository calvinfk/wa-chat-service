package middleware

import (
	"slices"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/errs"

	"github.com/gofiber/fiber/v3"
)

// Role is a middleware that checks if the user's role, which is extracted from the access token and stored in the context by the AccessToken middleware, is included in the list of allowed roles for the route.
// If the user's role is not in the allowed roles, it responds with a 403 Forbidden code and an appropriate error message.
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
