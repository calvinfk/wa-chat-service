package model

type StorageMedia struct {
	DocumentID   string `firestore:"-"`        // generated uuid
	MediaID      string `firestore:"media_id"` // file ID returned by WhatsApp Business API after uploading media
	OriginalName string `firestore:"original_name"`
	MimeType     string `firestore:"mime_type"`
	URL          string `firestore:"url"`
	AccessURL    string `firestore:"access_url"`
	CreatedAt    int64  `firestore:"created_at"`
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
