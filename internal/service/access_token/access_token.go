package access_token_service

import (
	"strconv"
	"strings"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/pkg/errs"

	"go.uber.org/zap"
)

// accessTokenService is a struct that implements the JWT interface defined in internal/service/types.go. It provides methods for generating and parsing JWT access tokens using the RS256 signing algorithm and the configured JWK. The service uses the configuration provided in the config.Config struct to determine the token expiry time and the JWK for signing and validating tokens.
type accessTokenService struct {
	config *config.Config
	zslog  *zap.SugaredLogger
}

// NewAccessTokenService creates a new instance of accessTokenService with the provided configuration. This service is responsible for generating and parsing JWT access tokens using the RS256 signing algorithm and the configured JWK. It implements the JWT interface defined in internal/service/types.go, allowing it to be used as a dependency in other parts of the application that require JWT functionality.
func NewAccessTokenService(config *config.Config, zslog *zap.SugaredLogger) *accessTokenService {
	return &accessTokenService{
		config: config,
		zslog:  zslog,
	}
}

func (s *accessTokenService) GenerateAccessToken(sub string) (string, error) {
	exp := time.Now().Add(s.config.JOSE.AccessTokenExpiry).Unix()
	token := sub + "." + strconv.FormatInt(exp, 10)
	return token, nil
}

func (s *accessTokenService) ParseAccessTokenSub(tokenStr string) (string, error) {
	subStr, expStr, ok := strings.Cut(tokenStr, ".")
	if !ok {
		s.zslog.Errorf("[ParseAccessTokenSub] token format invalid: %s", tokenStr)
		return "", errs.ErrAuthInvalidAccessToken
	}
	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		s.zslog.Errorf("[ParseAccessTokenSub] error parsing token expiration: %v", err)
		return "", errs.ErrAuthInvalidAccessToken
	}
	if time.Now().Unix() > expUnix {
		s.zslog.Errorf("[ParseAccessTokenSub] access token expired")
		return subStr, errs.ErrAuthExpiredAccessToken
	}
	return subStr, nil
}
