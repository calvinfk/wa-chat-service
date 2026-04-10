package message_components

import "wa_chat_service/pkg/utils"

type Template struct {
	Name       string            `json:"name" validate:"required"`
	Language   LanguageCode      `json:"language" validate:"required"`
	Components *[]map[string]any `json:"components,omitempty" validate:"omitempty,dive"`
}

type LanguageCode struct {
	Code string `json:"code" validate:"required"`
}

// type TemplateComponent interface {
// 	GetType() MessageType
// 	GetMessage() string
// }

// type TemplateComponentBase struct {
// 	Type string `json:"type" validate:"required"` // header, body, button, etc
// }

// func (t TemplateComponentBase) GetType() MessageType {
// 	return TemplateMessageType
// }

// func (t TemplateComponentBase) GetMessage() string {
// 	return "(Template)"
// }

func (c Template) GetType() MessageType {
	return TemplateMessageType
}

func (c Template) GetPayload() map[string]any {
	jsonData, err := utils.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Template) Validate() error {
	validator := utils.NewValidator()
	data := struct {
		Template Template `json:"template" validate:"required"`
	}{
		Template: c,
	}
	return validator.Struct(data)
}

func (c Template) GetMessage() string {
	return "(Template + " + c.Name + ")"
}
