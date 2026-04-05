package dto

import (
	"mime/multipart"
	"wa_chat_service/internal/model"
)

type (
	StorageMediaUploadRequest struct {
		File          *multipart.FileHeader `form:"file" validate:"required,file"`
		PhoneNumberID string                `form:"phone_number_id" validate:"required"`
	}
	StorageMediaGetRequest struct {
		ID string `query:"id" validate:"required"`
	}

	StorageMediaUploadResponse struct {
		ID           string `json:"id"`
		MediaID      string `json:"media_id"`
		OriginalName string `json:"original_name"`
		MimeType     string `json:"mime_type"`
		// URL          string `json:"url"`
		AccessURL string `json:"access_url"`
	}

	StorageMediaDeleteRequest struct {
		MediaID       string `query:"media_id" validate:"required"`
		PhoneNumberID string `query:"phone_number_id" validate:"required"`
	}
)

func (r *StorageMediaUploadResponse) FromModel(media model.StorageMedia) {
	r.ID = media.DocumentID
	r.OriginalName = media.OriginalName
	r.MimeType = media.MimeType
	r.MediaID = media.MediaID
	// r.URL = media.URL
	r.AccessURL = media.AccessURL
}
