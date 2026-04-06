package model

type StorageMedia struct {
	DocumentID   string  `json:"__name__" firestore:"-"`        // generated uuid
	MediaID      *string `json:"media_id" firestore:"media_id"` // file ID returned by WhatsApp Business API after uploading media
	OriginalName string  `json:"original_name" firestore:"original_name"`
	MimeType     string  `json:"mime_type" firestore:"mime_type"`
	URL          string  `json:"url" firestore:"url"`
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
