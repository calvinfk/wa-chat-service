package dto

import (
	"mime/multipart"
	"wa_chat_service/internal/model"
)

type (
	StorageMediaUploadRequest struct {
		File *multipart.FileHeader `form:"file" validate:"required,file"`
	}
	StorageMediaGetRequest struct {
		ID string `query:"id" validate:"required"`
	}

	StorageMediaUploadResponse struct {
		DocumentID   string `json:"document_id"`
		OriginalName string `json:"original_name"`
		// URL          string `json:"url"`
		AccessURL string `json:"access_url"`
	}
)

func (r *StorageMediaUploadResponse) FromModel(media model.StorageMedia) {
	r.DocumentID = media.DocumentID
	r.OriginalName = media.OriginalName
	// r.URL = media.URL
	r.AccessURL = media.AccessURL
}
