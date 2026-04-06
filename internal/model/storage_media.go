package model

import (
	"fmt"
	"wa_chat_service/pkg/formatter"
)

type StorageMedia struct {
	DocumentID   string  `json:"__name__" firestore:"-"`        // generated uuid
	MediaID      *string `json:"media_id" firestore:"media_id"` // file ID returned by WhatsApp Business API after uploading media
	OriginalName string  `json:"original_name" firestore:"original_name"`
	MimeType     string  `json:"mime_type" firestore:"mime_type"`
	URL          string  `json:"url" firestore:"url"`
	AccessURL    string  `json:"access_url" firestore:"access_url"` // URL to access the media file, can be the same as URL or a different one depending on the storage solution
	CreatedAt    int64   `json:"created_at" firestore:"created_at"`
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
func (m StorageMedia) IsAccessURLExpired() (bool, error) {
	isExpired, err := formatter.IsGCSSignedURLExpired(m.AccessURL)
	if err != nil {
		return false, fmt.Errorf("failed to check access URL expiration: %w", err)
	}
	return isExpired, nil
}
