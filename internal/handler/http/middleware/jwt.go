package middleware

import (
	"net/http"
	"wa_chat_service/internal/service"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func Jwt(encryptService service.Encrypt, jwtService service.JWT, failCode int, pass bool, zsLog *zap.SugaredLogger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		var err error
		encryptedToken := ctx.Get("Authorization")
		if encryptedToken == "" {
			zsLog.Errorf("[ERROR][internal/handler/http/middleware/Jwt] Authorization header or cookie is required: %v", err)
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
			zsLog.Errorf("[ERROR][internal/handler/http/middleware/Jwt] Failed to decrypt token: %v", err)
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
