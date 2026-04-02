package model

type StorageMedia struct {
	DocumentID   string `firestore:"-"` // uploaded id
	OriginalName string `firestore:"original_name"`
	URL          string `firestore:"url"`
	AccessURL    string `firestore:"access_url"`
}
