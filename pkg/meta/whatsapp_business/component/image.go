package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
)

type Image struct {
	Media
	Caption *string `json:"caption,omitempty" validate:"omitempty,max=1024"`
}

func (c Image) GetType() MessageType {
	return ImageMessageType
}

func (c Image) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Image) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Image Image `json:"image" validate:"required"`
	}{
		Image: c,
	}
	return validator.Validate(data)
}

func (c Image) GetMessage() string {
	if c.Caption != nil {
		return *c.Caption
	}
	return "(Image)"
}
