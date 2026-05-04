package middleware

import (
	"net/http"
	"wa_chat_service/internal/service"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Checks for the presence of a JWT token in the Authorization header.
// It validates the token using the provided JWT service and, if valid, extracts the user ID (sub) and sets it in the Gin context for use in subsequent handlers.
// If the token is missing or invalid, it responds with failCode for the http response code and an appropriate error message.
// The 'pass' parameter allows the middleware to continue to the next handler even if the token is missing or invalid, which can be useful for routes that allow both authenticated and unauthenticated access.
func Jwt(encryptService service.Encrypt, jwtService service.JWT, failCode int, pass bool, zsLog *zap.SugaredLogger) fiber.Handler {
	if failCode == 0 {
		failCode = http.StatusUnauthorized
	}
	return func(ctx fiber.Ctx) error {
		var err error
		encryptedToken := ctx.Get("Authorization")
		if encryptedToken == "" {
			zsLog.Errorf("[Jwt] Authorization header or cookie is required: %v", err)
			if pass {
				return ctx.Next()
			}
			return ctx.Status(failCode).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"data":    nil,
				"message": "Authorization header is required",
				"errors":  nil,
			})
		}
		if len(encryptedToken) > 7 && encryptedToken[:7] == "Bearer " {
			encryptedToken = encryptedToken[7:]
		}
		if encryptedToken == "" {
			if pass {
				return ctx.Next()
			}
			return ctx.Status(failCode).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"data":    nil,
				"message": "Token is required",
				"errors":  nil,
			})
		}
		tokenString, err := encryptService.Decrypt(encryptedToken)
		if err != nil {
			zsLog.Errorf("[Jwt] Failed to decrypt token: %v", err)
			if pass {
				return ctx.Next()
			}
			return ctx.Status(failCode).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"data":    nil,
				"message": "Invalid token",
				"errors":  nil,
			})
		}
		sub, err := jwtService.ParseJWT(tokenString)
		if err != nil {
			zsLog.Errorf("[Jwt] Failed to parse JWT: %v", err)
			if pass {
				return ctx.Next()
			}
			return ctx.Status(failCode).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"data":    nil,
				"message": err.Error(),
				"errors":  nil,
			})
		}
		ctx.Locals("jwt_sub", sub)
		return ctx.Next()
	}
}
