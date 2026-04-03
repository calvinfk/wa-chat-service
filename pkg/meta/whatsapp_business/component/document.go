package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
)

type Document struct {
	Media
	Caption  *string `json:"caption,omitempty" validate:"omitempty,max=1024"`
	Filename *string `json:"filename,omitempty"` // Document filename, with extension.
}

func (c Document) GetType() MessageType {
	return DocumentMessageType
}

func (c Document) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Document) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Document Document `json:"document" validate:"required"`
	}{
		Document: c,
	}
	return validator.Validate(data)
}

func (c Document) GetMessage() string {
	if c.Caption != nil {
		return *c.Caption
	} else if c.Filename != nil {
		return *c.Filename
	}
	return "(Document)"
}
