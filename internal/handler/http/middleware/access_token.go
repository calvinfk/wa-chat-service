package middleware

import (
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Checks for the presence of a JWT token in the Authorization header.
// It validates the token using the provided JWT service and, if valid, extracts the user ID (sub) and sets it in the Gin context for use in subsequent handlers.
// If the token is missing or invalid, it responds with a 401 Unauthorized code and an appropriate error message.
func AccessToken(accessTokenService service.AccessToken, encryptService service.Encrypt, zsLog *zap.SugaredLogger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		var err error
		tokenString := ctx.Cookies("access_token", "")
		if tokenString == "" {
			ctx.Locals("token_error_message", "Access token is missing")
			return ctx.Next()
		}
		decryptedToken, err := encryptService.Decrypt(tokenString)
		if err != nil {
			zsLog.Errorf("[ERROR][internal/handler/http/middleware/jwt.go][AccessToken] error decrypting token: %v", err)
			ctx.Locals("token_error_message", "Invalid token")
			return ctx.Next()
		}
		sub, err := accessTokenService.ParseAccessTokenSub(string(decryptedToken))
		if err != nil {
			if err == errs.ErrAuthExpiredAccessToken {
				ctx.Locals("token_sub", sub)
				ctx.Locals("token_error_message", "Token expired")
			} else {
				zsLog.Errorf("[ERROR][internal/handler/http/middleware/jwt.go][AccessToken] error parsing token: %v", err)
				ctx.Locals("token_error_message", "Invalid token")
			}
			return ctx.Next()
		}
		ctx.Locals("token_sub", sub)
		return ctx.Next()
	}
}
