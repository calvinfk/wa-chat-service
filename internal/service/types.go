package service

import (
	"context"
	"net/http"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/pkg/meta/whatsapp_business"

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
		// Uploads a file to Google Cloud Storage. It takes the file data as a byte slice, the destination path in the storage bucket, and the content type of the file.
		// It returns the URL of the uploaded file if the upload is successful, or an error if there is an issue during the upload process.
		UploadFile(ctx context.Context, fileData []byte, fileURL string) (*storage.ObjectAttrs, error)
		// GetFile retrieves a file from Google Cloud Storage. It takes the file URL as a parameter
		// It returns a reader for the file, the file's attributes, and an error if there is an issue during the retrieval process.
		GetFile(ctx context.Context, fileURL string) (*storage.Reader, *storage.ObjectAttrs, error)
		// DeleteFile deletes a file from Google Cloud Storage. It takes the file URL as a parameter
		// It returns an error if there is an issue during the deletion process.
		DeleteFile(ctx context.Context, fileURL string) error
		// Generates a signed URL for accessing a file in Google Cloud Storage. It takes the file URL and the expiration duration for the signed URL as parameters. If the duration is 0, the signed URL will use the max duration.
		// It returns the generated signed URL if successful, or an error if there is an issue during the URL generation process.
		GenerateV4GetObjectSignedURL(fileURL string, expiration time.Duration) (string, error)
		// ParseGoogleStorageURL parses a Google Cloud Storage URL and returns the bucket name and object name.
		// It returns bucket name, object name, and an error if the URL is invalid.
		ParseGoogleStorageURL(fileURL string) (bucketName, filePath string, err error)
		// IsSignedURL checks if the provided file URL is a signed URL for a file in Google Cloud Storage. It takes the file URL as a parameter
		// It returns true if the URL is a signed URL, or false and an error if the URL is invalid.
		IsSignedURL(url string) (bool, error)
		// IsValidSignedURL checks if the provided file URL is a valid signed URL for a file in the project's Google Cloud Storage and haven't expired. It takes the file URL as a parameter
		// It returns true if the URL is a valid signed URL, or false and an error if the URL is invalid.
		IsValidSignedURL(ctx context.Context, url string) (bool, error)
		// GetDefaultFileURL generates a default file URL for accessing a file in Google Cloud Storage. It takes the file path as a parameter and returns the generated file URL.
		// This method is used to generate a default file URL to access files using this service.
		GetDefaultFileURL(filePath string) string
		// GetFileURL generates a file URL for accessing a file in Google Cloud Storage. It takes the file URL as a parameter and returns the generated file URL.
		// This method is used to generate a file URL to access files using this service. It can be used to generate a signed URL or a default URL based on the implementation.
		GetFileURL(ctx context.Context, bucketName, filePath string) string
		ParseSignedURLToFileURL(ctx context.Context, signedURL string) (string, error)
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
	}

	WhatsappBusiness interface {
		SendMessage(client *whatsapp_business.Client, to string, payload whatsapp_business.MessageComponent) (whatsapp_business.MessageResponse, int, error)
		GetTemplateList(client *whatsapp_business.Client) ([]whatsapp_business.TemplateResponse, int, error)
		UploadMedia(client *whatsapp_business.Client, fileBytes []byte, filename, mimeType string) (string, int, error)
		GetMediaURL(client *whatsapp_business.Client, mediaID string) (string, int, error)
		DownloadMedia(client *whatsapp_business.Client, mediaID string) ([]byte, http.Header, int, error)
		DeleteMedia(client *whatsapp_business.Client, mediaID string) (int, error)
		ResumableUpload(client *whatsapp_business.Client, inputData dto.ResumableUploadRequest) (string, int, error)
		CreateTemplate(client *whatsapp_business.Client, inputData dto.TemplateCreateRequest) (whatsapp_business.TemplateCreateResponse, int, error)
	}
)
