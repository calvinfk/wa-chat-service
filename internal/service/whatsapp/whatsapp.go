package whatsapp_service

import (
	"context"
	"log"
	"wa_chat_service/pkg/meta/whatsapp_business"
	whatsapp_business_component "wa_chat_service/pkg/meta/whatsapp_business/component"
)

type WhatsappService struct {
	whatsappClient *whatsapp_business.Client
}

func NewWhatsappService(whatsappClient *whatsapp_business.Client) *WhatsappService {
	return &WhatsappService{
		whatsappClient: whatsappClient,
	}
}

func (ws *WhatsappService) SendMessage(ctx context.Context, phoneNumberID, to string, payload whatsapp_business_component.MessageComponent) (whatsapp_business.MessageResponse, error) {
	response, err := ws.whatsappClient.SendMessage(phoneNumberID, to, payload)
	if err != nil {
		if waError, ok := err.(whatsapp_business.WhatsAppBusinessError); ok && (waError.ErrorData.Code == whatsapp_business.PARAMETER_NOT_VALID || waError.ErrorData.Code == whatsapp_business.REQUIRED_PARAMETER_MISSING) {
			log.Println("[ERROR][internal/service/whatsapp/whatsapp.go][SendMessage] Parameter not valid, payload:", payload.GetPayloadString())
		}
	}
	return response, err
}
