package middleware

import (
	"log"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"

	"github.com/gofiber/fiber/v3"
)

// Checks for the presence of a JWT token in the Authorization header.
// It validates the token using the provided JWT service and, if valid, extracts the user ID (sub) and sets it in the Gin context for use in subsequent handlers.
// If the token is missing or invalid, it responds with a 401 Unauthorized code and an appropriate error message.
func AccessToken(accessTokenService service.AccessToken, encryptService service.Encrypt) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		var err error
		tokenString := ctx.Cookies("access_token", "")
		if tokenString == "" {
			ctx.Set("jwt_error_message", "Access token is missing")
			ctx.Next()
			return nil
		}
		decryptedToken, err := encryptService.Decrypt(tokenString)
		if err != nil {
			log.Printf("[ERROR][internal/handler/http/middleware/jwt.go][AccessToken] error decrypting token: %v", err)
			ctx.Set("jwt_error_message", "Invalid token")
			ctx.Next()
			return nil
		}
		sub, err := accessTokenService.ParseAccessTokenSub(string(decryptedToken))
		if err != nil {
			if err == errs.ErrAuthExpiredAccessToken {
				ctx.Set("userID", sub.String())
				ctx.Set("jwt_error_message", "Token expired")
			} else {
				log.Printf("[ERROR][internal/handler/http/middleware/jwt.go][AccessToken] error parsing token: %v", err)
				ctx.Set("jwt_error_message", "Invalid token")
			}
			ctx.Next()
			return nil
		}
		ctx.Set("userID", sub.String())
		ctx.Next()
		return nil
	}
}
