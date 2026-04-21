package storage_media_usecase

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/meta/whatsapp_business"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type StorageMediaUsecase struct {
	storageMediaRepository repository.StorageMedia
	tenantUsecase          usecase.Tenant
	googleStorageService   service.GoogleStorage
	whatsappService        service.WhatsappBusiness
	zslog                  *zap.SugaredLogger
}

func NewStorageMediaUsecase(storageMediaRepository repository.StorageMedia, tenantUsecase usecase.Tenant, googleStorageService service.GoogleStorage, whatsappService service.WhatsappBusiness, zslog *zap.SugaredLogger) *StorageMediaUsecase {
	return &StorageMediaUsecase{
		storageMediaRepository: storageMediaRepository,
		tenantUsecase:          tenantUsecase,
		googleStorageService:   googleStorageService,
		whatsappService:        whatsappService,
		zslog:                  zslog,
	}
}

func (u *StorageMediaUsecase) UploadMedia(ctx context.Context, inputData dto.StorageMediaUploadRequest) (dto.StorageMediaResponse, bool, error) {
	var response dto.StorageMediaResponse
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[UploadMedia] Failed to get tenant data from repository: %v", err)
		if err == errs.ErrGenericNotFound {
			return response, false, err
		}
		return response, true, err
	}
	documentID, err := uuid.NewV7()
	if err != nil {
		u.zslog.Errorf("[UploadMedia] Failed to generate UUID: %v", err)
		return response, true, err
	}
	files := inputData.File
	if len(files) == 0 {
		u.zslog.Errorf("[UploadMedia] No file provided in the request")
		return response, false, fmt.Errorf("no file provided")
	}
	originalFileName := files[0].Filename
	filePath := "whatsapp-media/" + documentID.String() + filepath.Ext(files[0].Filename)
	file, err := files[0].Open()
	if err != nil {
		u.zslog.Errorf("[UploadMedia] Failed to open file: %v", err)
		return response, true, err
	}
	defer file.Close()
	fileSize := files[0].Size
	fileData := make([]byte, fileSize)

	// read the whole file into fileData
	_, err = file.Read(fileData)
	if err != nil {
		u.zslog.Errorf("[UploadMedia] Failed to read file data: %v", err)
		return response, true, err
	}
	media := model.StorageMedia{
		DocumentID:   documentID.String(),
		TenantID:     tenantID,
		OriginalName: originalFileName,
		MimeType:     files[0].Header.Get("Content-Type"),
		CreatedAt:    time.Now(),
	}
	var accessURL *string

	if inputData.SaveMeta {
		mediaResponse, httpCode, err := whatsappClient.UploadMedia(fileData, media.OriginalName, media.MimeType)
		if err != nil {
			u.zslog.Errorf("[UploadMedia] Failed to upload media to WhatsApp Business API: %v", err)
			return response, httpCode >= http.StatusInternalServerError, err
		}
		media.MediaID = &mediaResponse.ID
	}
	if inputData.SaveResumable {
		uploadSession, httpCode, err := whatsappClient.StartResumableUploadSession(media.OriginalName, fileSize, media.MimeType)
		if err != nil {
			u.zslog.Errorf("[UploadMedia] Failed to upload media resumably: %v", err)
			return response, httpCode >= http.StatusInternalServerError, err
		}
		assetHandle, httpCode, err := whatsappClient.StartResumableUpload(uploadSession.ID, uploadSession.FileOffset, fileData)
		if err != nil {
			u.zslog.Errorf("[UploadMedia] Failed to upload media to resumable session: %v", err)
			return response, httpCode >= http.StatusInternalServerError, err
		}
		media.AssetHandle = &assetHandle.H
	}
	if inputData.SaveStorage {
		// upload to firebase storage
		fileURL := u.googleStorageService.GetDefaultFileURL(filePath)
		_, err = u.googleStorageService.UploadFile(ctx, fileData, fileURL)
		if err != nil {
			u.zslog.Errorf("[UploadMedia] Failed to upload file: %v", err)
			return response, true, err
		}
		url, err := u.googleStorageService.GenerateV4GetObjectSignedURL(fileURL, 0)
		if err != nil {
			u.zslog.Errorf("[UploadMedia] Failed to generate attachment URL: %v", err)
			return response, true, err
		}
		media.IsURLFromStorage = true
		media.URL = &fileURL
		accessURL = &url
	}
	_, err = u.storageMediaRepository.Insert(ctx, nil, media)
	if err != nil {
		u.zslog.Errorf("[UploadMedia] Failed to insert media data to repository: %v", err)
		return response, true, err
	}
	response = response.FromModel(media, accessURL)
	return response, false, nil
}

// TODO: if url not exists, check media id and stream that instead
func (u *StorageMediaUsecase) GetMedia(ctx context.Context, inputData dto.StorageMediaGetRequest) (dto.StorageMediaGetMediaResponse, bool, error) {
	var response dto.StorageMediaGetMediaResponse
	media, err := u.storageMediaRepository.GetByDocumentID(ctx, inputData.ID)
	if err != nil {
		u.zslog.Errorf("[GetMedia] Failed to get media data from repository: %v", err)
		return response, true, err
	}
	if media.URL == nil {
		u.zslog.Errorf("[GetMedia] Media URL is nil")
		return response, true, fmt.Errorf("media URL is nil")
	} else if !media.IsURLFromStorage {
		u.zslog.Errorf("[GetMedia] Media URL is not from storage, cannot be accessed")
		return response, true, fmt.Errorf("media URL is not from storage, cannot be accessed")
	}

	rc, attrs, err := u.googleStorageService.GetFile(ctx, *media.URL)
	if err != nil {
		u.zslog.Errorf("[GetMedia] Failed to get file from Google Cloud Storage: %v", err)
		return response, true, err
	}
	response.Reader = rc
	response.ContentType = attrs.ContentType
	response.FileName = attrs.Name
	response.Size = attrs.Size
	return response, false, nil
}

func (u *StorageMediaUsecase) DeleteMedia(ctx context.Context, inputData dto.StorageMediaDeleteRequest) (bool, error) {
	whatsappClient, _, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[DeleteMedia] Failed to get WhatsApp client: %v", err)
		return true, err
	}
	media, err := u.storageMediaRepository.GetByDocumentID(ctx, inputData.ID)
	if err != nil {
		u.zslog.Errorf("[DeleteMedia] Failed to get media data from repository: %v", err)
		if err == errs.ErrGenericNotFound {
			return false, err
		}
		return true, err
	}
	if media.MediaID != nil {
		_, httpCode, err := whatsappClient.DeleteMedia(*media.MediaID)
		if err != nil {
			u.zslog.Errorf("[DeleteMedia] Failed to delete media from WhatsApp Business API (HTTP code: %d): %v", httpCode, err)
			return httpCode >= http.StatusInternalServerError, err
		}
	}
	if media.IsURLFromStorage && media.URL != nil {
		err = u.googleStorageService.DeleteFile(ctx, *media.URL)
		if err != nil {
			u.zslog.Errorf("[DeleteMedia] Failed to delete file from Google Cloud Storage: %v", err)
			return true, err
		}
	}
	err = u.storageMediaRepository.Delete(ctx, nil, media.DocumentID)
	if err != nil {
		u.zslog.Errorf("[DeleteMedia] Failed to delete media data from repository: %v", err)
		return true, err
	}
	return false, nil
}

func (u *StorageMediaUsecase) SaveMediaID(ctx context.Context, inputData dto.StorageMediaSaveMediaIDRequest) (dto.StorageMediaSaveMediaIDResponse, bool, error) {
	var emptyResponse dto.StorageMediaSaveMediaIDResponse
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[UploadMediaUsingMediaID] Failed to get WhatsApp client: %v", err)
		if err == errs.ErrGenericNotFound {
			return emptyResponse, false, err
		}
		return emptyResponse, true, err
	}
	url, httpCode, err := whatsappClient.GetMediaURL(inputData.MediaID)
	if err != nil {
		u.zslog.Errorf("[UploadMediaUsingMediaID] Failed to get media URL from WhatsApp Business API: %v", err)
		return emptyResponse, httpCode >= http.StatusInternalServerError, err
	}
	// download the file to get the mime type and original filename
	var originalFileName string
	urlHeaders, err := whatsappClient.GetHeaders(url.URL)
	if err != nil {
		u.zslog.Errorf("[UploadMediaUsingMediaID] Failed to get media headers: %v", err)
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
		u.zslog.Errorf("[UploadMediaUsingMediaID] Failed to generate UUID: %v", err)
		return emptyResponse, true, err
	}
	media := model.StorageMedia{
		DocumentID:       mediaDocumentID.String(),
		OriginalName:     originalFileName,
		TenantID:         tenantID,
		MediaID:          &inputData.MediaID,
		MimeType:         mimeType,
		IsURLFromStorage: false,
		CreatedAt:        time.Now(),
	}
	_, err = u.storageMediaRepository.Insert(ctx, nil, media)
	if err != nil {
		u.zslog.Errorf("[UploadMediaUsingMediaID] Failed to create media record in repository: %v", err)
		return emptyResponse, true, err
	}
	return dto.StorageMediaSaveMediaIDResponse{}.FromModel(media), false, nil
}

func (u *StorageMediaUsecase) GetFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.StorageMediaGetListRequest]) (filter_request.FilterResponse[dto.StorageMediaResponse], bool, error) {
	response, err := u.storageMediaRepository.GetFiltered(ctx, inputData)
	if err != nil {
		u.zslog.Errorf("[GetFiltered] Failed to get filtered media data from repository: %v", err)
		return filter_request.FilterResponse[dto.StorageMediaResponse]{}, true, err
	}
	return response, false, nil
}
