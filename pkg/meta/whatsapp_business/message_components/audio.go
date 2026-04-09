package message_components

import "wa_chat_service/pkg/utils"

type Audio struct {
	Media
	Voice *bool `json:"voice,omitempty"` // Only include if sending voice message
}

func (c Audio) GetType() MessageType {
	return AudioMessageType
}

func (c Audio) GetPayload() map[string]any {
	jsonData, err := utils.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Audio) Validate() error {
	validator := utils.NewValidator()
	data := struct {
		Audio Audio `json:"audio" validate:"required"`
	}{
		Audio: c,
	}
	return validator.Struct(data)
}

func (c Audio) GetMessage() string {
	return "(Audio)"
}
