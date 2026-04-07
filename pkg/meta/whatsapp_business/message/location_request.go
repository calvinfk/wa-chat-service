package message_components

import (
	"wa_chat_service/pkg/formatter"
)

type LocationRequest struct {
	Type   string                `json:"type" validate:"required,eq=location_request_message"`
	Body   InteractiveBody       `json:"body" validate:"required"`
	Action LocationRequestAction `json:"action" validate:"required"`
}

type LocationRequestAction struct {
	Name string `json:"name" validate:"required,eq=send_location"`
}

func (c LocationRequest) GetType() MessageType {
	return InteractiveMessageType
}

func (c LocationRequest) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c LocationRequest) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Interactive LocationRequest `json:"interactive" validate:"required"`
	}{
		Interactive: c,
	}
	return validator.Validate(data)
}

func (c LocationRequest) GetMessage() string {
	return c.Body.Text
}
