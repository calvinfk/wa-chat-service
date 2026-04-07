package encrypt_service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"wa_chat_service/config"
)

type EncryptService struct {
	config *config.Encrypt
}

func NewEncryptService(config *config.Encrypt) *EncryptService {
	return &EncryptService{config: config}
}
func (s *EncryptService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.config.Key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, ciphertext...)

	return base64.RawURLEncoding.EncodeToString(payload), nil
}
func (s *EncryptService) Decrypt(cipherText string) (string, error) {
	cipherTextDecoded, err := base64.RawURLEncoding.DecodeString(cipherText)
	if err != nil {
		// Backward compatibility for tokens previously encoded in hex.
		cipherTextDecoded, err = hex.DecodeString(cipherText)
		if err != nil {
			return "", err
		}
	}

	block, err := aes.NewCipher(s.config.Key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherTextDecoded) < nonceSize {
		return "", fmt.Errorf("invalid encrypted payload")
	}

	nonce, encryptedData := cipherTextDecoded[:nonceSize], cipherTextDecoded[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
