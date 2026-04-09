package dto

import (
	"mime/multipart"
	"wa_chat_service/internal/model"

	"cloud.google.com/go/storage"
)

type (
	StorageMediaUploadRequest struct {
		File          []*multipart.FileHeader `form:"file" validate:"min_files=1,max_files=1"`
		PhoneNumberID string                  `form:"phone_number_id" validate:"required"`
	}

	StorageMediaUploadResponse struct {
		ID           string  `json:"id"`
		OriginalName string  `json:"original_name"`
		MimeType     string  `json:"mime_type"`
		AccessURL    *string `json:"access_url"`
	}
	StorageMediaGetRequest struct {
		ID string `query:"id" validate:"required,uuid"`
	}

	StorageMediaGetMediaResponse struct {
		// for now only return storage reader
		Reader      *storage.Reader `json:"-"`
		ContentType string          `json:"content_type"`
		FileName    string          `json:"file_name"`
		Size        int64           `json:"size"`
	}

	StorageMediaDeleteRequest struct {
		ID            string `query:"id" validate:"required_without=MediaID,omitempty,uuid"`
		MediaID       string `query:"media_id" validate:"required_without=ID"`
		PhoneNumberID string `query:"phone_number_id" validate:"required"`
	}

	StorageMediaSaveMediaIDRequest struct {
		MediaID       string `json:"media_id" validate:"required"`
		PhoneNumberID string `json:"phone_number_id" validate:"required"`
	}

	StorageMediaSaveMediaIDResponse struct {
		ID string `json:"id"`
	}

	StorageMediaResumableUploadRequest struct {
		StorageMediaID string `json:"storage_media_id" validate:"required,uuid"`
		PhoneNumberID  string `json:"phone_number_id" validate:"required"`
	}

	StorageMediaResumableUploadResponse struct {
		H string `json:"h"`
	}

	StorageMediaUploadMetaRequest struct {
		PhoneNumberID string `json:"phone_number_id" validate:"required"`
		ID            string `json:"id" validate:"required,uuid"`
	}

	StorageMediaUploadMetaResponse struct {
		MediaID string `json:"media_id"`
	}
)

func (StorageMediaUploadResponse) FromModel(media model.StorageMedia, accessURL *string) StorageMediaUploadResponse {
	return StorageMediaUploadResponse{
		ID:           media.DocumentID,
		OriginalName: media.OriginalName,
		MimeType:     media.MimeType,
		AccessURL:    accessURL,
	}
}

func (StorageMediaSaveMediaIDResponse) FromModel(media model.StorageMedia) StorageMediaSaveMediaIDResponse {
	return StorageMediaSaveMediaIDResponse{
		ID: media.DocumentID,
	}
}
