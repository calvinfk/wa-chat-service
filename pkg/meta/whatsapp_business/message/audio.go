package message_components

import (
	"wa_chat_service/pkg/formatter"
)

type Audio struct {
	Media
	Voice *bool `json:"voice,omitempty"` // Only include if sending voice message
}

func (c Audio) GetType() MessageType {
	return AudioMessageType
}

func (c Audio) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Audio) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Audio Audio `json:"audio" validate:"required"`
	}{
		Audio: c,
	}
	return validator.Validate(data)
}

func (c Audio) GetMessage() string {
	return "(Audio)"
}
