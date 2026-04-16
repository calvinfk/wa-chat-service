package storage_media_usecase

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
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
	tenantRepository       repository.Tenant
	tenantUsecase          usecase.Tenant
	googleStorageService   service.GoogleStorage
	whatsappService        service.WhatsappBusiness
	zslog                  *zap.SugaredLogger
}

func NewStorageMediaUsecase(storageMediaRepository repository.StorageMedia, tenantRepository repository.Tenant, tenantUsecase usecase.Tenant, googleStorageService service.GoogleStorage, whatsappService service.WhatsappBusiness, zslog *zap.SugaredLogger) *StorageMediaUsecase {
	return &StorageMediaUsecase{
		storageMediaRepository: storageMediaRepository,
		tenantRepository:       tenantRepository,
		tenantUsecase:          tenantUsecase,
		googleStorageService:   googleStorageService,
		whatsappService:        whatsappService,
		zslog:                  zslog,
	}
}

func (u *StorageMediaUsecase) UploadMedia(ctx context.Context, inputData dto.StorageMediaUploadRequest) (dto.StorageMediaResponse, bool, error) {
	var response dto.StorageMediaResponse
	tenant, err := u.tenantRepository.GetByPhoneNumberID(ctx, inputData.PhoneNumberID)
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
		return response, true, fmt.Errorf("no file provided")
	}
	originalFileName := files[0].Filename
	filePath := "whatsapp-media/" + documentID.String() + filepath.Ext(files[0].Filename)
	file, err := files[0].Open()
	if err != nil {
		u.zslog.Errorf("[UploadMedia] Failed to open file: %v", err)
		return response, true, err
	}
	defer file.Close()
	fileData := make([]byte, files[0].Size)

	// read the whole file into fileData
	_, err = file.Read(fileData)
	if err != nil {
		u.zslog.Errorf("[UploadMedia] Failed to read file data: %v", err)
		return response, true, err
	}

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
	media := model.StorageMedia{
		DocumentID:       documentID.String(),
		TenantID:         tenant.DocumentID,
		URL:              &fileURL,
		OriginalName:     originalFileName,
		IsURLFromStorage: true,
		MimeType:         files[0].Header.Get("Content-Type"),
		CreatedAt:        time.Now(),
	}
	_, err = u.storageMediaRepository.Insert(ctx, nil, media)
	if err != nil {
		u.zslog.Errorf("[UploadMedia] Failed to insert media data to repository: %v", err)
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
	var media *model.StorageMedia
	if inputData.MediaID != "" {
		mediaData, err := u.storageMediaRepository.GetByMediaID(ctx, inputData.MediaID)
		if err == nil {
			media = &mediaData
		} else if err != errs.ErrGenericNotFound {
			u.zslog.Errorf("[DeleteMedia] Failed to get media data from repository: %v", err)
			return true, err
		}
	}
	if inputData.ID != "" {
		mediaData, err := u.storageMediaRepository.GetByDocumentID(ctx, inputData.ID)
		if err == nil {
			media = &mediaData
		} else if err != errs.ErrGenericNotFound {
			u.zslog.Errorf("[DeleteMedia] Failed to get media data from repository: %v", err)
			return true, err
		}
	}
	if media == nil {
		u.zslog.Errorf("[DeleteMedia] Media not found with provided ID or MediaID")
		return true, errs.ErrGenericNotFound
	}
	if media.MediaID != nil {
		httpCode, err := u.whatsappService.DeleteMedia(whatsappClient, *media.MediaID)
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
	url, httpCode, err := u.whatsappService.GetMediaURL(whatsappClient, inputData.MediaID)
	if err != nil {
		u.zslog.Errorf("[UploadMediaUsingMediaID] Failed to get media URL from WhatsApp Business API: %v", err)
		return emptyResponse, httpCode >= http.StatusInternalServerError, err
	}
	// download the file to get the mime type and original filename
	var originalFileName string
	urlHeaders, err := whatsappClient.GetHeaders(url)
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

func (u *StorageMediaUsecase) UploadResumableMedia(ctx context.Context, inputData dto.StorageMediaResumableUploadRequest) (dto.StorageMediaResumableUploadResponse, bool, error) {
	var response dto.StorageMediaResumableUploadResponse
	whatsappClient, _, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[UploadResumableMedia] Failed to get WhatsApp client: %v", err)
		return response, true, err
	}
	storageMedia, err := u.storageMediaRepository.GetByDocumentID(ctx, inputData.StorageMediaID)
	if err != nil {
		u.zslog.Errorf("[UploadResumableMedia] Failed to get media data from repository: %v", err)
		return response, true, err
	}
	if storageMedia.AssetHandle != nil {
		response.H = *storageMedia.AssetHandle
		return response, false, nil
	}
	// TODO: if the file is too large, we should read and upload it in chunks instead of reading the whole file into memory
	fileBytes, fileSize, err := u.downloadMedia(ctx, whatsappClient, storageMedia)
	if err != nil {
		u.zslog.Errorf("[UploadResumableMedia] Failed to download media: %v", err)
		return response, true, err
	}
	uploadSession, httpCode, err := whatsappClient.StartResumableUploadSession(storageMedia.OriginalName, fileSize, storageMedia.MimeType)
	if err != nil {
		u.zslog.Errorf("[UploadResumableMedia] Failed to upload media resumably: %v", err)
		return response, httpCode >= http.StatusInternalServerError, err
	}
	u.zslog.Infof("[UploadResumableMedia] Started resumable upload session: %+v", uploadSession)
	assetHandle, httpCode, err := whatsappClient.StartResumableUpload(uploadSession.ID, uploadSession.FileOffset, fileBytes)
	if err != nil {
		u.zslog.Errorf("[UploadResumableMedia] Failed to upload media to resumable session: %v", err)
		return response, httpCode >= http.StatusInternalServerError, err
	}
	u.zslog.Infof("[UploadResumableMedia] Completed resumable upload, got asset handle: %v", assetHandle.H)
	storageMedia.AssetHandle = &assetHandle.H
	err = u.storageMediaRepository.Update(ctx, nil, storageMedia)
	if err != nil {
		u.zslog.Errorf("[UploadResumableMedia] Failed to update media record in repository with MediaID: %v", err)
		return response, true, err
	}
	response.H = assetHandle.H
	return response, false, nil
}

func (u *StorageMediaUsecase) downloadMedia(ctx context.Context, whatsappClient *whatsapp_business.Client, media model.StorageMedia) ([]byte, int64, error) {
	if media.URL != nil {
		if media.IsURLFromStorage {
			rc, attrs, err := u.googleStorageService.GetFile(ctx, *media.URL)
			if err != nil {
				return nil, 0, err
			}
			defer rc.Close()
			fileByte, err := io.ReadAll(rc)
			if err != nil {
				return nil, 0, err
			}
			return fileByte, attrs.Size, nil
		} else {
			resp, err := http.Get(*media.URL)
			if err != nil {
				u.zslog.Errorf("[downloadMedia] Failed to download media from URL: %v", err)
				return nil, 0, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				u.zslog.Errorf("[downloadMedia] Failed to download media, HTTP status code: %d", resp.StatusCode)
				return nil, 0, fmt.Errorf("failed to download media, HTTP status code: %d", resp.StatusCode)
			}
			fileByte, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, 0, err
			}
			return fileByte, resp.ContentLength, nil
		}
	} else if media.MediaID != nil {
		fileBody, header, _, err := u.whatsappService.DownloadMedia(whatsappClient, *media.MediaID)
		if err != nil {
			u.zslog.Errorf("[downloadMedia] Failed to get media URL from WhatsApp Business API: %v", err)
			return nil, 0, fmt.Errorf("failed to get media URL from WhatsApp Business API: %w", err)
		}
		contentLength, err := strconv.ParseInt(header.Get("Content-Length"), 10, 64)
		if err != nil {
			u.zslog.Errorf("[downloadMedia] Failed to parse Content-Length header: %v", err)
			return nil, 0, fmt.Errorf("failed to parse Content-Length header: %w", err)
		}
		return fileBody, contentLength, nil
	}
	return nil, 0, fmt.Errorf("media not found or inaccessible")
}

func (u *StorageMediaUsecase) UploadMediaMeta(ctx context.Context, inputData dto.StorageMediaUploadMetaRequest) (dto.StorageMediaUploadMetaResponse, bool, error) {
	whatsappClient, _, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[UploadMediaMeta] Failed to get WhatsApp client: %v", err)
		return dto.StorageMediaUploadMetaResponse{}, true, err
	}
	media, err := u.storageMediaRepository.GetByDocumentID(ctx, inputData.ID)
	if err != nil {
		u.zslog.Errorf("[UploadMediaMeta] Failed to get media data from repository: %v", err)
		return dto.StorageMediaUploadMetaResponse{}, true, err
	}
	if media.MediaID != nil {
		// check if its still exists in WhatsApp Business API
		_, httpCode, err := u.whatsappService.GetMediaURL(whatsappClient, *media.MediaID)
		if err == nil {
			return dto.StorageMediaUploadMetaResponse{MediaID: *media.MediaID}, false, nil
		} else if httpCode >= http.StatusInternalServerError {
			u.zslog.Errorf("[UploadMediaMeta] Failed to get media URL from WhatsApp Business API: %v", err)
			return dto.StorageMediaUploadMetaResponse{}, true, err
		} else {
			// if the media is not found in WhatsApp Business API, we will upload it again
			fileBytes, _, err := u.downloadMedia(ctx, whatsappClient, media)
			if err != nil {
				u.zslog.Errorf("[UploadMediaMeta] Failed to download media for re-uploading: %v", err)
				return dto.StorageMediaUploadMetaResponse{}, true, err
			}
			mediaID, httpCode, err := u.whatsappService.UploadMedia(whatsappClient, fileBytes, media.OriginalName, media.MimeType)
			if err != nil {
				u.zslog.Errorf("[UploadMediaMeta] Failed to upload media to WhatsApp Business API: %v", err)
				return dto.StorageMediaUploadMetaResponse{}, httpCode >= http.StatusInternalServerError, err
			}
			// Update media record in repository with new MediaID
			media.MediaID = &mediaID
			err = u.storageMediaRepository.Update(ctx, nil, media)
			if err != nil {
				u.zslog.Errorf("[UploadMediaMeta] Failed to update media record in repository with new MediaID: %v", err)
				return dto.StorageMediaUploadMetaResponse{}, true, err
			}
			return dto.StorageMediaUploadMetaResponse{MediaID: mediaID}, false, nil
		}
	} else if media.URL != nil {
		// upload to WhatsApp Business API
		fileBytes, _, err := u.downloadMedia(ctx, whatsappClient, media)
		if err != nil {
			u.zslog.Errorf("[UploadMediaMeta] Failed to download media for uploading: %v", err)
			return dto.StorageMediaUploadMetaResponse{}, true, err
		}
		mediaID, httpCode, err := u.whatsappService.UploadMedia(whatsappClient, fileBytes, media.OriginalName, media.MimeType)
		if err != nil {
			u.zslog.Errorf("[UploadMediaMeta] Failed to upload media to WhatsApp Business API: %v", err)
			return dto.StorageMediaUploadMetaResponse{}, httpCode >= http.StatusInternalServerError, err
		}
		// Update media record in repository with new MediaID
		media.MediaID = &mediaID
		err = u.storageMediaRepository.Update(ctx, nil, media)
		if err != nil {
			u.zslog.Errorf("[UploadMediaMeta] Failed to update media record in repository with new MediaID: %v", err)
			return dto.StorageMediaUploadMetaResponse{}, true, err
		}
		return dto.StorageMediaUploadMetaResponse{MediaID: mediaID}, false, nil
	}
	return dto.StorageMediaUploadMetaResponse{}, true, fmt.Errorf("media not found or inaccessible")
}

func (u *StorageMediaUsecase) GetFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.StorageMediaGetListRequest]) (filter_request.FilterResponse[dto.StorageMediaResponse], bool, error) {
	response, err := u.storageMediaRepository.GetFiltered(ctx, inputData)
	if err != nil {
		u.zslog.Errorf("[GetFiltered] Failed to get filtered media data from repository: %v", err)
		return filter_request.FilterResponse[dto.StorageMediaResponse]{}, true, err
	}
	return response, false, nil
}
