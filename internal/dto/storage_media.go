package dto

import (
	"io"
	"mime/multipart"
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"

	"github.com/go-playground/validator/v10"
)

type (
	StorageMediaUploadRequest struct {
		PhoneNumberId *string                 `query:"phone_number_id" validate:"required_with=SaveMeta"`
		SaveResumable *bool                   `query:"save_resumable" validate:"required_without_all=SaveMeta SaveStorage"`
		SaveMeta      *bool                   `query:"save_meta" validate:"required_without_all=SaveResumable SaveStorage"`
		SaveStorage   *bool                   `query:"save_storage" validate:"required_without_all=SaveResumable SaveMeta"`
		File          []*multipart.FileHeader `form:"file" validate:"min_files=1,max_files=1"`
	}

	StorageMediaResponse struct {
		ID           string  `json:"id"`
		MediaId      *string `json:"media_id,omitempty"`
		AssetHandle  *string `json:"asset_handle,omitempty"`
		OriginalName string  `json:"original_name"`
		MimeType     string  `json:"mime_type"`
		AccessURL    *string `json:"access_url"`
	}
	StorageMediaGetRequest struct {
		Media          string `query:"media" validate:"required"` // can be either encrypted media id or media URL
		StorageMediaID *string
		Url            *string
	}

	StorageMediaGetMediaResponse struct {
		Reader       io.ReadCloser `json:"-"`
		ContentType  string        `json:"content_type"`
		FileName     string        `json:"file_name"`
		Size         int64         `json:"size"`
		ExpiresIn    time.Duration `json:"expires_in"`
		StatusCode   int           `json:"status_code"`
		ContentRange string        `json:"content_range,omitempty"`
	}
	StorageMediaDeleteRequest struct {
		ID string `query:"id" validate:"required,uuid"`
	}

	StorageMediaSaveMediaIDRequest struct {
		MediaId       string `json:"media_id" validate:"required"`
		PhoneNumberId string `json:"phone_number_id" validate:"required"`
	}

	StorageMediaSaveMediaIDResponse struct {
		ID string `json:"id"`
	}

	StorageMediaGetListRequest struct {
	}

	StorageMediaEncryptLinkRequest struct {
		Link string `json:"link" validate:"required,url"`
	}
)

func (StorageMediaResponse) FromModel(media model.StorageMedia, accessURL *string) StorageMediaResponse {
	return StorageMediaResponse{
		ID:           media.DocumentID,
		OriginalName: media.OriginalName,
		MimeType:     media.MimeType,
		AccessURL:    accessURL,
		MediaId:      media.MediaId,
		AssetHandle:  media.AssetHandle,
	}
}

func (StorageMediaSaveMediaIDResponse) FromModel(media model.StorageMedia) StorageMediaSaveMediaIDResponse {
	return StorageMediaSaveMediaIDResponse{
		ID: media.DocumentID,
	}
}

func (r StorageMediaGetListRequest) Validate() map[string]string {
	validator := validator.New()
	if err := validator.Struct(r); err != nil {
		return utils.GetValidatorErrorMessages(err)
	}
	return nil
}
