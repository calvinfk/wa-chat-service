package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
)

type InteractiveList struct {
	Type   string                `json:"type" validate:"required,eq=list"`
	Body   InteractiveBody       `json:"body" validate:"required"`
	Action InteractiveListAction `json:"action" validate:"required"`
}

type InteractiveListAction struct {
	Button   string                   `json:"button" validate:"required,max=20"`
	Sections []InteractiveListSection `json:"sections" validate:"min=1,max=10,dive"`
}
type InteractiveListSection struct {
	Title string               `json:"title" validate:"required,max=24"`
	Rows  []InteractiveListRow `json:"rows" validate:"min=1,max=10,dive"`
}

type InteractiveListRow struct {
	ID          string `json:"id" validate:"required,max=200"`
	Title       string `json:"title" validate:"required,max=24"`
	Description string `json:"description,omitempty" validate:"omitempty,max=72"`
}

func (c InteractiveList) GetType() MessageType {
	return InteractiveMessageType
}

func (c InteractiveList) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c InteractiveList) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Interactive InteractiveList `json:"interactive" validate:"required"`
	}{
		Interactive: c,
	}
	return validator.Validate(data)
}

func (c InteractiveList) GetMessage() string {
	return c.Body.Text
}
