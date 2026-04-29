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
	zsLog  *zap.SugaredLogger
}

func NewEncryptService(config *config.Encrypt, zsLog *zap.SugaredLogger) *EncryptService {
	return &EncryptService{
		config: config,
		zsLog:  zsLog,
	}
}
func (s *EncryptService) Encrypt(plaintext string) (string, error) {
	// Initialize the AES block cipher using your secret key.
	// The key length must be 16, 24, or 32 bytes for AES-128, 192, or 256.
	block, err := aes.NewCipher(s.config.Key)
	if err != nil {
		s.zsLog.Errorf("[Encrypt] error creating AES cipher: %v", err)
		return "", err
	}

	// Wrap the block cipher in GCM mode.
	// This adds the "Authenticated" part of AEAD, generating an auth tag.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.zsLog.Errorf("[Encrypt] error creating GCM: %v", err)
		return "", err
	}

	// Generate a cryptographically secure random nonce (number used once).
	// Standard GCM nonce size is 12 bytes (96 bits).
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		s.zsLog.Errorf("[Encrypt] error generating nonce: %v", err)
		return "", err
	}

	// Encrypt the data and append the authentication tag.
	// Seal takes: (dst, nonce, plaintext, additionalData).
	// It returns the ciphertext + the authentication tag (MAC).
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Concatenate the nonce and the encrypted data.
	// The nonce is required for decryption, so we ship it with the ciphertext.
	payload := append(nonce, ciphertext...)

	// Encode the combined nonce and ciphertext in base64 for safe transport.
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func (s *EncryptService) Decrypt(cipherText string) (string, error) {
	// Decode the Base64 string back into raw bytes.
	cipherTextDecoded, err := base64.RawURLEncoding.DecodeString(cipherText)
	if err != nil {
		s.zsLog.Errorf("[Decrypt] error decoding cipher text: %v", err)
		return "", err
	}

	// Re-initialize the AES and GCM cipher.
	// The key must be identical to the one used for encryption.
	block, err := aes.NewCipher(s.config.Key)
	if err != nil {
		s.zsLog.Errorf("[Decrypt] error creating AES cipher: %v", err)
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.zsLog.Errorf("[Decrypt] error creating GCM: %v", err)
		return "", err
	}

	//Validate that the payload is large enough to contain at least the nonce.
	// If it's smaller, it's malformed data.
	nonceSize := gcm.NonceSize()
	if len(cipherTextDecoded) < nonceSize {
		s.zsLog.Errorf("[Decrypt] invalid encrypted payload: too short")
		return "", fmt.Errorf("invalid encrypted payload")
	}

	// Split the data back into its constituent parts: Nonce and Ciphertext.
	nonce, encryptedData := cipherTextDecoded[:nonceSize], cipherTextDecoded[nonceSize:]

	// Decrypt and Authenticate.
	// 'Open' does two things:
	// a) Checks the authentication tag (integrity). If tampered with, it returns an error.
	// b) Decrypts the ciphertext (confidentiality).
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		s.zsLog.Errorf("[Decrypt] error decrypting data: %v", err)
		return "", err
	}

	return string(plaintext), nil
}
