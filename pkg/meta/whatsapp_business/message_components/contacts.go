package message_components

import (
	"wa_chat_service/pkg/utils"
)

type Contacts []Contact

type Contact struct {
	Addresses []Address `json:"addresses,omitempty" validate:"dive"`
	Birthday  *string   `json:"birthday,omitempty"` // Format: YYYY-MM-DD
	Emails    []Email   `json:"emails,omitempty" validate:"dive"`
	Name      Name      `json:"name" validate:"required"`
	Org       *Org      `json:"org,omitempty"`
	Phones    []Phone   `json:"phones,omitempty" validate:"dive"`
	Urls      []Url     `json:"urls,omitempty" validate:"dive"`
}

type Address struct {
	Street      *string `json:"street,omitempty"`
	City        *string `json:"city,omitempty"`
	State       *string `json:"state,omitempty"`
	Zip         *string `json:"zip,omitempty"`
	Country     *string `json:"country,omitempty"`
	CountryCode *string `json:"country_code,omitempty"`
	Type        *string `json:"type,omitempty"` // e.g., "home", "work"
}

type Email struct {
	Email *string `json:"email,omitempty"` // Email address
	Type  *string `json:"type,omitempty"`  // e.g., "home", "work"
}

type Name struct {
	FormattedName string  `json:"formatted_name" validate:"required"`
	FirstName     string  `json:"first_name" validate:"required"`
	LastName      *string `json:"last_name,omitempty"`
	MiddleName    *string `json:"middle_name,omitempty"`
	Suffix        *string `json:"suffix,omitempty"`
	Prefix        *string `json:"prefix,omitempty"`
}

type Org struct {
	Company    *string `json:"company,omitempty"`
	Department *string `json:"department,omitempty"`
	Title      *string `json:"title,omitempty"`
}

type Phone struct {
	Phone *string `json:"phone,omitempty"` // Whatsapp user phone number in international format
	Type  *string `json:"type,omitempty"`  // e.g., "home", "work", "mobile"
	WaID  *string `json:"wa_id,omitempty"` // If omitted, the message will display an Invite to WhatsApp button instead of the standard buttons
}

type Url struct {
	Url  *string `json:"url"`            // URL string
	Type *string `json:"type,omitempty"` // e.g., "home", "company"
}

func (c Contacts) GetType() MessageType {
	return ContactsMessageType
}

func (c Contacts) GetPayload() map[string]any {
	var contactPayloads []map[string]any
	for _, contact := range c {
		// Convert each contact to a map[string]any if needed
		contactPayload, err := utils.StructToMap(contact, true)
		if err != nil {
			panic(err)
		}
		contactPayloads = append(contactPayloads, contactPayload)
	}
	return map[string]any{
		string(c.GetType()): contactPayloads,
	}
}

func (c Contacts) Validate() error {
	validator := utils.NewValidator()
	validatePayload := struct {
		Contacts Contacts `json:"contacts" validate:"dive"`
	}{
		Contacts: c,
	}
	return validator.Struct(validatePayload)
}

func (c Contacts) GetMessage() string {
	return "(Contacts)"
}
