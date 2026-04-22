package service

import (
	"context"
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/meta/whatsapp_business"

	"cloud.google.com/go/storage"
)

type (
	// AccessToken is an interface that defines methods for generating and parsing access tokens.
	AccessToken interface {
		// GenerateAccessToken generates an access token string for a given subject (sub).
		// It returns the generated access token string or an error if there is an issue during token generation.
		GenerateAccessToken(sub string) (string, error)
		// ParseAccessToken parses a AccessToken access token string and extracts the user ID (subject) from the token claims.
		// It validates the token using the configured JWK and returns the user ID as a UUID if the token is valid.
		// If the token is expired, it returns an error indicating that the token has expired and the sub.
		// If the token is invalid for any other reason, it returns only the error indicating that the token is invalid.
		ParseAccessTokenSub(tokenStr string) (string, error)
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
		// GetFileAttrs retrieves object attributes from Google Cloud Storage without opening a reader.
		GetFileAttrs(ctx context.Context, fileURL string) (*storage.ObjectAttrs, error)
		// GetFileRange retrieves a ranged reader for a file from Google Cloud Storage.
		GetFileRange(ctx context.Context, fileURL string, offset, length int64) (*storage.Reader, *storage.ObjectAttrs, error)
		// DeleteFile deletes a file from Google Cloud Storage. It takes the file URL as a parameter
		// It returns an error if there is an issue during the deletion process.
		DeleteFile(ctx context.Context, fileURL string) error
		// Generates a signed URL for accessing a file in Google Cloud Storage. It takes the file URL and the expiration duration for the signed URL as parameters. If the duration is 0, the signed URL will use the max duration.
		// It returns the generated signed URL if successful, or an error if there is an issue during the URL generation process.
		GenerateV4GetObjectSignedURL(fileURL string, expiration time.Duration) (string, error)
		// GetDefaultFileURL generates a default file URL for accessing a file in Google Cloud Storage. It takes the file path as a parameter and returns the generated file URL.
		// This method is used to generate a default file URL to access files using this service.
		GetDefaultFileURL(filePath string) string
	}

	// WhatsappBusiness is an interface that defines methods for helping with whatsapp business operations.
	WhatsappBusiness interface {
		// ValidateTemplatePayload validates the payload of a WhatsApp Business message template against the template definition stored in the database.
		// It takes a WhatsApp Business client, a template definition from the database, and the message components to be sent as parameters.
		// It returns an error if the payload is invalid according to the template definition, or nil if the payload is valid.
		ValidateTemplatePayload(client *whatsapp_business.Client, templateDB model.Template, templateSend whatsapp_business.MessageComponent) error
		// ExtractSendComponentParameterValues extracts the parameter values from the message components to be sent based on the parameter format defined in the template.
		// It takes the parameter format defined in the template and the message components to be sent as parameters.
		// It returns a map of parameter names to their corresponding values extracted from the message components, or an error if there is an issue during the extraction process.
		ExtractSendComponentParameterValues(parameterFormat string, sendComponents []map[string]any) (map[string]map[string]string, error)
		// ParseTemplateComponentParameter parses a template component parameter value based on the parameter format defined in the template.
		// It takes a parameter value as a string and returns the parsed parameter value as a string. The parsing logic is based on the parameter format defined in the template.
		ParseTemplateComponentParameter(value string) string
	}

	// GoogleTask is an interface that defines methods for managing Google Cloud Tasks.
	GoogleTask interface {
		// CreateBroadcastTask creates a new broadcast task in Google Cloud Tasks.
		CreateBroadcastTask(broadcastID string, scheduleTime time.Time) error
		// DeleteBroadcastTask deletes a broadcast task from Google Cloud Tasks based on the provided broadcast ID.
		DeleteBroadcastTask(broadcastID string) error
	}

	// JWT is an interface that defines methods for generating and parsing JSON Web Tokens.
	JWT interface {
		// GenerateJWT generates a JWT token string for a given subject (sub) and expiration time (expiredAt).
		// It returns the generated JWT token string or an error if there is an issue during token generation.
		GenerateJWT(sub any, expiredAt int64) (string, error)
		// ParseJWT parses a JWT token string and extracts the claims from the token. It validates the token using the configured JWK and returns the claims if the token is valid.
		// If the token is expired, it returns an error indicating that the token has expired and the claims.
		// If the token is invalid for any other reason, it returns only the error indicating that the token is invalid.
		ParseJWT(tokenString string) (any, error)
	}
)
