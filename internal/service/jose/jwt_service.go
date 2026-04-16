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
	zslog *zap.SugaredLogger
}

func NewJWTService(cfg *config.JOSE, zslog *zap.SugaredLogger) *jwtService {
	return &jwtService{
		cfg:   cfg,
		zslog: zslog,
	}
}

func (s *jwtService) GenerateJWT(sub any, expiredAt int64) (string, error) {
	token := jwt.New()
	token.Set(jwt.SubjectKey, sub)
	token.Set(jwt.ExpirationKey, expiredAt)
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256(), s.cfg.RSAPrivateKey))
	if err != nil {
		s.zslog.Errorf("[GenerateJWT] error signing JWT: %v", err)
		return "", err
	}
	return string(signed), err
}

func (s *jwtService) ParseJWT(tokenStr string) (any, error) {
	claims, err := jwt.ParseString(tokenStr, jwt.WithKey(jwa.RS256(), s.cfg.RSAPrivateKey.Public().(*rsa.PublicKey)))
	if err != nil {
		s.zslog.Errorf("[ParseJWT] error parsing JWT: %v", err)
		return nil, err
	}
	sub, ok := claims.Subject()
	if !ok {
		s.zslog.Errorf("[ParseJWT] subject claim not found in token")
		return nil, fmt.Errorf("subject claim not found in token")
	}
	return sub, nil
}
