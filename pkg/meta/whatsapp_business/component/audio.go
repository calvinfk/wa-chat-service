package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
)

type Audio struct {
	ID    *string `json:"id"`              // Only if using uploaded media
	Link  *string `json:"link,omitempty"`  // Only if using hosted media (not recommended)
	Voice *bool   `json:"voice,omitempty"` // Only include if sending voice message
}

func (a Audio) GetType() string {
	return "audio"
}

func (a Audio) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(a, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		a.GetType(): jsonData,
	}
}

func (a Audio) GetPayloadString() string {
	jsonData := a.GetPayload()[a.GetType()]
	jsonString, err := formatter.AnyToJsonString(jsonData)
	if err != nil {
		panic(err)
	}
	return jsonString
}

func (a Audio) Validate() error {
	// No required fields, but you can add custom validation if needed
	return nil
}
