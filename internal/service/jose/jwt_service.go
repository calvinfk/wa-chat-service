package jose_service

import (
	"crypto/rsa"
	"fmt"
	"wa_chat_service/config"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"go.uber.org/zap"
)

type jwtService struct {
	cfg   *config.JOSE
	zsLog *zap.SugaredLogger
}

// NewJWTService creates a new instance of jwtService with the provided configuration and logger. This service is responsible for generating and parsing JWT tokens using RSA encryption based on the provided configuration.
func NewJWTService(cfg *config.JOSE, zsLog *zap.SugaredLogger) *jwtService {
	return &jwtService{
		cfg:   cfg,
		zsLog: zsLog,
	}
}

func (s *jwtService) GenerateJWT(sub any, expiredAt int64) (string, error) {
	token := jwt.New()
	token.Set(jwt.SubjectKey, sub)
	token.Set(jwt.ExpirationKey, expiredAt)
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256(), s.cfg.RSAPrivateKey))
	if err != nil {
		s.zsLog.Errorf("[GenerateJWT] error signing JWT: %v", err)
		return "", err
	}
	return string(signed), err
}

func (s *jwtService) ParseJWT(tokenStr string) (any, error) {
	claims, err := jwt.ParseString(tokenStr, jwt.WithKey(jwa.RS256(), s.cfg.RSAPrivateKey.Public().(*rsa.PublicKey)))
	if err != nil {
		s.zsLog.Errorf("[ParseJWT] error parsing JWT: %v", err)
		return nil, err
	}
	sub, ok := claims.Subject()
	if !ok {
		s.zsLog.Errorf("[ParseJWT] subject claim not found in token")
		return nil, fmt.Errorf("subject claim not found in token")
	}
	return sub, nil
}
