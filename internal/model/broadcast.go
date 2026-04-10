package model

import "time"

type (
	Broadcast struct {
		DocumentID    string    `json:"__name__" firestore:"-"` // uuid v7
		Name          string    `json:"name" firestore:"name"`
		TemplateID    string    `json:"template_id" firestore:"template_id"`
		PhoneNumberID string    `json:"phone_number_id" firestore:"phone_number_id"`
		Payload       string    `json:"payload" firestore:"payload"` // raw json string of template
		Status        string    `json:"status" firestore:"status"`   // scheduled, sent, cancelled
		SendAt        time.Time `json:"send_at" firestore:"send_at"`
		CreatedAt     time.Time `json:"created_at" firestore:"created_at"`
		UpdatedAt     time.Time `json:"updated_at" firestore:"updated_at"`
	}
)

func (b Broadcast) TableName() string {
	return "broadcasts"
}
