package message_components

import (
	"wa_chat_service/pkg/formatter"
)

type Location struct {
	Latitude  string  `json:"latitude" validate:"required,numeric"`
	Longitude string  `json:"longitude" validate:"required,numeric"`
	Name      *string `json:"name,omitempty"`
	Address   *string `json:"address,omitempty"`
}

func (c Location) GetType() MessageType {
	return LocationMessageType
}

func (c Location) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c Location) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Location Location `json:"location" validate:"required"`
	}{
		Location: c,
	}
	return validator.Validate(data)
}

func (c Location) GetMessage() string {
	if c.Name != nil {
		return *c.Name
	} else if c.Address != nil {
		return *c.Address
	}
	return "(Location " + c.Latitude + ", " + c.Longitude + ")"
}
