package message_components

import (
	"wa_chat_service/pkg/utils"
)

type InteractiveButton struct {
	Type   string                   `json:"type" validate:"required,eq=button"`
	Header *InteractiveButtonHeader `json:"header,omitempty"`
	Body   InteractiveBody          `json:"body" validate:"required"`
	Footer *InteractiveFooter       `json:"footer,omitempty"`
	Action InteractiveButtonAction  `json:"action" validate:"required"`
}

type InteractiveButtonHeader struct {
	Type     string  `json:"type" validate:"required,oneof=text image document video"`
	Text     *string `json:"text" validate:"omitempty,max=60"`
	Image    *Media  `json:"image,omitempty"`
	Document *Media  `json:"document,omitempty"`
	Video    *Media  `json:"video,omitempty"`
}

type InteractiveButtonAction struct {
	Buttons []InteractiveButtonActionButton `json:"buttons" validate:"min=1,max=3,dive"`
}

type InteractiveButtonActionButton struct {
	Type  string                             `json:"type" validate:"required,eq=reply"`
	Reply InteractiveButtonActionButtonReply `json:"reply" validate:"required"`
}
type InteractiveButtonActionButtonReply struct {
	ID    string `json:"id" validate:"required,max=256"`
	Title string `json:"title" validate:"required,max=20"`
}

func (c InteractiveButton) GetType() MessageType {
	return InteractiveMessageType
}

func (c InteractiveButton) GetPayload() map[string]any {
	jsonData, err := utils.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c InteractiveButton) Validate() error {
	validator := utils.NewValidator()
	data := struct {
		Interactive InteractiveButton `json:"interactive" validate:"required"`
	}{
		Interactive: c,
	}
	return validator.Struct(data)
}

func (c InteractiveButton) GetMessage() string {
	return c.Body.Text
}
