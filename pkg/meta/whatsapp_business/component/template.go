package whatsapp_business_component

import "wa_chat_service/pkg/formatter"

type Template struct {
	Name       string               `json:"name" validate:"required"`
	Language   LanguageCode         `json:"language" validate:"required"`
	Components *[]TemplateComponent `json:"components,omitempty" validate:"omitempty,dive"`
}

type LanguageCode struct {
	Code string `json:"code" validate:"required"`
}

type TemplateComponent interface {
	GetType() MessageType
	GetMessage() string
}
type TemplateComponentBase struct {
	Type string `json:"type" validate:"required"` // header, body, button, etc
}

func (t TemplateComponentBase) GetType() MessageType {
	return TemplateMessageType
}

func (t TemplateComponentBase) GetMessage() string {
	return "(Template)"
}

type TemplateResponse struct {
	Category                    string `json:"category" validate:"required"`   // MARKETING, UTILITY, etc
	Components                  []any  `json:"components" validate:"required"` // header, body, button, etc
	ID                          string `json:"id" validate:"required"`
	IsPrimaryDeviceDeliveryOnly bool   `json:"is_primary_device_delivery_only"`
	Language                    string `json:"language" validate:"required"`
	MessageSendTTLSeconds       int    `json:"message_send_ttl_seconds"`
	Name                        string `json:"name" validate:"required"`
	ParameterFormat             string `json:"parameter_format" validate:"required"`
	Status                      string `json:"status" validate:"required"` // approved, rejected, etc
}

func (t *Template) GetType() MessageType {
	return TemplateMessageType
}
func (c Template) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Template) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Template Template `json:"template" validate:"required"`
	}{
		Template: c,
	}
	return validator.Validate(data)
}

func (c Template) GetMessage() string {
	return "(Template + " + c.Name + ")"
}
