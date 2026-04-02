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
	default:
		return nil, fmt.Errorf("unsupported message component type: %s", componentType)
	}
	if err := messageComponent.Validate(); err != nil {
		return nil, err
	}
	return messageComponent, nil
}
