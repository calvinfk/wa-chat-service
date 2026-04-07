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
		MediaID      *string `json:"media_id"`
		OriginalName string  `json:"original_name"`
		MimeType     string  `json:"mime_type"`
		// URL          string `json:"url"`
		AccessURL *string `json:"access_url"`
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
)

func (r StorageMediaUploadResponse) FromModel(media model.StorageMedia, accessURL *string) StorageMediaUploadResponse {
	r.ID = media.DocumentID
	r.OriginalName = media.OriginalName
	r.MimeType = media.MimeType
	r.MediaID = media.MediaID
	// r.URL = media.URL
	r.AccessURL = accessURL
	return r
}

func (r StorageMediaSaveMediaIDResponse) FromModel(media model.StorageMedia) StorageMediaSaveMediaIDResponse {
	r.ID = media.DocumentID
	return r
}
