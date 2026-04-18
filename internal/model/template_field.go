package model

type TemplateField struct {
	DocumentID string `json:"id" firestore:"-"`        // uuid v7
	TenantID   string `json:"tenant_id" firestore:"-"` // reference to tenant id from tenant collection
	Name       string `json:"name" firestore:"name"`
	Key        string `json:"key" firestore:"key"`     // e.g. {{contact-name}} => key: contact-name (used in template content)
	Field      string `json:"field" firestore:"field"` // e.g. contact.name, contact.phone_number (used to fetch value from contact collection)
}

func (m TemplateField) TableName() string {
	return "template_fields"
}
