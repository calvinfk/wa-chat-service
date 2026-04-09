package message_components

import (
	"wa_chat_service/pkg/utils"
)

type Video struct {
	Media
	Caption *string `json:"caption,omitempty" validate:"omitempty,max=1024"`
}

func (c Video) GetType() MessageType {
	return VideoMessageType
}

func (c Video) GetPayload() map[string]any {
	jsonData, err := utils.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Video) Validate() error {
	validator := utils.NewValidator()
	data := struct {
		Video Video `json:"video" validate:"required"`
	}{
		Video: c,
	}
	return validator.Struct(data)
}
func (c Video) GetMessage() string {
	if c.Caption != nil {
		return *c.Caption
	}
	return "(Video)"
}
