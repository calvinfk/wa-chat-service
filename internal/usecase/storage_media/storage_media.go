package storage_media_usecase

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/meta/whatsapp_business"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type StorageMediaUsecase struct {
	storageMediaRepository repository.StorageMedia
	phoneNumberRepository  repository.PhoneNumber
	firebaseService        service.GoogleFirebase
	googleStorageService   service.GoogleStorage
	encryptService         service.Encrypt
	whatsappService        service.WhatsappService
}

func NewStorageMediaUsecase(storageMediaRepository repository.StorageMedia, phoneNumberRepository repository.PhoneNumber, firebaseService service.GoogleFirebase, googleStorageService service.GoogleStorage, encryptService service.Encrypt, whatsappService service.WhatsappService) *StorageMediaUsecase {
	return &StorageMediaUsecase{
		storageMediaRepository: storageMediaRepository,
		phoneNumberRepository:  phoneNumberRepository,
		firebaseService:        firebaseService,
		googleStorageService:   googleStorageService,
		encryptService:         encryptService,
		whatsappService:        whatsappService,
	}
}

func (u *StorageMediaUsecase) UploadMedia(ctx context.Context, inputData dto.StorageMediaUploadRequest) (dto.StorageMediaUploadResponse, bool, error) {
	var response dto.StorageMediaUploadResponse
	documentID, err := uuid.NewV7()
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to generate UUID:", err)
		return response, true, err
	}
	phoneNumber, err := u.phoneNumberRepository.GetByPhoneNumberID(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to get phone number:", err)
		return response, true, err
	}
	decyptedAccessToken, err := u.encryptService.Decrypt(phoneNumber.AccessToken)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to decrypt access token:", err)
		return response, true, err
	}
	whatsappClient := whatsapp_business.New(decyptedAccessToken, phoneNumber.WabaID, phoneNumber.PhoneNumberID)
	originalFileName := inputData.File.Filename
	filePath := "whatsapp-media/" + documentID.String() + filepath.Ext(inputData.File.Filename)
	file, err := inputData.File.Open()
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to open file:", err)
		return response, true, err
	}
	defer file.Close()
	fileData := make([]byte, inputData.File.Size)
	// read the whole file into fileData
	_, err = file.Read(fileData)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to read file data:", err)
		return response, true, err
	}
	// upload media to WhatsApp Business API
	mediaID, httpCode, err := u.whatsappService.UploadMedia(ctx, whatsappClient, fileData, originalFileName, inputData.File.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to upload media to WhatsApp Business API (HTTP code: %d): %v", httpCode, err)
		return response, httpCode == http.StatusInternalServerError, err
	}

	// upload to firebase storage
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
		MediaID:      &mediaID,
		OriginalName: originalFileName,
		MimeType:     inputData.File.Header.Get("Content-Type"),
		URL:          fileURL,
		CreatedAt:    time.Now().Unix(),
	}
	_, err = u.storageMediaRepository.Insert(ctx, nil, media)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to insert media data to repository:", err)
		return response, true, err
	}

	response.FromModel(media, url)
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

func (u *StorageMediaUsecase) DeleteMedia(ctx context.Context, inputData dto.StorageMediaDeleteRequest) (bool, error) {
	phoneNumber, err := u.phoneNumberRepository.GetByPhoneNumberID(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to get phone number:", err)
		return true, err
	}
	decyptedAccessToken, err := u.encryptService.Decrypt(phoneNumber.AccessToken)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to decrypt access token:", err)
		return true, err
	}
	whatsappClient := whatsapp_business.New(decyptedAccessToken, phoneNumber.WabaID, phoneNumber.PhoneNumberID)
	var media *model.StorageMedia
	if inputData.MediaID != "" {
		mediaData, err := u.storageMediaRepository.GetByMediaID(ctx, inputData.MediaID)
		if err == nil {
			media = &mediaData
		} else if err != errs.ErrGenericNotFound {
			log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][DeleteMedia] Failed to get media data from repository:", err)
			return true, err
		}
	}
	if inputData.ID != "" {
		mediaData, err := u.storageMediaRepository.GetByDocumentID(ctx, inputData.ID)
		if err == nil {
			media = &mediaData
		} else if err != errs.ErrGenericNotFound {
			log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][DeleteMedia] Failed to get media data from repository:", err)
			return true, err
		}
	}
	if media == nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][DeleteMedia] Media not found with provided ID or MediaID")
		return true, errs.ErrGenericNotFound
	}
	if media.MediaID != nil {
		httpCode, err := u.whatsappService.DeleteMedia(ctx, whatsappClient, *media.MediaID)
		if err != nil {
			log.Printf("[ERROR][internal/usecase/storage_media/storage_media.go][DeleteMedia] Failed to delete media from WhatsApp Business API (HTTP code: %d): %v", httpCode, err)
			return httpCode == http.StatusInternalServerError, err
		}
	}
	err = u.storageMediaRepository.Delete(ctx, nil, media.DocumentID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][DeleteMedia] Failed to delete media data from repository:", err)
		return true, err
	}
	err = u.firebaseService.DeleteFile(ctx, media.URL)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][DeleteMedia] Failed to delete file from Firebase Storage:", err)
		return true, err
	}
	return false, nil
}
