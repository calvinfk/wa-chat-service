package service

import (
	"context"
	"time"
	"wa_chat_service/pkg/meta/whatsapp_business"
	whatsapp_business_component "wa_chat_service/pkg/meta/whatsapp_business/component"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type (
	// AccessToken is an interface that defines methods for generating and parsing access tokens.
	AccessToken interface {
		// ParseAccessToken parses a AccessToken access token string and extracts the user ID (subject) from the token claims.
		// It validates the token using the configured JWK and returns the user ID as a UUID if the token is valid.
		// If the token is expired, it returns an error indicating that the token has expired and the sub.
		// If the token is invalid for any other reason, it returns only the error indicating that the token is invalid.
		ParseAccessTokenSub(tokenStr string) (uuid.UUID, error)
	}
	Encrypt interface {
		// Encrypt encrypts a byte slice and returns the encrypted string. If there is an error during encryption, it returns the error.
		Encrypt(plainText string) (string, error)
		// Decrypt decrypts a byte slice containing the encrypted data and returns the original byte slice. If there is an error during decryption, it returns the error.
		Decrypt(ciphertext string) (string, error)
	}

	GoogleStorage interface {
		// UploadFile uploads a file to Google Cloud Storage. It takes the file data as a byte slice, the destination path in the storage bucket, and the content type of the file. It returns the public URL of the uploaded file if the upload is successful, or an error if there is an issue during the upload process.
		UploadFile(ctx context.Context, fileData []byte, bucket string, destinationPath string, contentType string) (string, error)
		// GetFile retrieves a file from Google Cloud Storage. It takes the file URL as a parameter and returns a storage.Reader for reading the file content, the file's attributes, and an error if there is an issue during the retrieval process.
		// The storage.Reader allows you to read the file content directly from Google Cloud Storage. The caller is responsible for closing the reader after use to free up resources.
		GetFile(ctx context.Context, fileURL string) (*storage.Reader, *storage.ObjectAttrs, error)
		// GenerateV4GetObjectSignedURL generates a signed URL for accessing an object in Google Cloud Storage. It takes the bucket name and object name as parameters and returns the signed URL as a string if the generation is successful, or an error if there is an issue during the generation process.
		GenerateV4GetObjectSignedURL(bucketName, objectName string, expiration time.Duration) (string, error)
	}

	GoogleFirebase interface {
		// SendNotification sends a notification to the specified device tokens using Firebase Cloud Messaging. It takes the notification title, body, and a slice of device tokens as parameters. It returns an error if there is an issue during the sending process.
		SendNotification(ctx context.Context, title string, body string, tokens []string) error
		// SubscribeToTopic subscribes the specified device tokens to topics in Firebase Cloud Messaging. It takes the topic names and a slice of device tokens as parameters. It returns an error if there is an issue during the subscription process.
		SubscribeToTopic(ctx context.Context, topics []string, tokens []string) error
		// UnsubscribeFromTopic unsubscribes the specified device tokens from topics in Firebase Cloud Messaging. It takes the topic names and a slice of device tokens as parameters. It returns an error if there is an issue during the unsubscription process.
		UnsubscribeFromTopic(ctx context.Context, topics []string, tokens []string) error
		// SendNotificationToTopic sends a notification to the specified topic in Firebase Cloud Messaging. It takes the notification title, body, and the topic name as parameters. It returns an error if there is an issue during the sending process.
		SendNotificationToTopic(ctx context.Context, title string, body string, topic string) error
		// UploadFile uploads a file to Firebase Storage. It takes the file data as a byte slice, the destination path in the storage bucket, and the content type of the file. It returns the public URL of the uploaded file if the upload is successful, or an error if there is an issue during the upload process.
		UploadFile(ctx context.Context, filePath string, file []byte) (*storage.ObjectAttrs, error)
	}

	WhatsappService interface {
		SendMessage(ctx context.Context, client *whatsapp_business.Client, to string, payload whatsapp_business_component.MessageComponent) (whatsapp_business.MessageResponse, int, error)
		GetTemplateList(ctx context.Context, client *whatsapp_business.Client) ([]any, int, error)
	}
)
