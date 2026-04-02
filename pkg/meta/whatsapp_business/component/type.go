package whatsapp_business_component

import (
	"encoding/json"
	"fmt"
)

type MessageComponent interface {
	GetType() string
	GetPayload() map[string]any
	GetPayloadString() string
	Validate() error
}

func ValidateMapMessageComponent(componentType string, component any) (MessageComponent, error) {
	var err error
	var componentBytes []byte
	var messageComponent MessageComponent
	switch component := component.(type) {
	case []byte:
		componentBytes = component
	case string:
		componentBytes = []byte(component)
	case map[string]any, []any:
		componentBytes, err = json.Marshal(component)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported component payload type: %T", component)
	}
	switch componentType {
	case "text":
		var textComponent Text
		if err := json.Unmarshal(componentBytes, &textComponent); err != nil {
			return nil, err
		}
		messageComponent = textComponent
	case "audio":
		var audioComponent Audio
		if err := json.Unmarshal(componentBytes, &audioComponent); err != nil {
			return nil, err
		}
		messageComponent = audioComponent
	case "contacts":
		var contactsComponent Contacts
		if err := json.Unmarshal(componentBytes, &contactsComponent); err != nil {
			return nil, err
		}
		messageComponent = contactsComponent
	case "document":
		var documentComponent Document
		if err := json.Unmarshal(componentBytes, &documentComponent); err != nil {
			return nil, err
		}
		messageComponent = documentComponent
	case "image":
		var imageComponent Image
		if err := json.Unmarshal(componentBytes, &imageComponent); err != nil {
			return nil, err
		}
		messageComponent = imageComponent
	case "location":
		var locationComponent Location
		if err := json.Unmarshal(componentBytes, &locationComponent); err != nil {
			return nil, err
		}
		messageComponent = locationComponent
	case "reaction":
		var reactionComponent Reaction
		if err := json.Unmarshal(componentBytes, &reactionComponent); err != nil {
			return nil, err
		}
		messageComponent = reactionComponent
	case "sticker":
		var stickerComponent Sticker
		if err := json.Unmarshal(componentBytes, &stickerComponent); err != nil {
			return nil, err
		}
		messageComponent = stickerComponent
	case "video":
		var videoComponent Video
		if err := json.Unmarshal(componentBytes, &videoComponent); err != nil {
			return nil, err
		}
		messageComponent = videoComponent
	case "interactive":
		var interactiveComponent Interactive
		if err := json.Unmarshal(componentBytes, &interactiveComponent); err != nil {
			return nil, err
		}
		switch interactiveComponent.Type {
		case "cta_url":
			var interactiveCTAUrlComponent InteractiveCTAUrl
			if err := json.Unmarshal(componentBytes, &interactiveCTAUrlComponent); err != nil {
				return nil, err
			}
			messageComponent = interactiveCTAUrlComponent
		case "list":
			var interactiveListComponent InteractiveList
			if err := json.Unmarshal(componentBytes, &interactiveListComponent); err != nil {
				return nil, err
			}
			messageComponent = interactiveListComponent
		case "carousel":
			var interactiveCarouselComponent InteractiveCarousel
			if err := json.Unmarshal(componentBytes, &interactiveCarouselComponent); err != nil {
				return nil, err
			}
			messageComponent = interactiveCarouselComponent
		case "button":
			var interactiveButtonComponent InteractiveButton
			if err := json.Unmarshal(componentBytes, &interactiveButtonComponent); err != nil {
				return nil, err
			}
			messageComponent = interactiveButtonComponent
		default:
			return nil, fmt.Errorf("unsupported interactive message component type: %s", interactiveComponent.Type)
		}
	default:
		return nil, fmt.Errorf("unsupported message component type: %s", componentType)
	}
	if err := messageComponent.Validate(); err != nil {
		return nil, err
	}
	return messageComponent, nil
}

type Interactive struct {
	Type string `json:"type" validate:"required,oneof=cta_url list carousel button"`
}

type InteractiveBody struct {
	Text string `json:"text" validate:"required,max=1024"`
}

type InteractiveFooter struct {
	Text string `json:"text" validate:"required,max=60"`
}

// type InteractiveHeader struct {
// 	Type     string  `json:"type" validate:"required,oneof=text image document video"`
// 	Text     *string `json:"text" validate:"omitempty,max=60"`
// 	Image    *Media  `json:"image,omitempty"`
// 	Document *Media  `json:"document,omitempty"`
// 	Video    *Media  `json:"video,omitempty"`
// }

type Media struct {
	ID   *string `json:"id,omitempty" validate:"required_without=Link,excluded_with=Link,omitempty,min=1"` // Only if using uploaded media, Required if using uploaded media, otherwise omit.
	Link *string `json:"link,omitempty" validate:"required_without=ID,excluded_with=ID,omitempty,uri"`     // Only if using hosted media (not recommended), Required if using hosted media, otherwise omit.
}

type MediaAssetURL struct {
	Link string `json:"link,omitempty" validate:"uri"` // Only if using hosted media (not recommended), Required if using hosted media, otherwise omit.
}

type QuickReplyButton struct {
	Type       string                     `json:"type" validate:"required,eq=quick_reply"`
	QuickReply QuickReplyButtonQuickReply `json:"quick_reply" validate:"required"`
}

type QuickReplyButtonQuickReply struct {
	ID          string `json:"id" validate:"required,max=256"`
	DisplayText string `json:"display_text" validate:"required,max=20"`
}
