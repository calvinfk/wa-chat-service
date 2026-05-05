package google_service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"wa_chat_service/config"
	"wa_chat_service/internal/dto"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
)

type googleStorageService struct {
	client *storage.Client
	config *config.GCP
	zsLog  *zap.SugaredLogger
}

func NewGoogleStorageService(client *storage.Client, config *config.GCP, zsLog *zap.SugaredLogger) *googleStorageService {
	return &googleStorageService{
		client: client,
		config: config,
		zsLog:  zsLog,
	}
}

func (s *googleStorageService) UploadFile(ctx context.Context, fileData []byte, fileURL string) (*storage.ObjectAttrs, error) {
	// Check if file data is empty before proceeding with upload
	if len(fileData) == 0 {
		s.zsLog.Error("[UploadFile] file data is empty")
		return nil, errs.ErrGenericEmptyFile
	}
	// Parse the file URL to extract bucket name and file path
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zsLog.Errorf("[UploadFile] failed to parse Google Storage URL: %v", err)
		return nil, err
	}

	// Check if the bucket exists before attempting to upload the file
	bucket := s.client.Bucket(bucketName).UserProject(s.config.ProjectID)
	if _, err := bucket.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			s.zsLog.Errorf("[UploadFile] bucket %q does not exist", bucketName)
			return nil, errs.ErrGenericNotFound
		}
		s.zsLog.Errorf("[UploadFile] failed to get bucket: %v", err)
		return nil, err
	}
	// Use the If condition to ensure that the file is only uploaded if it does not already exist in the bucket. This prevents overwriting existing files and ensures data integrity.
	obj := bucket.Object(filePath).If(storage.Conditions{DoesNotExist: true})
	// Create a new writer for the object and write the file data to it. The writer will automatically handle the upload process.
	// After writing the data, we close the writer to finalize the upload and return the attributes of the uploaded object.
	writer := obj.NewWriter(ctx)
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

func (s *googleStorageService) GetFile(ctx context.Context, fileURL string, rangeHeader string) (dto.StorageMediaGetMediaResponse, bool, error) {
	var response dto.StorageMediaGetMediaResponse
	// Parse the file URL to extract bucket name and file path.
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zsLog.Errorf("[GetFileAttrs] failed to parse Google Storage URL: %v", err)
		return response, false, err
	}
	// Create a reference to the object in Google Cloud Storage using the bucket name and file path.
	obj := s.client.Bucket(bucketName).UserProject(s.config.ProjectID).Object(filePath)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		s.zsLog.Errorf("[GetFileAttrs] failed to get file attributes: %v", err)
		return response, true, err
	}
	// Parse the Range header to determine if a partial content response is needed. If the Range header is valid and specifies a byte range, we will return only that portion of the file.
	// If the Range header is invalid or not provided, we will return the entire file.
	startRange, endRange, hasRange, err := utils.ParseRangeHeader(rangeHeader, attrs.Size)
	if err != nil {
		s.zsLog.Warnf("[GetMedia] Invalid range header: %q (size=%d)", rangeHeader, attrs.Size)
		return response, false, err
	}
	if hasRange {
		// Calculate the length of the requested byte range and create a new RangeReader for that range.
		// The RangeReader will allow us to read only the specified portion of the file, which is more efficient than reading the entire file when only a part of it is needed.
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
		// If no valid Range header is provided, we will return the entire file. We create a new reader for the object, which allows us to read the full content of the file from Google Cloud Storage.
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

func (s *googleStorageService) DeleteFile(ctx context.Context, fileURL string) error {
	bucketName, filePath, err := s.ParseGoogleStorageURL(fileURL)
	if err != nil {
		s.zsLog.Errorf("[DeleteFile] failed to parse Google Storage URL: %v", err)
		return err
	}
	obj := s.client.Bucket(bucketName).UserProject(s.config.ProjectID).Object(filePath)
	return obj.Delete(ctx)
}

func (s *googleStorageService) GetDefaultFileURL(filePath string) string {
	return fmt.Sprintf("gs://%s/%s", s.config.DefaultBucket, filePath)
}

func (s *googleStorageService) ParseGoogleStorageURL(fileURL string) (string, string, error) {
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
