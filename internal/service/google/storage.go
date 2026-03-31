package google_service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/pkg/errs"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
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

func (s *GoogleStorageService) UploadReportAttachment(ctx context.Context, data *multipart.FileHeader, reportNo string) (string, string, error) {
	if data == nil {
		return "", "", errs.ErrGenericEmptyFile
	}
	file, err := data.Open()
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	fileData := make([]byte, data.Size)
	_, err = file.Read(fileData)
	if err != nil {
		return "", "", err
	}
	fileName := s.generateAttachmentName(reportNo, filepath.Ext(data.Filename))
	contentType := data.Header.Get("Content-Type")
	fileURL, err := s.UploadFile(ctx, fileData, s.cfg.AttachmentBucket, fileName, contentType)
	if err != nil {
		return "", "", err
	}
	return fileName, fileURL, nil
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

func (s *GoogleStorageService) GenerateV4GetObjectSignedURL(bucketName, objectName string) (string, error) {
	// 1. Define the permissions and duration
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(time.Duration(s.cfg.AttachmentLinkExpiry) * time.Second),
	}

	// 2. Generate the URL
	url, err := s.client.Bucket(bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("storage.SignedURL: %w", err)
	}

	return url, nil
}

func (s *GoogleStorageService) generateAttachmentName(reportNo, ext string) string {
	reportNoFormatted := strings.ReplaceAll(reportNo, "/", "_")
	uuidPart := uuid.New().String()
	return fmt.Sprintf("%s/%s%s", reportNoFormatted, uuidPart, ext)
}

func (s *GoogleStorageService) GenerateAttachmentURL(fileName string) (string, error) {
	fileURL, err := s.GenerateV4GetObjectSignedURL(s.cfg.AttachmentBucket, fileName)
	if err != nil {
		return "", err
	}
	return fileURL, nil
}

func (s *GoogleStorageService) GetAttachmentLinkExpiration() time.Duration {
	return time.Duration(s.cfg.AttachmentLinkExpiry) * time.Second
}
