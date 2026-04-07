package whatsapp_service

import (
	"context"
	"log"
	"net/http"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/meta/whatsapp_business"
	whatsapp_business_component "wa_chat_service/pkg/meta/whatsapp_business/component"
)

type WhatsappBusiness struct {
}

func NewWhatsappService() *WhatsappBusiness {
	return &WhatsappBusiness{}
}

func (ws *WhatsappBusiness) SendMessage(ctx context.Context, client *whatsapp_business.Client, to string, payload whatsapp_business_component.MessageComponent) (whatsapp_business.MessageResponse, int, error) {
	response, httpCode, err := client.SendMessage(client.PhoneNumberID, to, "individual", payload)
	if err != nil {
		if httpCode == http.StatusBadRequest {
			waError, ok := err.(whatsapp_business.WhatsAppBusinessError)
			if ok {
				payloadData, err := formatter.AnyToJsonString(payload.GetPayload())
				log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][SendMessage] WhatsApp Business API error: %s (code: %d, subcode: %d)", waError.ErrorData.Message, waError.ErrorData.Code, waError.ErrorData.ErrorSubcode)
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
func (ws *WhatsappBusiness) GetTemplateList(ctx context.Context, client *whatsapp_business.Client) ([]any, int, error) {
	response, httpCode, err := client.GetTemplateList()
	if err != nil {
		return nil, httpCode, err
	}
	return response, httpCode, nil
}
func (ws *WhatsappBusiness) UploadMedia(ctx context.Context, client *whatsapp_business.Client, fileBytes []byte, filename, mimeType string) (string, int, error) {
	metaResponse, httpCode, err := client.UploadMedia(fileBytes, filename, mimeType)
	if err != nil {
		if waErr, ok := err.(whatsapp_business.WhatsAppBusinessError); ok {
			log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][UploadMedia] WhatsApp Business API error: %s (code: %d, subcode: %d)", waErr.ErrorData.Message, waErr.ErrorData.Code, waErr.ErrorData.ErrorSubcode)
			return "", httpCode, err
		}
		log.Println("[ERROR][internal/usecase/storage_media/storage_media.go][UploadMedia] WhatsApp Business API returned non-200 status code:", httpCode)
		return "", httpCode, err
	}
	return metaResponse.ID, httpCode, nil
}

func (ws *WhatsappBusiness) GetMediaURL(ctx context.Context, client *whatsapp_business.Client, mediaID string) (string, int, error) {
	mediaURLResponse, httpCode, err := client.GetMediaURL(mediaID)
	if err != nil {
		if waErr, ok := err.(whatsapp_business.WhatsAppBusinessError); ok {
			log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][GetMediaURL] WhatsApp Business API error: %s (code: %d, subcode: %d)", waErr.ErrorData.Message, waErr.ErrorData.Code, waErr.ErrorData.ErrorSubcode)
			return "", httpCode, err
		}
		log.Println("[ERROR][internal/service/whatsapp/whatsapp.go][GetMediaURL] Failed to get media URL: ", err)
		return "", httpCode, err
	}
	return mediaURLResponse.URL, httpCode, nil
}

func (ws *WhatsappBusiness) DownloadMedia(ctx context.Context, client *whatsapp_business.Client, mediaURL string) ([]byte, http.Header, int, error) {
	mediaData, urlHeaders, httpCode, err := client.DownloadMedia(mediaURL)
	if err != nil {
		log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][DownloadMedia] Failed to download media from URL %s: %v", mediaURL, err)
		return nil, nil, httpCode, err
	}
	return mediaData, urlHeaders, httpCode, nil
}

func (ws *WhatsappBusiness) DeleteMedia(ctx context.Context, client *whatsapp_business.Client, mediaID string) (int, error) {
	_, httpCode, err := client.DeleteMedia(mediaID)
	if err != nil {
		if waErr, ok := err.(whatsapp_business.WhatsAppBusinessError); ok {
			log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][DeleteMedia] WhatsApp Business API error: %s (code: %d, subcode: %d)", waErr.ErrorData.Message, waErr.ErrorData.Code, waErr.ErrorData.ErrorSubcode)
			return httpCode, err
		}
		log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][DeleteMedia] Failed to delete media with ID %s: %v", mediaID, err)
		return httpCode, err
	}
	return httpCode, nil
}
