package whatsapp_service

import (
	"context"
	"log"
	"net/http"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/meta/whatsapp_business"
	whatsapp_business_component "wa_chat_service/pkg/meta/whatsapp_business/component"
)

type WhatsappService struct {
}

func NewWhatsappService() *WhatsappService {
	return &WhatsappService{}
}

func (ws *WhatsappService) SendMessage(ctx context.Context, client *whatsapp_business.Client, to string, payload whatsapp_business_component.MessageComponent) (whatsapp_business.MessageResponse, int, error) {
	response, httpCode, err := client.SendMessage(client.PhoneNumberID, "individual", to, payload)
	if err != nil {
		if httpCode == http.StatusBadRequest {
			waError, ok := err.(whatsapp_business.WhatsAppBusinessError)
			if ok {
				log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][SendMessage] WhatsApp Business API error: %s (code: %d, subcode: %d)", waError.ErrorData.Message, waError.ErrorData.Code, waError.ErrorData.ErrorSubcode)
				payloadData, err := formatter.AnyToJsonString(payload.GetPayload())
				if err != nil {
					log.Println("[ERROR][internal/service/whatsapp/whatsapp.go][SendMessage] Failed to convert payload to JSON")
				} else {
					log.Println("[ERROR][internal/service/whatsapp/whatsapp.go][SendMessage] Parameter not valid, payload:", payloadData)
				}
			} else {
				log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][SendMessage] Failed to send message: %v", err)
			}
		}
		return whatsapp_business.MessageResponse{}, httpCode, err
	}
	return response, httpCode, err
}
func (ws *WhatsappService) GetTemplateList(ctx context.Context, client *whatsapp_business.Client) ([]any, int, error) {
	response, httpCode, err := client.GetTemplateList()
	if err != nil {
		return nil, httpCode, err
	}
	return response, httpCode, nil
}
