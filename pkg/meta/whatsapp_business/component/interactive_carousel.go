package whatsapp_business_component

import (
	"wa_chat_service/pkg/formatter"
)

type InteractiveCarousel struct {
	Type   string                    `json:"type" validate:"required,eq=carousel"`
	Body   InteractiveBody           `json:"body" validate:"required"`
	Action InteractiveCarouselAction `json:"action" validate:"required"`
}

type InteractiveCarouselAction struct {
	Cards []InteractiveCarouselCard `json:"cards" validate:"min=1,max=10,dive"`
}

type InteractiveCarouselCard struct {
	CardIndex int                           `json:"card_index" validate:"min=0"` // TODO: validate if the index is sequential and starts from 0
	Type      string                        `json:"type" validate:"required,eq=cta_url"`
	Header    InteractiveCarouselCardHeader `json:"header" validate:"required"`
	Body      *InteractiveCarouselCardBody  `json:"body"`
	Action    InteractiveCarouselCardAction `json:"action" validate:"required"`
}

type InteractiveCarouselCardHeader struct {
	Type  string         `json:"type" validate:"required,oneof=image video"`
	Image *MediaAssetURL `json:"image,omitempty"`
	Video *MediaAssetURL `json:"video,omitempty"`
}

type InteractiveCarouselCardBody struct {
	Text string `json:"text" validate:"required,max=1024"`
}

type InteractiveCarouselCardAction struct {
	InteractiveCTAUrlAction
	Buttons *[]QuickReplyButton `json:"buttons" validate:"omitempty,dive"`
}

func (c InteractiveCarousel) GetType() MessageType {
	return InteractiveMessageType
}

func (c InteractiveCarousel) GetPayload() map[string]any {
	jsonData, err := formatter.StructToMap(c, true)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		string(c.GetType()): jsonData,
	}
}

func (c InteractiveCarousel) Validate() error {
	validator := formatter.Validator()
	data := struct {
		Interactive InteractiveCarousel `json:"interactive" validate:"required"`
	}{
		Interactive: c,
	}
	err := validator.Validate(data)
	if err != nil {
		return err
	}
	return nil
}

func (c InteractiveCarousel) GetMessage() string {
	return c.Body.Text
}
