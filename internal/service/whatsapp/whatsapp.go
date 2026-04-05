package whatsapp_service

import (
	"context"
	"log"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/meta/whatsapp_business"
	whatsapp_business_component "wa_chat_service/pkg/meta/whatsapp_business/component"
)

type WhatsappService struct {
}

func NewWhatsappService() *WhatsappService {
	return &WhatsappService{}
}

func (ws *WhatsappService) SendMessage(ctx context.Context, client *whatsapp_business.Client, to string, payload whatsapp_business_component.MessageComponent) (whatsapp_business.MessageResponse, error) {
	response, err := client.SendMessage(client.PhoneNumberID, to, payload)
	if err != nil {
		if waError, ok := err.(whatsapp_business.WhatsAppBusinessError); ok &&
			(waError.ErrorData.Code == whatsapp_business.PARAMETER_NOT_VALID ||
				waError.ErrorData.Code == whatsapp_business.REQUIRED_PARAMETER_MISSING ||
				waError.ErrorData.Code == whatsapp_business.INVALID_PARAMETER ||
				waError.ErrorData.Code == 132000) {
			payloadData, err := formatter.AnyToJsonString(payload.GetPayload())
			if err != nil {
				log.Println("[ERROR][internal/service/whatsapp/whatsapp.go][SendMessage] Failed to convert payload to JSON")
			} else {
				log.Println("[ERROR][internal/service/whatsapp/whatsapp.go][SendMessage] Parameter not valid, payload:", payloadData)
			}
		}
	}
	return response, err
}
func (ws *WhatsappService) GetTemplateList(ctx context.Context, client *whatsapp_business.Client) ([]any, error) {
	response, err := client.GetTemplateList()
	if err != nil {
		return nil, err
	}
	return response, nil
}
