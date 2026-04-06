package google_service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/pkg/errs"

	"cloud.google.com/go/storage"
)

type GoogleStorageService struct {
	client *storage.Client
	cfg    *config.GCP
}

func NewGoogleStorageService(client *storage.Client, cfg *config.GCP) *GoogleStorageService {
	return &GoogleStorageService{
		client: client,
		cfg:    cfg,
	}
}

func (s *GoogleStorageService) UploadFile(ctx context.Context, fileData []byte, bucketID string, destinationPath string, contentType string) (string, error) {
	if len(fileData) == 0 {
		return "", errs.ErrGenericEmptyFile
	}
	bucket := s.client.Bucket(bucketID).UserProject(s.cfg.ProjectID)
	if _, err := bucket.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			return "", fmt.Errorf("bucket %q does not exist", bucketID)
		}
		return "", fmt.Errorf("failed to get bucket: %w", err)
	}

	obj := bucket.Object(destinationPath).If(storage.Conditions{DoesNotExist: true})
	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType
	_, err := writer.Write(fileData)
	if err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("gs://%s/%s", bucketID, destinationPath), nil
}
func (s *GoogleStorageService) GetFile(ctx context.Context, fileURL string) (*storage.Reader, *storage.ObjectAttrs, error) {
	// Parse the file URL to extract bucket and object names
	parts := strings.Split(strings.TrimPrefix(fileURL, "gs://"), "/")
	if len(parts) < 2 {
		return nil, nil, fmt.Errorf("invalid file URL: %s", fileURL)
	}
	bucketID := parts[0]
	filePath := strings.Join(parts[1:], "/")

	obj := s.client.Bucket(bucketID).UserProject(s.cfg.ProjectID).Object(filePath)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, nil, err
	}
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, nil, err
	}
	return reader, attrs, nil

}

func (s *GoogleStorageService) GenerateV4GetObjectSignedURL(bucketName, objectName string, expiration time.Duration) (string, error) {
	// 1. Define the permissions and duration
	maxV4 := 7 * 24 * time.Hour
	if expiration <= 0 || expiration > maxV4 {
		expiration = maxV4
	}

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	// 2. Generate the URL
	url, err := s.client.Bucket(bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("storage.SignedURL: %w", err)
	}

	return url, nil
}

func (s *GoogleStorageService) GenerateV4GetObjectSignedURLFromURL(fileURL string, expiration time.Duration) (string, error) {
	bucketName, objectName, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse Google Storage URL: %w", err)
	}
	return s.GenerateV4GetObjectSignedURL(bucketName, objectName, expiration)
}

func (s *GoogleStorageService) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	obj := s.client.Bucket(bucketName).UserProject(s.cfg.ProjectID).Object(objectName)
	return obj.Delete(ctx)
}

func (s *GoogleStorageService) DeleteFileByURL(ctx context.Context, fileURL string) error {
	bucketName, objectName, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		return fmt.Errorf("failed to parse Google Storage URL: %w", err)
	}
	return s.DeleteFile(ctx, bucketName, objectName)
}

func (s *GoogleStorageService) ParseGoogleStorageURL(fileURL string) (bucketName, objectName string, err error) {
	parts := strings.Split(strings.TrimPrefix(fileURL, "gs://"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid file URL: %s", fileURL)
	}
	bucketName = parts[0]
	objectName = strings.Join(parts[1:], "/")
	return bucketName, objectName, nil
}
