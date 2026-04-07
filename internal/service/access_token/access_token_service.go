package access_token_service

import (
	"log"
	"strconv"
	"strings"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/pkg/errs"

	"github.com/google/uuid"
)

// accessTokenService is a struct that implements the JWT interface defined in internal/service/types.go. It provides methods for generating and parsing JWT access tokens using the RS256 signing algorithm and the configured JWK. The service uses the configuration provided in the config.Config struct to determine the token expiry time and the JWK for signing and validating tokens.
type accessTokenService struct {
	config *config.Config
}

// NewAccessTokenService creates a new instance of accessTokenService with the provided configuration. This service is responsible for generating and parsing JWT access tokens using the RS256 signing algorithm and the configured JWK. It implements the JWT interface defined in internal/service/types.go, allowing it to be used as a dependency in other parts of the application that require JWT functionality.
func NewAccessTokenService(config *config.Config) *accessTokenService {
	return &accessTokenService{config: config}
}

func (s *accessTokenService) ParseAccessTokenSub(tokenStr string) (uuid.UUID, error) {
	subStr, expStr, ok := strings.Cut(tokenStr, ".")
	if !ok {
		log.Printf("[ERROR][internal/service/access_token/access_token_service.go][ParseAccessTokenSub] token format invalid: %s", tokenStr)
		return uuid.Nil, errs.ErrAuthInvalidAccessToken
	}
	sub, err := uuid.Parse(subStr)
	if err != nil {
		log.Printf("[ERROR][internal/service/access_token/access_token_service.go][ParseAccessTokenSub] error parsing token subject: %v", err)
		return uuid.Nil, errs.ErrAuthInvalidAccessToken
	}
	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		log.Printf("[ERROR][internal/service/access_token/access_token_service.go][ParseAccessTokenSub] error parsing token expiration: %v", err)
		return uuid.Nil, errs.ErrAuthInvalidAccessToken
	}
	if time.Now().Unix() > expUnix {
		return sub, errs.ErrAuthExpiredAccessToken
	}
	return sub, nil
}
