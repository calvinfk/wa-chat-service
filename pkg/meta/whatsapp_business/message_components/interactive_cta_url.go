package message_components

import (
	"wa_chat_service/pkg/utils"
)

type InteractiveCTAUrl struct {
	Type   string                   `json:"type" validate:"required,eq=cta_url"`
	Header *InteractiveCTAUrlHeader `json:"header,omitempty"`
	Body   InteractiveBody          `json:"body" validate:"required"`
	Action InteractiveCTAUrlAction  `json:"action"`
	Footer *InteractiveFooter       `json:"footer,omitempty"`
}

type InteractiveCTAUrlHeader struct {
	Type     string         `json:"type" validate:"required,oneof=text image document video"`
	Text     *string        `json:"text" validate:"omitempty,max=60"`
	Image    *MediaAssetURL `json:"image,omitempty"`
	Document *MediaAssetURL `json:"document,omitempty"`
	Video    *MediaAssetURL `json:"video,omitempty"`
}

type InteractiveCTAUrlAction struct {
	Name       string                     `json:"name" validate:"required,eq=cta_url"`
	Parameters InteractiveCTAUrlParameter `json:"parameters"`
}

type InteractiveCTAUrlParameter struct {
	DisplayText string `json:"display_text" validate:"required,max=20"`
	URL         string `json:"url" validate:"required,uri"`
}

func (c InteractiveCTAUrl) GetType() MessageType {
	return InteractiveMessageType
}

func (c InteractiveCTAUrl) GetPayload() map[string]any {
	jsonData, err := utils.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c InteractiveCTAUrl) Validate() error {
	validator := utils.NewValidator()
	data := struct {
		Interactive InteractiveCTAUrl `json:"interactive" validate:"required"`
	}{
		Interactive: c,
	}
	return validator.Struct(data)
}

func (c InteractiveCTAUrl) GetMessage() string {
	return c.Body.Text
}
