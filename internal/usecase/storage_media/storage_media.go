package storage_media_usecase

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/meta/whatsapp_business"

	"github.com/google/uuid"
)

type StorageMediaUsecase struct {
	storageMediaRepository repository.StorageMedia
	phoneNumberRepository  repository.PhoneNumber
	googleStorageService   service.GoogleStorage
	encryptService         service.Encrypt
	whatsappService        service.WhatsappService
}

func NewStorageMediaUsecase(storageMediaRepository repository.StorageMedia, phoneNumberRepository repository.PhoneNumber, googleStorageService service.GoogleStorage, encryptService service.Encrypt, whatsappService service.WhatsappService) *StorageMediaUsecase {
	return &StorageMediaUsecase{
		storageMediaRepository: storageMediaRepository,
		phoneNumberRepository:  phoneNumberRepository,
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
	files := inputData.File
	if len(files) == 0 {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] No file provided in the request")
		return response, true, fmt.Errorf("no file provided")
	}
	originalFileName := files[0].Filename
	filePath := "whatsapp-media/" + documentID.String() + filepath.Ext(files[0].Filename)
	file, err := files[0].Open()
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to open file:", err)
		return response, true, err
	}
	defer file.Close()
	fileData := make([]byte, files[0].Size)

	// read the whole file into fileData
	_, err = file.Read(fileData)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to read file data:", err)
		return response, true, err
	}

	// upload to firebase storage
	fileURL := u.googleStorageService.GetDefaultFileURL(filePath)
	_, err = u.googleStorageService.UploadFile(ctx, fileData, fileURL)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to upload file:", err)
		return response, true, err
	}
	url, err := u.googleStorageService.GenerateV4GetObjectSignedURL(fileURL, 0)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to generate attachment URL:", err)
		return response, true, err
	}
	media := model.StorageMedia{
		DocumentID:       documentID.String(),
		MediaID:          nil,
		URL:              &fileURL,
		OriginalName:     originalFileName,
		IsURLFromStorage: true,
		MimeType:         files[0].Header.Get("Content-Type"),
		CreatedAt:        time.Now(),
	}
	_, err = u.storageMediaRepository.Insert(ctx, nil, media)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] Failed to insert media data to repository:", err)
		return response, true, err
	}
	response = response.FromModel(media, &url)
	return response, false, nil
}

// TODO: if url not exists, check media id and stream that instead
func (u *StorageMediaUsecase) GetMedia(ctx context.Context, inputData dto.StorageMediaGetRequest) (dto.StorageMediaGetMediaResponse, bool, error) {
	var response dto.StorageMediaGetMediaResponse
	media, err := u.storageMediaRepository.GetByDocumentID(ctx, inputData.ID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][GetMedia] Failed to get media data from repository:", err)
		return response, true, err
	}
	if media.URL == nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][GetMedia] Media URL is nil")
		return response, true, fmt.Errorf("media URL is nil")
	}
	rc, attrs, err := u.googleStorageService.GetFile(ctx, *media.URL)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][GetMedia] Failed to get file from Google Cloud Storage:", err)
		return response, true, err
	}
	response.Reader = rc
	response.ContentType = attrs.ContentType
	response.FileName = attrs.Name
	response.Size = attrs.Size
	return response, false, nil
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
			return httpCode >= http.StatusInternalServerError, err
		}
	}
	if media.URL != nil {
		err = u.googleStorageService.DeleteFile(ctx, *media.URL)
		if err != nil {
			log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][DeleteMedia] Failed to delete file from Google Cloud Storage:", err)
			return true, err
		}
	}
	err = u.storageMediaRepository.Delete(ctx, nil, media.DocumentID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][DeleteMedia] Failed to delete media data from repository:", err)
		return true, err
	}
	return false, nil
}

func (u *StorageMediaUsecase) SaveMediaID(ctx context.Context, inputData dto.StorageMediaSaveMediaIDRequest) (dto.StorageMediaSaveMediaIDResponse, bool, error) {
	var emptyResponse dto.StorageMediaSaveMediaIDResponse
	phoneNumber, err := u.phoneNumberRepository.GetByPhoneNumberID(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to get phone number:", err)
		return emptyResponse, true, err
	}
	decyptedAccessToken, err := u.encryptService.Decrypt(phoneNumber.AccessToken)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to decrypt access token:", err)
		return emptyResponse, true, err
	}
	whatsappClient := whatsapp_business.New(decyptedAccessToken, phoneNumber.WabaID, phoneNumber.PhoneNumberID)
	url, httpCode, err := u.whatsappService.GetMediaURL(ctx, whatsappClient, inputData.MediaID)
	if err != nil {
		log.Printf("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMediaUsingMediaID] Failed to get media URL from WhatsApp Business API: %v", err)
		return emptyResponse, httpCode >= http.StatusInternalServerError, err
	}
	// download the file to get the mime type and original filename
	var originalFileName string
	urlHeaders, err := whatsappClient.GetHeaders(url)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMediaUsingMediaID] Failed to get media headers:", err)
		return emptyResponse, true, err
	}
	mimeType := urlHeaders.Get("Content-Type")
	contentDisposition := urlHeaders.Get("Content-Disposition")
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			originalFileName = params["filename"]
		}
	}
	if originalFileName == "" {
		originalFileName = fmt.Sprintf("%s.%s", inputData.MediaID, whatsapp_business.ParseMediaExtension(mimeType))
	}
	// store to repository
	mediaDocumentID, err := uuid.NewV7()
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMediaUsingMediaID] Failed to generate UUID:", err)
		return emptyResponse, true, err
	}
	media := model.StorageMedia{
		DocumentID:       mediaDocumentID.String(),
		OriginalName:     originalFileName,
		MediaID:          &inputData.MediaID,
		MimeType:         mimeType,
		IsURLFromStorage: false,
		CreatedAt:        time.Now(),
	}
	_, err = u.storageMediaRepository.Insert(ctx, nil, media)
	if err != nil {
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMediaUsingMediaID] Failed to create media record in repository:", err)
		return emptyResponse, true, err
	}
	return emptyResponse.FromModel(media), false, nil
}

func (u *StorageMediaUsecase) StoreMediaFromURL(ctx context.Context, mediaURL string) (model.StorageMedia, bool, error) {
	var emptyMedia model.StorageMedia
	// create new storage media record with original media link as access URL
	newMediaID, err := uuid.NewV7()
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to generate new media ID:", err)
		return emptyMedia, true, err
	}
	// download the file
	fileData, urlHeaders, err := formatter.DownloadFile(mediaURL)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to download media file:", err)
		return emptyMedia, true, err
	}
	// upload to firebase storage
	var filename string
	contentDisposition := urlHeaders.Get("Content-Disposition")
	if contentDisposition == "" {
		// check the url path for filename if Content-Disposition header is not present
		filename = formatter.GetFileNameFromURL(mediaURL)
	} else {
		// extract filename from Content-Disposition header
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			filename = params["filename"]
		}
	}
	if filename == "" {
		filename = fmt.Sprintf("%s.%s", newMediaID.String(), strings.Split(urlHeaders.Get("Content-Type"), "/")[1]) // default filename if not provided
	}
	filePath := "whatsapp-media/" + newMediaID.String() + filepath.Ext(filename)
	fileURL := u.googleStorageService.GetDefaultFileURL(filePath)
	attrs, err := u.googleStorageService.UploadFile(ctx, fileData, fileURL)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to upload media file to storage:", err)
		return emptyMedia, true, err
	}
	newMedia := model.StorageMedia{
		DocumentID:       newMediaID.String(),
		OriginalName:     filename,
		MimeType:         attrs.ContentType,
		URL:              &fileURL,
		IsURLFromStorage: true,
		CreatedAt:        time.Now(),
	}
	_, err = u.storageMediaRepository.Insert(ctx, nil, newMedia)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to insert new storage media record:", err)
	}
	return newMedia, false, nil
}
