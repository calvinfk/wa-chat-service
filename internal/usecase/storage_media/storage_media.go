package storage_media_usecase

import (
	"context"
	"log"
	"path/filepath"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type StorageMediaUsecase struct {
	storageMediaRepository repository.StorageMedia
	firebaseService        service.GoogleFirebase
	googleStorageService   service.GoogleStorage
}

func NewStorageMediaUsecase(storageMediaRepository repository.StorageMedia, firebaseService service.GoogleFirebase, googleStorageService service.GoogleStorage) *StorageMediaUsecase {
	return &StorageMediaUsecase{
		storageMediaRepository: storageMediaRepository,
		firebaseService:        firebaseService,
		googleStorageService:   googleStorageService,
	}
}

func (u *StorageMediaUsecase) UploadMedia(ctx context.Context, inputData dto.StorageMediaUploadRequest) (dto.StorageMediaUploadResponse, bool, error) {
	var response dto.StorageMediaUploadResponse
	documentID, err := uuid.NewV7()
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to generate UUID:", err)
		return response, true, err
	}
	originalFileName := inputData.File.Filename
	// upload to firebase storage
	filePath := "whatsapp-media/" + documentID.String() + filepath.Ext(inputData.File.Filename)
	file, err := inputData.File.Open()
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to open file:", err)
		return response, true, err
	}
	defer file.Close()
	fileData := make([]byte, inputData.File.Size)
	_, err = file.Read(fileData)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to read file data:", err)
		return response, true, err
	}
	attrs, err := u.firebaseService.UploadFile(ctx, filePath, fileData)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to upload file:", err)
		return response, true, err
	}
	fileURL := "gs://" + attrs.Bucket + "/" + attrs.Name
	url, err := u.googleStorageService.GenerateV4GetObjectSignedURL(attrs.Bucket, attrs.Name, 0)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to generate attachment URL:", err)
		return response, true, err
	}
	media := model.StorageMedia{
		DocumentID:   documentID.String(),
		OriginalName: originalFileName,
		URL:          fileURL,
		AccessURL:    url,
	}
	_, err = u.storageMediaRepository.Insert(ctx, nil, media)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to insert media data to repository:", err)
		return response, true, err
	}

	response.FromModel(media)
	return response, false, nil
}

func (u *StorageMediaUsecase) GetMedia(ctx context.Context, inputData dto.StorageMediaGetRequest) (*storage.Reader, *storage.ObjectAttrs, bool, error) {
	media, err := u.storageMediaRepository.GetByDocumentID(ctx, inputData.ID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][GetMedia] Failed to get media data from repository:", err)
		return nil, nil, true, err
	}
	rc, attrs, err := u.googleStorageService.GetFile(ctx, media.URL)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][GetMedia] Failed to get file from Google Cloud Storage:", err)
		return nil, nil, true, err
	}
	defer rc.Close()

	fileData := make([]byte, attrs.Size)
	_, err = rc.Read(fileData)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][GetMedia] Failed to read file data:", err)
		return nil, nil, true, err
	}

	return rc, attrs, false, nil
}
