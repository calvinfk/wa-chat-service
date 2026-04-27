package google_service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/internal/dto"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
)

type GoogleStorageService struct {
	client *storage.Client
	config *config.GCP
	zsLog  *zap.SugaredLogger
}

func NewGoogleStorageService(client *storage.Client, config *config.GCP, zsLog *zap.SugaredLogger) *GoogleStorageService {
	return &GoogleStorageService{
		client: client,
		config: config,
		zsLog:  zsLog,
	}
}

func (s *GoogleStorageService) UploadFile(ctx context.Context, fileData []byte, fileURL string) (*storage.ObjectAttrs, error) {
	if len(fileData) == 0 {
		s.zsLog.Error("[UploadFile] file data is empty")
		return nil, errs.ErrGenericEmptyFile
	}
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zsLog.Errorf("[UploadFile] failed to parse Google Storage URL: %v", err)
		return nil, err
	}
	bucket := s.client.Bucket(bucketName).UserProject(s.config.ProjectID)
	if _, err := bucket.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			s.zsLog.Errorf("[UploadFile] bucket %q does not exist", bucketName)
			return nil, errs.ErrGenericNotFound
		}
		s.zsLog.Errorf("[UploadFile] failed to get bucket: %v", err)
		return nil, err
	}
	obj := bucket.Object(filePath).If(storage.Conditions{DoesNotExist: true})
	writer := obj.NewWriter(ctx)
	// writer.ContentType = contentType

	if _, err := writer.Write(fileData); err != nil {
		s.zsLog.Errorf("[UploadFile] failed to write file: %v", err)
		return nil, err
	}

	if err := writer.Close(); err != nil {
		s.zsLog.Errorf("[UploadFile] failed to close file writer: %v", err)
		return nil, err
	}
	return writer.Attrs(), nil
}

func (s *GoogleStorageService) GetFile(ctx context.Context, fileURL string, rangeHeader string) (dto.StorageMediaGetMediaResponse, bool, error) {
	var response dto.StorageMediaGetMediaResponse
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zsLog.Errorf("[GetFileAttrs] failed to parse Google Storage URL: %v", err)
		return response, false, err
	}
	obj := s.client.Bucket(bucketName).UserProject(s.config.ProjectID).Object(filePath)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		s.zsLog.Errorf("[GetFileAttrs] failed to get file attributes: %v", err)
		return response, true, err
	}
	startRange, endRange, hasRange, err := utils.ParseRangeHeader(rangeHeader, attrs.Size)
	if err != nil {
		s.zsLog.Warnf("[GetMedia] Invalid range header: %q (size=%d)", rangeHeader, attrs.Size)
		return response, false, err
	}
	if hasRange {
		length := endRange - startRange + 1
		rcRange, err := obj.NewRangeReader(ctx, startRange, length)
		if err != nil {
			s.zsLog.Errorf("[GetMedia] Failed to get ranged file from Google Cloud Storage: %v", err)
			return response, true, err
		}
		response.Reader = rcRange
		response.Size = length
		response.StatusCode = http.StatusPartialContent
		response.ContentRange = fmt.Sprintf("bytes %d-%d/%d", startRange, endRange, attrs.Size)
	} else {
		rcFull, err := obj.NewReader(ctx)
		if err != nil {
			s.zsLog.Errorf("[GetMedia] Failed to get full file from Google Cloud Storage: %v", err)
			return response, true, err
		}
		response.Reader = rcFull
		response.Size = attrs.Size
		response.StatusCode = http.StatusOK
	}
	response.ContentType = attrs.ContentType
	response.FileName = attrs.Name
	return response, false, nil
}

func (s *GoogleStorageService) DeleteFile(ctx context.Context, fileURL string) error {
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zsLog.Errorf("[DeleteFile] failed to parse Google Storage URL: %v", err)
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
		s.zsLog.Errorf("[GenerateV4GetObjectSignedURL] error generating signed URL for file %q in bucket %q: %v", filePath, bucketName, err)
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
		s.zsLog.Errorf("[ParseGoogleStorageURL] invalid file URL: %s", fileURL)
		return "", "", fmt.Errorf("invalid file URL: %s", fileURL)
	}
	bucketName := parts[0]
	filePath := strings.Join(parts[1:], "/")
	if bucketName == "" || filePath == "" {
		s.zsLog.Errorf("[ParseGoogleStorageURL] invalid file URL: %s", fileURL)
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
// 		s.zsLog.Errorf("[isSignedURL] invalid signed URL: %s", url)
// 		return false, fmt.Errorf("invalid signed URL: %s", url)
// 	}
// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		s.zsLog.Errorf("[isSignedURL] error creating HTTP request: %v", err)
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
// 		s.zsLog.Errorf("[IsValidSignedURL] error validating signed URL: %v", err)
// 		return false, err
// 	} else if !isSigned {
// 		return false, nil
// 	}
// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		s.zsLog.Errorf("[IsValidSignedURL] error creating HTTP request: %v", err)
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
// 		s.zsLog.Errorf("[IsValidSignedURL] error fetching service account: %v", err)
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
// 		s.zsLog.Errorf("[IsValidSignedURL] error parsing expires duration: %v", err)
// 		return false, err
// 	}
// 	if expires <= 0 {
// 		return false, nil
// 	}
// 	date, err := time.Parse("20060102T150405Z", dateStr)
// 	if err != nil {
// 		s.zsLog.Errorf("[IsValidSignedURL] error parsing date: %v", err)
// 		return false, err
// 	}
// 	if time.Now().After(date.Add(expires)) {
// 		return false, nil
// 	}
// 	return true, nil
// }

// func (s *GoogleStorageService) ParseSignedURLToFileURL(ctx context.Context, signedURL string) (string, error) {
// 	if isValid, err := s.IsValidSignedURL(ctx, signedURL); err != nil {
// 		s.zsLog.Errorf("[ParseSignedURLToFileURL] error validating signed URL: %v", err)
// 		return "", err
// 	} else if !isValid {
// 		s.zsLog.Errorf("[ParseSignedURLToFileURL] invalid signed URL: %s", signedURL)
// 		return "", fmt.Errorf("invalid signed URL")
// 	}
// 	parts := strings.Split(strings.TrimPrefix(signedURL, "https://storage.googleapis.com/"), "/")
// 	if len(parts) < 2 {
// 		s.zsLog.Errorf("[ParseSignedURLToFileURL] invalid signed URL: %s", signedURL)
// 		return "", fmt.Errorf("invalid signed URL")
// 	}
// 	bucketName := parts[0]
// 	filePath := strings.Join(parts[1:], "/")
// 	filePath = strings.Split(filePath, "?")[0] // remove query parameters
// 	if bucketName == "" || filePath == "" {
// 		s.zsLog.Errorf("[ParseSignedURLToFileURL] invalid signed URL: missing bucket name or file path in URL: %s", signedURL)
// 		return "", fmt.Errorf("invalid signed URL: missing bucket name or file path")
// 	}
// 	return fmt.Sprintf("gs://%s/%s", bucketName, filePath), nil
// }
