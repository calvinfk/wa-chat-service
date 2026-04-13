package whatsapp_business

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"wa_chat_service/pkg/meta/whatsapp_business/message_components"
)

func NewComponent(componentType string, component any) (MessageComponent, error) {
	if component == nil {
		return nil, fmt.Errorf("component is nil")
	}
	if componentType == "" {
		return nil, fmt.Errorf("component type is empty")
	}
	var messageStruct MessageComponent
	var ok bool
	messageType := message_components.MessageType(componentType)
	if messageType == message_components.InteractiveMessageType {
		interactiveType := component.(map[string]any)["type"].(string)
		messageStruct, ok = interactiveMessageRegistry[interactiveType]
	} else {
		messageStruct, ok = messageRegistry[message_components.MessageType(messageType)]
	}
	if !ok || messageStruct == nil {
		return nil, fmt.Errorf("unsupported message type: %s", messageType)
	}
	messageBytes, err := json.Marshal(component)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}
	if err := json.Unmarshal(messageBytes, &messageStruct); err != nil {
		return nil, err
	}
	if messageStruct == nil {
		return nil, fmt.Errorf("message struct is nil after unmarshalling")
	}
	if err := messageStruct.Validate(); err != nil {
		return nil, err
	}
	return messageStruct, nil
}

func GetMedia(component MessageComponent) *message_components.Media {
	switch component.GetType() {
	case message_components.AudioMessageType:
		return &component.(*message_components.Audio).Media
	case message_components.DocumentMessageType:
		return &component.(*message_components.Document).Media
	case message_components.ImageMessageType:
		return &component.(*message_components.Image).Media
	case message_components.StickerMessageType:
		return &component.(*message_components.Sticker).Media
	case message_components.VideoMessageType:
		return &component.(*message_components.Video).Media
	default:
		return nil
	}
}

func NewTemplateComponent(payloadBytes []byte) (message_components.Template, error) {
	var payloadMap map[string]any
	if err := json.Unmarshal(payloadBytes, &payloadMap); err != nil {
		return message_components.Template{}, fmt.Errorf("failed to unmarshal template payload: %v", err)
	}
	if payloadMap["template"] == nil {
		return message_components.Template{}, fmt.Errorf("invalid template payload: missing 'template' field")
	}
	templateBytes, err := json.Marshal(payloadMap["template"])
	if err != nil {
		return message_components.Template{}, fmt.Errorf("failed to marshal template component: %v", err)
	}
	var templateStruct message_components.Template
	if err := json.Unmarshal(templateBytes, &templateStruct); err != nil {
		return templateStruct, err
	}
	if err := templateStruct.Validate(); err != nil {
		return templateStruct, err
	}
	return templateStruct, nil

}

func IsMediaMessageType(messageTypeStr string) bool {
	messageTypeStr = strings.ToLower(messageTypeStr)
	messageType := message_components.MessageType(messageTypeStr)
	switch messageType {
	case message_components.AudioMessageType,
		message_components.DocumentMessageType,
		message_components.ImageMessageType,
		message_components.StickerMessageType,
		message_components.VideoMessageType:
		return true
	default:
		return false
	}
}

func IsMediaAllowed(messageTypeStr, mimeTypeStr string) bool {
	messageType := message_components.MessageType(strings.ToLower(messageTypeStr))
	allowedTypes, exists := allowedMediaTypes[messageType]
	if !exists {
		return false
	}
	for _, allowedMimeType := range allowedTypes {
		if allowedMimeType == mimeTypeStr {
			return true
		}
	}
	return false
}

func (wb *Client) SendMessage(to, recipientType string, payload MessageComponent) (MessageResponse, int, error) {
	payloadData := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    recipientType,
		"to":                to,
		"type":              payload.GetType(),
	}
	maps.Copy(payloadData, payload.GetPayload())
	endpoint := fmt.Sprintf("%s/%s/messages", wb.GetBaseURLVersion(), wb.PhoneNumberID)
	body, httpCode, err := wb.accessAPI(http.MethodPost, endpoint, payloadData)
	if err != nil {
		return MessageResponse{}, httpCode, err
	} else if httpCode != http.StatusOK {
		return parseMetaErrorResponse(MessageResponse{}, body, httpCode)
	}
	var response MessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return MessageResponse{}, httpCode, err
	}
	return response, httpCode, nil
}
