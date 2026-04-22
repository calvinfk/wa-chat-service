package google_service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/pkg/errs"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
)

type GoogleStorageService struct {
	client *storage.Client
	config *config.GCP
	zslog  *zap.SugaredLogger
}

func NewGoogleStorageService(client *storage.Client, config *config.GCP, zslog *zap.SugaredLogger) *GoogleStorageService {
	return &GoogleStorageService{
		client: client,
		config: config,
		zslog:  zslog,
	}
}

func (s *GoogleStorageService) UploadFile(ctx context.Context, fileData []byte, fileURL string) (*storage.ObjectAttrs, error) {
	if len(fileData) == 0 {
		s.zslog.Error("[UploadFile] file data is empty")
		return nil, errs.ErrGenericEmptyFile
	}
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zslog.Errorf("[UploadFile] failed to parse Google Storage URL: %v", err)
		return nil, err
	}
	bucket := s.client.Bucket(bucketName).UserProject(s.config.ProjectID)
	if _, err := bucket.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			s.zslog.Errorf("[UploadFile] bucket %q does not exist", bucketName)
			return nil, errs.ErrGenericNotFound
		}
		s.zslog.Errorf("[UploadFile] failed to get bucket: %v", err)
		return nil, err
	}
	obj := bucket.Object(filePath).If(storage.Conditions{DoesNotExist: true})
	writer := obj.NewWriter(ctx)
	// writer.ContentType = contentType

	if _, err := writer.Write(fileData); err != nil {
		s.zslog.Errorf("[UploadFile] failed to write file: %v", err)
		return nil, err
	}

	if err := writer.Close(); err != nil {
		s.zslog.Errorf("[UploadFile] failed to close file writer: %v", err)
		return nil, err
	}
	return writer.Attrs(), nil
}

func (s *GoogleStorageService) GetFile(ctx context.Context, fileURL string) (*storage.Reader, *storage.ObjectAttrs, error) {
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zslog.Errorf("[GetFile] failed to parse Google Storage URL: %v", err)
		return nil, nil, err
	}
	obj := s.client.Bucket(bucketName).UserProject(s.config.ProjectID).Object(filePath)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		s.zslog.Errorf("[GetFile] failed to get file attributes: %v", err)
		return nil, nil, err
	}
	reader, err := obj.NewReader(ctx)
	if err != nil {
		s.zslog.Errorf("[GetFile] failed to create file reader: %v", err)
		return nil, nil, err
	}
	return reader, attrs, nil
}

func (s *GoogleStorageService) GetFileAttrs(ctx context.Context, fileURL string) (*storage.ObjectAttrs, error) {
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zslog.Errorf("[GetFileAttrs] failed to parse Google Storage URL: %v", err)
		return nil, err
	}
	obj := s.client.Bucket(bucketName).UserProject(s.config.ProjectID).Object(filePath)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		s.zslog.Errorf("[GetFileAttrs] failed to get file attributes: %v", err)
		return nil, err
	}
	return attrs, nil
}

func (s *GoogleStorageService) GetFileRange(ctx context.Context, fileURL string, offset, length int64) (*storage.Reader, *storage.ObjectAttrs, error) {
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zslog.Errorf("[GetFileRange] failed to parse Google Storage URL: %v", err)
		return nil, nil, err
	}
	obj := s.client.Bucket(bucketName).UserProject(s.config.ProjectID).Object(filePath)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		s.zslog.Errorf("[GetFileRange] failed to get file attributes: %v", err)
		return nil, nil, err
	}
	reader, err := obj.NewRangeReader(ctx, offset, length)
	if err != nil {
		s.zslog.Errorf("[GetFileRange] failed to create file range reader: %v", err)
		return nil, nil, err
	}
	return reader, attrs, nil
}

func (s *GoogleStorageService) DeleteFile(ctx context.Context, fileURL string) error {
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zslog.Errorf("[DeleteFile] failed to parse Google Storage URL: %v", err)
		return err
	}
	obj := s.client.Bucket(bucketName).UserProject(s.config.ProjectID).Object(filePath)
	return obj.Delete(ctx)
}

func (s *GoogleStorageService) GenerateV4GetObjectSignedURL(fileURL string, expiration time.Duration) (string, error) {
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse Google Storage URL: %w", err)
	}
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
	url, err := s.client.Bucket(bucketName).SignedURL(filePath, opts)
	if err != nil {
		s.zslog.Errorf("[GenerateV4GetObjectSignedURL] error generating signed URL for file %q in bucket %q: %v", filePath, bucketName, err)
		return "", err
	}

	return url, nil
}
func (s *GoogleStorageService) GetDefaultFileURL(filePath string) string {
	return fmt.Sprintf("gs://%s/%s", s.config.DefaultBucket, filePath)
}

func (s *GoogleStorageService) ParseGoogleStorageURL(fileURL string) (string, string, error) {
	parts := strings.Split(strings.TrimPrefix(fileURL, "gs://"), "/")
	if len(parts) < 2 {
		s.zslog.Errorf("[ParseGoogleStorageURL] invalid file URL: %s", fileURL)
		return "", "", fmt.Errorf("invalid file URL: %s", fileURL)
	}
	bucketName := parts[0]
	filePath := strings.Join(parts[1:], "/")
	if bucketName == "" || filePath == "" {
		s.zslog.Errorf("[ParseGoogleStorageURL] invalid file URL: %s", fileURL)
		return "", "", fmt.Errorf("invalid file URL: %s", fileURL)
	}
	return bucketName, filePath, nil
}

// func (s *GoogleStorageService) isSignedURL(url string) (bool, error) {
// 	if !strings.HasPrefix(url, "https://storage.googleapis.com/") {
// 		return false, nil
// 	}
// 	parts := strings.Split(strings.TrimPrefix(url, "https://storage.googleapis.com/"), "/")
// 	if len(parts) < 2 {
// 		s.zslog.Errorf("[isSignedURL] invalid signed URL: %s", url)
// 		return false, fmt.Errorf("invalid signed URL: %s", url)
// 	}
// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		s.zslog.Errorf("[isSignedURL] error creating HTTP request: %v", err)
// 		return false, err
// 	}
// 	queries := req.URL.Query()
// 	var requiredParams = []string{
// 		"X-Goog-Algorithm",
// 		"X-Goog-Credential",
// 		"X-Goog-Date",
// 		"X-Goog-Expires",
// 		"X-Goog-Signature",
// 		"X-Goog-SignedHeaders",
// 	}
// 	for _, param := range requiredParams {
// 		if _, ok := queries[param]; !ok {
// 			return false, nil
// 		}
// 	}
// 	return true, nil
// }

// func (s *GoogleStorageService) IsValidSignedURL(ctx context.Context, url string) (bool, error) {
// 	if isSigned, err := s.isSignedURL(url); err != nil {
// 		s.zslog.Errorf("[IsValidSignedURL] error validating signed URL: %v", err)
// 		return false, err
// 	} else if !isSigned {
// 		return false, nil
// 	}
// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		s.zslog.Errorf("[IsValidSignedURL] error creating HTTP request: %v", err)
// 		return false, err
// 	}
// 	queries := req.URL.Query()
// 	googCredential := req.Header.Get("X-Goog-Credential")
// 	if googCredential == "" {
// 		return false, nil
// 	}
// 	// check if credential contains service account email
// 	clientEmail, err := s.client.ServiceAccount(ctx, s.config.ProjectID)
// 	if err != nil {
// 		s.zslog.Errorf("[IsValidSignedURL] error fetching service account: %v", err)
// 		return false, err
// 	}
// 	if !strings.Contains(googCredential, clientEmail) {
// 		return false, nil
// 	}
// 	// check expires parameter
// 	expiresStr := queries.Get("X-Goog-Expires")
// 	dateStr := queries.Get("X-Goog-Date")
// 	if expiresStr == "" || dateStr == "" {
// 		return false, nil
// 	}
// 	expires, err := time.ParseDuration(expiresStr + "s")
// 	if err != nil {
// 		s.zslog.Errorf("[IsValidSignedURL] error parsing expires duration: %v", err)
// 		return false, err
// 	}
// 	if expires <= 0 {
// 		return false, nil
// 	}
// 	date, err := time.Parse("20060102T150405Z", dateStr)
// 	if err != nil {
// 		s.zslog.Errorf("[IsValidSignedURL] error parsing date: %v", err)
// 		return false, err
// 	}
// 	if time.Now().After(date.Add(expires)) {
// 		return false, nil
// 	}
// 	return true, nil
// }

// func (s *GoogleStorageService) ParseSignedURLToFileURL(ctx context.Context, signedURL string) (string, error) {
// 	if isValid, err := s.IsValidSignedURL(ctx, signedURL); err != nil {
// 		s.zslog.Errorf("[ParseSignedURLToFileURL] error validating signed URL: %v", err)
// 		return "", err
// 	} else if !isValid {
// 		s.zslog.Errorf("[ParseSignedURLToFileURL] invalid signed URL: %s", signedURL)
// 		return "", fmt.Errorf("invalid signed URL")
// 	}
// 	parts := strings.Split(strings.TrimPrefix(signedURL, "https://storage.googleapis.com/"), "/")
// 	if len(parts) < 2 {
// 		s.zslog.Errorf("[ParseSignedURLToFileURL] invalid signed URL: %s", signedURL)
// 		return "", fmt.Errorf("invalid signed URL")
// 	}
// 	bucketName := parts[0]
// 	filePath := strings.Join(parts[1:], "/")
// 	filePath = strings.Split(filePath, "?")[0] // remove query parameters
// 	if bucketName == "" || filePath == "" {
// 		s.zslog.Errorf("[ParseSignedURLToFileURL] invalid signed URL: missing bucket name or file path in URL: %s", signedURL)
// 		return "", fmt.Errorf("invalid signed URL: missing bucket name or file path")
// 	}
// 	return fmt.Sprintf("gs://%s/%s", bucketName, filePath), nil
// }
