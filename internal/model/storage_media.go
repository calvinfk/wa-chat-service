package model

import (
	"time"
)

type StorageMedia struct {
	DocumentID       string    `json:"id" firestore:"-"`                // generated uuid
	TenantID         string    `json:"tenant_id" firestore:"tenant_id"` // owner of the media
	OriginalName     string    `json:"original_name" firestore:"original_name"`
	MimeType         string    `json:"mime_type" firestore:"mime_type"`
	URL              *string   `json:"url" firestore:"url"`                                   // url created if media is uploaded to firebase storage
	MediaId          *string   `json:"media_id" firestore:"media_id,omitempty"`               // file ID returned by WhatsApp Business API after uploading media
	PhoneNumberId    *string   `json:"phone_number_id" firestore:"phone_number_id,omitempty"` // reference to phone number document, used to get WhatsApp client to download media from WhatsApp Business API
	AssetHandle      *string   `json:"asset_handle" firestore:"asset_handle,omitempty"`       // handle id by resumable api
	IsURLFromStorage bool      `json:"is_url_from_storage" firestore:"is_url_from_storage"`
	CreatedAt        time.Time `json:"created_at" firestore:"created_at"`
}

func (m StorageMedia) TableName() string {
	return "storage_medias"
}

func (m StorageMedia) AllowedFilterFields() []string {
	return []string{"tenant_id"}
}
func (m StorageMedia) AllowedSortFields() []string {
	return []string{"created_at"}
}
