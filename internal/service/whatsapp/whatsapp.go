package whatsapp_service

import (
	"log"
	"net/http"
	"wa_chat_service/internal/dto"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/meta/whatsapp_business"
)

type WhatsappBusiness struct {
}

func NewWhatsappService() *WhatsappBusiness {
	return &WhatsappBusiness{}
}

func (ws *WhatsappBusiness) SendMessage(client *whatsapp_business.Client, to string, payload whatsapp_business.MessageComponent) (whatsapp_business.MessageResponse, int, error) {
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

func (ws *WhatsappBusiness) GetTemplateList(client *whatsapp_business.Client) ([]whatsapp_business.TemplateResponse, int, error) {
	response, httpCode, err := client.GetTemplateList()
	if err != nil {
		if waErr, ok := err.(whatsapp_business.WhatsAppBusinessError); ok {
			log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][GetTemplateList] WhatsApp Business API error: %s (code: %d, subcode: %d)", waErr.ErrorData.Message, waErr.ErrorData.Code, waErr.ErrorData.ErrorSubcode)
		} else {
			log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][GetTemplateList] Failed to get template list: %v", err)
		}
		return nil, httpCode, err
	}
	return response, httpCode, nil
}

func (ws *WhatsappBusiness) UploadMedia(client *whatsapp_business.Client, fileBytes []byte, filename, mimeType string) (string, int, error) {
	metaResponse, httpCode, err := client.UploadMedia(fileBytes, filename, mimeType)
	if err != nil {
		if waErr, ok := err.(whatsapp_business.WhatsAppBusinessError); ok {
			log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][UploadMedia] WhatsApp Business API error: %s (code: %d, subcode: %d)", waErr.ErrorData.Message, waErr.ErrorData.Code, waErr.ErrorData.ErrorSubcode)
			return "", httpCode, err
		}
		log.Println("[ERROR][internal/service/whatsapp/whatsapp.go][UploadMedia] WhatsApp Business API returned non-200 status code:", httpCode)
		return "", httpCode, err
	}
	return metaResponse.ID, httpCode, nil
}

func (ws *WhatsappBusiness) GetMediaURL(client *whatsapp_business.Client, mediaID string) (string, int, error) {
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

func (ws *WhatsappBusiness) DownloadMedia(client *whatsapp_business.Client, mediaURL string) ([]byte, http.Header, int, error) {
	mediaData, urlHeaders, httpCode, err := client.DownloadMedia(mediaURL)
	if err != nil {
		log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][DownloadMedia] Failed to download media from URL %s: %v", mediaURL, err)
		return nil, nil, httpCode, err
	}
	return mediaData, urlHeaders, httpCode, nil
}

func (ws *WhatsappBusiness) DeleteMedia(client *whatsapp_business.Client, mediaID string) (int, error) {
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

func (ws *WhatsappBusiness) CreateTemplate(client *whatsapp_business.Client, inputData dto.TemplateCreateRequest) (whatsapp_business.TemplateCreateResponse, int, error) {
	template := whatsapp_business.TemplateCreateRequest{
		Name:            inputData.Name,
		Category:        inputData.Category,
		Language:        inputData.Language,
		ParameterFormat: inputData.ParameterFormat,
		Components:      inputData.Components,
	}
	response, httpCode, err := client.CreateTemplate(template)
	if err != nil {
		if waErr, ok := err.(whatsapp_business.WhatsAppBusinessError); ok {
			log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][CreateTemplate] WhatsApp Business API error: %s (code: %d, subcode: %d)", waErr.ErrorData.Message, waErr.ErrorData.Code, waErr.ErrorData.ErrorSubcode)
			return whatsapp_business.TemplateCreateResponse{}, httpCode, waErr
		}
		log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][CreateTemplate] Failed to create template: %v", err)
		return whatsapp_business.TemplateCreateResponse{}, httpCode, err
	}
	return response, httpCode, nil
}
