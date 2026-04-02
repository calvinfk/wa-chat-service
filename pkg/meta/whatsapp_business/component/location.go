package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/validate_struct"
)

type Location struct {
	Latitude  string  `json:"latitude" validate:"required,numeric"`
	Longitude string  `json:"longitude" validate:"required,numeric"`
	Name      *string `json:"name,omitempty"`
	Address   *string `json:"address,omitempty"`
}

func (c Location) GetType() string {
	return "location"
}

func (c Location) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		c.GetType(): jsonData,
	}
}

func (c Location) GetPayloadString() string {
	jsonData := c.GetPayload()[c.GetType()]
	jsonString, err := formatter.AnyToJsonString(jsonData)
	if err != nil {
		panic(err)
	}
	return jsonString
}

func (c Location) Validate() error {
	validator := validate_struct.New()
	data := struct {
		Location Location `json:"location" validate:"required"`
	}{
		Location: c,
	}
	return validator.Validate(data)
}
