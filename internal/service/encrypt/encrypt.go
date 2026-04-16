package encrypt_service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"wa_chat_service/config"

	"go.uber.org/zap"
)

type EncryptService struct {
	config *config.Encrypt
	zslog  *zap.SugaredLogger
}

func NewEncryptService(config *config.Encrypt, zslog *zap.SugaredLogger) *EncryptService {
	return &EncryptService{
		config: config,
		zslog:  zslog,
	}
}
func (s *EncryptService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.config.Key)
	if err != nil {
		s.zslog.Errorf("[Encrypt] error creating AES cipher: %v", err)
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.zslog.Errorf("[Encrypt] error creating GCM: %v", err)
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		s.zslog.Errorf("[Encrypt] error generating nonce: %v", err)
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, ciphertext...)

	return base64.RawURLEncoding.EncodeToString(payload), nil
}
func (s *EncryptService) Decrypt(cipherText string) (string, error) {
	cipherTextDecoded, err := base64.RawURLEncoding.DecodeString(cipherText)
	if err != nil {
		s.zslog.Errorf("[Decrypt] error decoding cipher text: %v", err)
		return "", err
	}

	block, err := aes.NewCipher(s.config.Key)
	if err != nil {
		s.zslog.Errorf("[Decrypt] error creating AES cipher: %v", err)
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.zslog.Errorf("[Decrypt] error creating GCM: %v", err)
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherTextDecoded) < nonceSize {
		s.zslog.Errorf("[Decrypt] invalid encrypted payload: too short")
		return "", fmt.Errorf("invalid encrypted payload")
	}

	nonce, encryptedData := cipherTextDecoded[:nonceSize], cipherTextDecoded[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		s.zslog.Errorf("[Decrypt] error decrypting data: %v", err)
		return "", err
	}

	return string(plaintext), nil
}
