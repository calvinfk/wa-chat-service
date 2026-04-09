package whatsapp_business

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"

	"wa_chat_service/pkg/meta/whatsapp_business/message_components"
)

func NewComponent(componentType string, component any) (MessageComponent, error) {
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

func (wb *Client) SendMessage(phoneNumberID, to, recipientType string, payload MessageComponent) (MessageResponse, int, error) {
	payloadData := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    recipientType,
		"to":                to,
		"type":              payload.GetType(),
	}
	maps.Copy(payloadData, payload.GetPayload())
	endpoint := fmt.Sprintf("%s/%s/messages", wb.GetBaseURLVersion(), phoneNumberID)
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
