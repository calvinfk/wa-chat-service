package storage_media_usecase

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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
	encryptService         service.Encrypt
	zslog                  *zap.SugaredLogger
	baseURL                string
	getEndpoint            string
	mediaUrlExpiryDuration time.Duration
}

type byteRange struct {
	start int64
	end   int64
}

func NewStorageMediaUsecase(
	storageMediaRepository repository.StorageMedia,
	tenantUsecase usecase.Tenant,
	googleStorageService service.GoogleStorage,
	whatsappService service.WhatsappBusiness,
	encryptService service.Encrypt,
	zslog *zap.SugaredLogger,
	baseURL string,
) *StorageMediaUsecase {
	return &StorageMediaUsecase{
		storageMediaRepository: storageMediaRepository,
		tenantUsecase:          tenantUsecase,
		googleStorageService:   googleStorageService,
		whatsappService:        whatsappService,
		encryptService:         encryptService,
		zslog:                  zslog,
		baseURL:                baseURL,
		getEndpoint:            "api/v1/storage-media/get",
		mediaUrlExpiryDuration: 30 * time.Second, // default expiry duration for media URLs
	}
}

func parseRangeHeader(rangeHeader string, totalSize int64) (byteRange, bool, error) {
	rangeHeader = strings.TrimSpace(rangeHeader)
	if rangeHeader == "" {
		return byteRange{}, false, nil
	}
	if totalSize <= 0 {
		return byteRange{}, false, errs.ErrGenericRangeNotSatisfiable
	}
	lowerHeader := strings.ToLower(rangeHeader)
	rangeSpec := rangeHeader
	if strings.HasPrefix(lowerHeader, "bytes=") {
		rangeSpec = strings.TrimSpace(rangeHeader[len("bytes="):])
	}
	if idx := strings.Index(rangeSpec, ","); idx >= 0 {
		rangeSpec = rangeSpec[:idx]
	}
	rangeSpec = strings.TrimSpace(rangeSpec)
	parts := strings.SplitN(rangeSpec, "-", 2)
	if len(parts) != 2 {
		return byteRange{}, false, errs.ErrGenericRangeNotSatisfiable
	}
	parts[0] = strings.TrimSpace(parts[0])
	parts[1] = strings.TrimSpace(parts[1])
	if parts[0] == "" {
		suffixLength, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || suffixLength <= 0 {
			return byteRange{}, false, errs.ErrGenericRangeNotSatisfiable
		}
		if suffixLength > totalSize {
			suffixLength = totalSize
		}
		return byteRange{start: totalSize - suffixLength, end: totalSize - 1}, true, nil
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= totalSize {
		return byteRange{}, false, errs.ErrGenericRangeNotSatisfiable
	}
	var end int64
	if parts[1] == "" {
		end = totalSize - 1
	} else {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end < start {
			return byteRange{}, false, errs.ErrGenericRangeNotSatisfiable
		}
		if end >= totalSize {
			end = totalSize - 1
		}
	}
	return byteRange{start: start, end: end}, true, nil
}

func (u *StorageMediaUsecase) UploadMedia(ctx context.Context, inputData dto.StorageMediaUploadRequest) (dto.StorageMediaResponse, bool, error) {
	var response dto.StorageMediaResponse
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClientByPhone(ctx, inputData.PhoneNumberID)
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
	// TODO: if the file is too large, this may cause memory issues. Consider streaming the file directly to WhatsApp Business API and/or Google Cloud Storage instead of reading it all into memory.
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
		url, err := u.generatePublicURL(media)
		if err != nil {
			u.zslog.Errorf("[UploadMedia] Failed to generate public URL for media (DocumentID: %s): %v", media.DocumentID, err)
		}
		media.IsURLFromStorage = true
		media.URL = &fileURL
		if url != "" {
			accessURL = &url
		}
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
func (u *StorageMediaUsecase) GetMedia(ctx context.Context, inputData dto.StorageMediaGetRequest, rangeHeader string) (dto.StorageMediaGetMediaResponse, bool, error) {
	var response dto.StorageMediaGetMediaResponse
	// Decrypt media payload to get the actual media document ID and validate expiry
	decrypted, err := u.encryptService.Decrypt(inputData.Media)
	if err != nil {
		u.zslog.Errorf("[GetMedia] Failed to decrypt media ID: %v", err)
		return response, false, errs.ErrGenericInvalidInput
	}
	splits := strings.SplitN(decrypted, ":", 2)
	if len(splits) != 2 {
		u.zslog.Errorf("[GetMedia] Invalid media ID format after decryption")
		return response, false, errs.ErrGenericInvalidInput
	}
	tUnix, err := strconv.ParseInt(splits[0], 10, 64)
	if err != nil {
		u.zslog.Errorf("[GetMedia] Failed to parse timestamp: %v", err)
		return response, false, errs.ErrGenericInvalidInput
	}
	expTime := time.Unix(tUnix, 0)
	if time.Now().After(expTime) {
		u.zslog.Errorf("[GetMedia] Media token has expired")
		return response, false, errs.ErrGenericGone
	}
	storageMediaID := splits[1]
	media, err := u.storageMediaRepository.GetByDocumentID(ctx, storageMediaID)
	if err != nil {
		u.zslog.Errorf("[GetMedia] Failed to get media data from repository: %v", err)
		return response, true, err
	}
	if media.URL != nil && media.IsURLFromStorage {
		// if media is stored in our storage, stream it from there
		attrs, err := u.googleStorageService.GetFileAttrs(ctx, *media.URL)
		if err != nil {
			u.zslog.Errorf("[GetMedia] Failed to get file attributes from Google Cloud Storage: %v", err)
			return response, true, err
		}
		requestedRange, hasRange, err := parseRangeHeader(rangeHeader, attrs.Size)
		if err != nil {
			u.zslog.Warnf("[GetMedia] Invalid range header: %q (size=%d)", rangeHeader, attrs.Size)
			return response, false, err
		}
		if hasRange {
			length := requestedRange.end - requestedRange.start + 1
			rcRange, attrs, err := u.googleStorageService.GetFileRange(ctx, *media.URL, requestedRange.start, length)
			if err != nil {
				u.zslog.Errorf("[GetMedia] Failed to get ranged file from Google Cloud Storage: %v", err)
				return response, true, err
			}
			response.Reader = rcRange
			response.Size = length
			response.StatusCode = http.StatusPartialContent
			response.ContentRange = fmt.Sprintf("bytes %d-%d/%d", requestedRange.start, requestedRange.end, attrs.Size)
			response.ContentType = attrs.ContentType
			response.FileName = attrs.Name
			response.ExpiresIn = u.mediaUrlExpiryDuration
			return response, false, nil
		}
		rcFull, attrs, err := u.googleStorageService.GetFile(ctx, *media.URL)
		if err != nil {
			u.zslog.Errorf("[GetMedia] Failed to get file from Google Cloud Storage: %v", err)
			return response, true, err
		}
		response.Reader = rcFull
		response.ContentType = attrs.ContentType
		response.FileName = attrs.Name
		response.Size = attrs.Size
		response.StatusCode = http.StatusOK
		response.ExpiresIn = u.mediaUrlExpiryDuration
	} else if media.URL != nil {
		// if URL exists but not from storage, stream it directly
		req, err := http.NewRequest(http.MethodGet, *media.URL, nil)
		if err != nil {
			u.zslog.Errorf("[GetMedia] Failed to get media from URL: %v", err)
			return response, true, err
		}
		if rangeHeader != "" {
			req.Header.Set("Range", rangeHeader)
		}
		rc, err := http.DefaultClient.Do(req)
		if err != nil {
			u.zslog.Errorf("[GetMedia] Failed to get media from URL: %v", err)
			return response, true, err
		}
		if rc.StatusCode != http.StatusOK && rc.StatusCode != http.StatusPartialContent {
			rc.Body.Close()
			u.zslog.Errorf("[GetMedia] Failed to get media from URL, status code: %d", rc.StatusCode)
			return response, rc.StatusCode >= http.StatusInternalServerError, fmt.Errorf("failed to get media from URL")
		}
		response.Reader = rc.Body
		response.ContentType = rc.Header.Get("Content-Type")
		response.FileName = media.OriginalName
		response.Size = rc.ContentLength
		response.StatusCode = rc.StatusCode
		response.ContentRange = rc.Header.Get("Content-Range")
		response.ExpiresIn = u.mediaUrlExpiryDuration // even if it's not from our storage, we can still set an expiry for caching purposes
	} else if media.MediaID != nil {
		// if media ID exists, get the media URL from WhatsApp Business API and stream it
		whatsappClient, _, err := u.tenantUsecase.GetWhatsappClientByTenant(ctx, media.TenantID)
		if err != nil {
			u.zslog.Errorf("[GetMedia] Failed to get WhatsApp client: %v", err)
			return response, true, err
		}
		url, httpCode, err := whatsappClient.GetMediaURL(*media.MediaID)
		if err != nil {
			u.zslog.Errorf("[GetMedia] Failed to get media URL from WhatsApp Business API: %v", err)
			return response, httpCode >= http.StatusInternalServerError, err
		}
		httpRes, err := whatsappClient.DownloadMedia(url.URL, rangeHeader)
		if err != nil {
			u.zslog.Errorf("[GetMedia] Failed to get media from URL: %v", err)
			return response, true, err
		}
		// defer httpRes.Body.Close()

		// bodyRes, err := io.ReadAll(httpRes.Body)
		// if err != nil {
		// 	u.zslog.Errorf("[GetMedia] Failed to read media body: %v", err)
		// 	return response, true, err
		// }
		// make reader
		// reader := bytes.NewReader(bodyRes)
		// response.Reader = io.NopCloser(reader)
		response.Reader = httpRes.Body
		response.ContentType = httpRes.Header.Get("Content-Type")
		response.FileName = media.OriginalName
		response.Size = httpRes.ContentLength
		response.StatusCode = httpRes.StatusCode
		response.ContentRange = httpRes.Header.Get("Content-Range")
		// response.Size = int64(len(bodyRes))
		// media from WhatsApp should have 5 minutes expiry as per WhatsApp's documentation
		// https://developers.facebook.com/documentation/business-messaging/whatsapp/reference/media/media-download-api#get-version-media-url
		response.ExpiresIn = 5 * time.Minute
	} else {
		u.zslog.Errorf("[GetMedia] No valid media source found for media (DocumentID: %s)", media.DocumentID)
		return response, false, errs.ErrGenericNotFound
	}

	return response, false, nil
}

func (u *StorageMediaUsecase) DeleteMedia(ctx context.Context, inputData dto.StorageMediaDeleteRequest) (bool, error) {
	whatsappClient, _, err := u.tenantUsecase.GetWhatsappClientByPhone(ctx, inputData.PhoneNumberID)
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
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClientByPhone(ctx, inputData.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[SaveMediaID] Failed to get WhatsApp client: %v", err)
		if err == errs.ErrGenericNotFound {
			return emptyResponse, false, err
		}
		return emptyResponse, true, err
	}
	url, httpCode, err := whatsappClient.GetMediaURL(inputData.MediaID)
	if err != nil {
		u.zslog.Errorf("[SaveMediaID] Failed to get media URL from WhatsApp Business API: %v", err)
		return emptyResponse, httpCode >= http.StatusInternalServerError, err
	}
	// download the file to get the mime type and original filename
	var originalFileName string
	urlHeaders, err := whatsappClient.GetHeaders(url.URL)
	if err != nil {
		u.zslog.Errorf("[SaveMediaID] Failed to get media headers: %v", err)
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
		u.zslog.Errorf("[SaveMediaID] Failed to generate UUID: %v", err)
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
		u.zslog.Errorf("[SaveMediaID] Failed to create media record in repository: %v", err)
		return emptyResponse, true, err
	}
	return dto.StorageMediaSaveMediaIDResponse{}.FromModel(media), false, nil
}

func (u *StorageMediaUsecase) GetFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.StorageMediaGetListRequest]) (filter_request.FilterResponse[dto.StorageMediaResponse], bool, error) {
	var response filter_request.FilterResponse[dto.StorageMediaResponse]
	data, paginate, totalItems, err := u.storageMediaRepository.GetFiltered(ctx, inputData)
	if err != nil {
		u.zslog.Errorf("[GetFiltered] Failed to get filtered media data from repository: %v", err)
		return filter_request.FilterResponse[dto.StorageMediaResponse]{}, true, err
	}
	var results []dto.StorageMediaResponse
	for i := range len(data) {
		var accessURL *string
		if data[i].URL != nil || data[i].MediaID != nil {
			generatedURL, err := u.generatePublicURL(data[i])
			if err != nil {
				u.zslog.Errorf("[GetFiltered] Failed to generate public URL for media (DocumentID: %s): %v", data[i].DocumentID, err)
			} else {
				accessURL = &generatedURL
			}
		}
		results = append(results, dto.StorageMediaResponse{}.FromModel(data[i], accessURL))
	}
	response = filter_request.NewFilterResponse(results, paginate, totalItems)
	return response, false, nil
}

func (u *StorageMediaUsecase) generatePublicURL(media model.StorageMedia) (string, error) {
	mediaLink, err := u.encryptService.Encrypt(fmt.Sprintf("%d:%s", time.Now().Add(u.mediaUrlExpiryDuration).Unix(), media.DocumentID))
	if err != nil {
		u.zslog.Errorf("[generatePublicURL] Failed to encrypt media link for media (DocumentID: %s): %v", media.DocumentID, err)
		return "", err
	}
	return fmt.Sprintf("%s/%s?media=%s", u.baseURL, u.getEndpoint, mediaLink), nil
}

func (u *StorageMediaUsecase) ParsePublicURL(url string) (string, error) {
	httpPrefix := fmt.Sprintf("%s/%s?media=", u.baseURL, u.getEndpoint)
	if !strings.HasPrefix(url, httpPrefix) {
		u.zslog.Errorf("[parsePublicURL] URL does not have expected prefix: %s", url)
		return "", errs.ErrGenericInvalidInput
	}
	return strings.TrimPrefix(url, httpPrefix), nil
}

func (u *StorageMediaUsecase) ParseMediaToken(mediaToken string) (string, bool, error) {
	decrypted, err := u.encryptService.Decrypt(mediaToken)
	if err != nil {
		u.zslog.Errorf("[ParseMediaToken] Failed to decrypt media token: %v", err)
		return "", true, err
	}
	splits := strings.SplitN(decrypted, ":", 2)
	if len(splits) != 2 {
		u.zslog.Errorf("[ParseMediaToken] Invalid decrypted media link format")
		return "", false, errs.ErrGenericInvalidInput
	}
	tUnix, err := strconv.ParseInt(splits[0], 10, 64)
	if err != nil {
		u.zslog.Errorf("[ParseMediaToken] Failed to parse timestamp from decrypted media link: %v", err)
		return "", true, err
	}
	expTime := time.Unix(tUnix, 0)
	if time.Now().After(expTime) {
		u.zslog.Errorf("[ParseMediaToken] Media link has expired")
		return "", false, errs.ErrGenericForbidden
	}
	storageMediaID := splits[1]
	return storageMediaID, false, nil
}
