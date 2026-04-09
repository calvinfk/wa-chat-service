package jose_service

import (
	"crypto/rsa"
	"fmt"
	"wa_chat_service/config"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

type jwtService struct {
	cfg *config.JOSE
}

func NewJWTService(cfg *config.JOSE) *jwtService {
	return &jwtService{cfg: cfg}
}

func (s *jwtService) GenerateJWT(sub any, expiredAt int64) (string, error) {
	token := jwt.New()
	token.Set(jwt.SubjectKey, sub)
	token.Set(jwt.ExpirationKey, expiredAt)
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256(), s.cfg.RSAPrivateKey))
	if err != nil {
		return "", err
	}
	return string(signed), err
}

func (s *jwtService) ParseJWT(tokenStr string) (any, error) {
	claims, err := jwt.ParseString(tokenStr, jwt.WithKey(jwa.RS256(), s.cfg.RSAPrivateKey.Public().(*rsa.PublicKey)))
	if err != nil {
		return nil, err
	}
	sub, ok := claims.Subject()
	if !ok {
		return nil, fmt.Errorf("subject claim not found in token")
	}
	return sub, nil
}
