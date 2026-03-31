package whatsapp_service

import (
	"context"
	"wa_chat_service/pkg/meta/whatsapp_business"
)

type WhatsappService struct {
	whatsappClient *whatsapp_business.Client
}

func NewWhatsappService(whatsappClient *whatsapp_business.Client) *WhatsappService {
	return &WhatsappService{
		whatsappClient: whatsappClient,
	}
}

func (ws *WhatsappService) SendMessage(ctx context.Context, phoneNumberID, to string, payload whatsapp_business.MessageComponent) (whatsapp_business.MessageResponse, error) {
	return ws.whatsappClient.SendMessage(phoneNumberID, to, payload)
}
