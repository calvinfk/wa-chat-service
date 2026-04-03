package model

import "strings"

type Chat struct {
	DocumentID  string `firestore:"-"`         // {recipient_id}-{phone_number_id}
	ChatType    string `firestore:"chat_type"` // individual or group
	DisplayName string `firestore:"display_name"`
	LastMessage string `firestore:"last_message"`
	CreatedAt   int64  `firestore:"created_at"`
	UpdatedAt   int64  `firestore:"updated_at"`
}

func (c *Chat) GetRecipientID() string {
	i := strings.LastIndex(c.DocumentID, "-")
	if i != -1 {
		return c.DocumentID[:i]
	}
	return ""
}

func (c *Chat) GetPhoneNumberID() string {
	i := strings.LastIndex(c.DocumentID, "-")
	if i != -1 {
		return c.DocumentID[i+1:]
	}
	return ""
}
