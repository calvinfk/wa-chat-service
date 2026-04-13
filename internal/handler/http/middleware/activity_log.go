package middleware

import (
	"net/http"
	"strings"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Logs API requests as activity logs.
// It extracts the user ID from the context, constructs a description of the API request,
// and then calls the activity log use case to insert a new activity log entry into the database.
// If there is an error during this process, it returns an appropriate HTTP response.
func ActivityLog(activityLogUsecase usecase.ActivityLog) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		inputData := dto.ActivityLogCreateRequest{
			Type: "API_REQUEST",
		}
		var descriptionBuilder strings.Builder
		descriptionBuilder.WriteString("API request made to")
		userID := ctx.Get("userID", "")
		if userID != "" {
			userIDParsed := uuid.MustParse(userID)
			inputData.UserID = &userIDParsed
		}
		if ctx.Request() != nil {
			descriptionBuilder.WriteString(" " + string(ctx.Request().Header.Method()))
			descriptionBuilder.WriteString(" " + ctx.Request().URI().String())
		}
		jwtError := ctx.Get("token_error_message", "")
		if jwtError == "Token expired" {
			descriptionBuilder.WriteString(" with expired token")
		}
		inputData.Description = descriptionBuilder.String()
		_, serverError, err := activityLogUsecase.Insert(ctx.Context(), inputData)
		if serverError || err != nil {
			ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"code":    http.StatusInternalServerError,
				"message": "Failed to log activity",
				"data":    nil,
				"errors":  nil,
			})
			return nil
		}
		return ctx.Next()
	}
}
