package model

import (
	"time"
)

type StorageMedia struct {
	DocumentID       string    `json:"__name__" firestore:"-"`                // generated uuid
	TenantID         string    `json:"tenant_id" firestore:"tenant_id"`       // owner of the media
	MediaID          *string   `json:"media_id" firestore:"media_id"`         // file ID returned by WhatsApp Business API after uploading media
	URL              *string   `json:"url" firestore:"url"`                   // url created if media is uploaded to firebase storage
	AssetHandle      *string   `json:"asset_handle" firestore:"asset_handle"` // handle id by resumable api
	OriginalName     string    `json:"original_name" firestore:"original_name"`
	IsURLFromStorage bool      `json:"is_url_from_storage" firestore:"is_url_from_storage"`
	MimeType         string    `json:"mime_type" firestore:"mime_type"`
	CreatedAt        time.Time `json:"created_at" firestore:"created_at"`
}

func (m StorageMedia) TableName() string {
	return "storage_medias"
}

func (m StorageMedia) AllowedFilterFields() []string {
	return []string{"document_id", "media_id", "original_name", "mime_type", "url", "access_url", "created_at"}
}
func (m StorageMedia) AllowedSortFields() []string {
	return []string{"created_at"}
}
