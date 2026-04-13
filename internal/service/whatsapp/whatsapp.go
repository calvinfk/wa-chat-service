package whatsapp_service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/utils"
)

type WhatsappBusiness struct {
}

func NewWhatsappService() *WhatsappBusiness {
	return &WhatsappBusiness{}
}

func (ws *WhatsappBusiness) SendMessage(client *whatsapp_business.Client, to string, payload whatsapp_business.MessageComponent) (whatsapp_business.MessageResponse, int, error) {
	response, httpCode, err := client.SendMessage(to, "individual", payload)
	if err != nil {
		if httpCode == http.StatusBadRequest {
			waError, ok := err.(whatsapp_business.WhatsAppBusinessError)
			if ok {
				payloadData, err := utils.AnyToJsonString(payload.GetPayload())
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

func (ws *WhatsappBusiness) ValidateTemplatePayload(client *whatsapp_business.Client, templateDB model.Template, templateSend whatsapp_business.MessageComponent) error {
	// assume language, name is already validated
	if templateDB.Status != "APPROVED" {
		return fmt.Errorf("template is not approved, current status: %s", templateDB.Status)
	}
	var err error
	if templateSend.GetType() != "template" {
		return fmt.Errorf("invalid type: expected 'template', got '%s'", templateSend.GetType())
	}
	sendPayload := templateSend.GetPayload()
	// parse template components
	var templateDBComponents []map[string]any
	err = json.Unmarshal([]byte(templateDB.Components), &templateDBComponents)
	if err != nil {
		return fmt.Errorf("failed to unmarshal template components: %v", err)
	}
	sendComponents, sendComponentsOk := sendPayload["components"].([]map[string]any)
	if templateDB.ParameterFormat == nil && sendComponentsOk {
		return fmt.Errorf("template does not have components but components found in the payload")
	} else if templateDB.ParameterFormat != nil && !sendComponentsOk {
		return fmt.Errorf("template has components but components missing in the payload")
	}
	if !sendComponentsOk {
		return nil // if template has no components and payload has no components, then it's valid
	}
	ws.validateTemplateParameter(templateDB.ParameterFormat, templateDBComponents, sendComponents)
	err = json.Unmarshal([]byte(templateDB.Components), &templateDBComponents)
	if err != nil {
		return fmt.Errorf("failed to unmarshal template components: %v", err)
	}
	if sendPayload["body"] == nil {
		return fmt.Errorf("body component is required")
	}
	return nil
}

func (ws *WhatsappBusiness) countParameters(parameterType string, components []map[string]any) map[string]int {
	return nil
}

func (ws *WhatsappBusiness) validateTemplateParameter(parameterType *string, dbComponents, sendComponents []map[string]any) error {
	// Check parameter format
	if parameterType == nil {
		return nil // if parameter format is not defined, skip parameter validation
	}
	data := make(map[string]int)
	checkComponentType := []string{"header", "body", "footer", "button"}
	for _, key := range checkComponentType {
		data[key] = 0
	}
	switch strings.ToUpper(*parameterType) {
	case "NAMED":
	case "POSITIONAL":
	}
	return nil
}

func (ws *WhatsappBusiness) validateTemplateHeader(client *whatsapp_business.Client, dbComponents map[string]any, sendPayload map[string]any) error {
	sendHeader, sendHeaderOk := sendPayload["header"].(map[string]any)
	dbHeader, dbHeaderOk := dbComponents["header"].(map[string]any)
	if !sendHeaderOk && dbHeaderOk {
		return fmt.Errorf("header component is required but missing in the payload")
	} else if sendHeaderOk && !dbHeaderOk {
		return fmt.Errorf("header component is not expected but found in the payload")
	} else if !sendHeaderOk && !dbHeaderOk {
		return nil
	}
	// check header type
	if sendHeader["type"] != dbHeader["format"] {
		return fmt.Errorf("header type mismatch: expected '%s', got '%s'", dbHeader["format"], sendHeader["type"])
	}
	// if header type is media, validate media ID
	headerType, ok := sendHeader["type"].(string)
	if !ok {
		return fmt.Errorf("header type is missing or not a string")
	}
	if whatsapp_business.IsMediaMessageType(headerType) {
		media, ok := sendHeader[headerType].(map[string]string)
		if !ok {
			return fmt.Errorf("media content is required for media header but missing or not a map")
		}
		mediaID, ok := media["id"]
		if !ok {
			return fmt.Errorf("media id is required for media header but missing or not a string")
		}
		httpCode, err := ws.validateMediaID(client, mediaID)
		if err != nil {
			return fmt.Errorf("failed to validate media ID '%s': %v (HTTP code: %d)", mediaID, err, httpCode)
		}
	} else if headerType == "text" {
		if sendHeader["text"] == nil {
			return fmt.Errorf("text content is required for text header but missing")
		}
	} else {
		return fmt.Errorf("unsupported header type: %s", headerType)
	}
	return nil
}

func (ws *WhatsappBusiness) validateMediaID(client *whatsapp_business.Client, mediaID string) (int, error) {
	_, httpCode, err := client.GetMediaURL(mediaID)
	if err != nil {
		if waErr, ok := err.(whatsapp_business.WhatsAppBusinessError); ok {
			log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][ValidateMediaID] WhatsApp Business API error: %s (code: %d, subcode: %d)", waErr.ErrorData.Message, waErr.ErrorData.Code, waErr.ErrorData.ErrorSubcode)
			return httpCode, waErr
		}
		log.Printf("[ERROR][internal/service/whatsapp/whatsapp.go][ValidateMediaID] Failed to validate media ID %s: %v", mediaID, err)
		return httpCode, err
	}
	return httpCode, nil
}
