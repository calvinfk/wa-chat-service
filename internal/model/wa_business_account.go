package model

import "time"

type WaBusinessAccount struct {
	DocumentID  string    `json:"id" firestore:"-"`
	TenantID    string    `json:"tenant_id" firestore:"-"` // foreign key to tenant
	WabaID      string    `json:"waba_id" firestore:"waba_id"`
	Name        string    `json:"name" firestore:"name"`
	AccessToken string    `json:"access_token" firestore:"access_token"` // encrypted
	CreatedAt   time.Time `json:"created_at" firestore:"created_at"`
}

func (WaBusinessAccount) TableName() string {
	return "wa_business_accounts"
}

func (WaBusinessAccount) AllowedFilterFields() []string {
	return []string{"id", "tenant_id", "waba_id"}
}

func (WaBusinessAccount) AllowedSortFields() []string {
	return []string{"created_at"}
}
