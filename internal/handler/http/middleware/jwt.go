package middleware

import (
	"log"
	"net/http"
	"wa_chat_service/internal/service"

	"github.com/gofiber/fiber/v3"
)

func Jwt(encryptService service.Encrypt, jwtService service.JWT) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		var err error
		encryptedToken := ctx.Get("Authorization")
		if encryptedToken == "" {
			log.Println("[ERROR][internal/handler/http/middleware/Jwt] Authorization header or cookie is required: ", err)
			return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
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
			return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"data":    nil,
				"message": "Token is required",
				"errors":  nil,
			})
		}
		tokenString, err := encryptService.Decrypt(encryptedToken)
		if err != nil {
			log.Println("[ERROR][internal/handler/http/middleware/Jwt] Failed to decrypt token: ", err)
			return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"data":    nil,
				"message": "Invalid token",
				"errors":  nil,
			})
		}
		sub, err := jwtService.ParseJWT(tokenString)
		if err != nil {
			return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
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
