package model

type StorageMedia struct {
	DocumentID   string `firestore:"-"` // uploaded id
	OriginalName string `firestore:"original_name"`
	URL          string `firestore:"url"`
	AccessURL    string `firestore:"access_url"`
	CreatedAt    int64  `firestore:"created_at"`
}

func (m StorageMedia) TableName() string {
	return "storage_medias"
}

func (m StorageMedia) AllowedFilterFields() []string {
	return []string{"document_id", "original_name", "url", "access_url", "created_at"}
}
func (m StorageMedia) AllowedSortFields() []string {
	return []string{"created_at"}
}
