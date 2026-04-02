package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/validate_struct"
)

type LocationRequest struct {
	Type   string                `json:"type" validate:"required,eq=location_request_message"`
	Body   InteractiveBody       `json:"body" validate:"required"`
	Action LocationRequestAction `json:"action" validate:"required"`
}

type LocationRequestAction struct {
	Name string `json:"name" validate:"required,eq=send_location"`
}

func (c LocationRequest) GetType() string {
	return "interactive"
}

func (c LocationRequest) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		c.GetType(): jsonData,
	}
}

func (c LocationRequest) GetPayloadString() string {
	jsonData := c.GetPayload()[c.GetType()]
	jsonString, err := formatter.AnyToJsonString(jsonData)
	if err != nil {
		panic(err)
	}
	return jsonString
}

func (c LocationRequest) Validate() error {
	validator := validate_struct.New()
	data := struct {
		Interactive LocationRequest `json:"interactive" validate:"required"`
	}{
		Interactive: c,
	}
	return validator.Validate(data)
}
