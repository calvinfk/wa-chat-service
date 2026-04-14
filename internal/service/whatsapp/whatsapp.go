package whatsapp_service

import (
	"fmt"
	"log"
	"net/http"
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

func (ws *WhatsappBusiness) ValidateTemplatePayload(client *whatsapp_business.Client, templateDB model.Template, templateSend whatsapp_business.MessageComponent, templateFields map[string]model.TemplateField) error {
	validationInput, err := parseTemplateValidationInput(templateDB, templateSend)
	if err != nil {
		return err
	}

	if validationInput.parameterFormat == nil {
		if len(validationInput.sendComponents) > 0 {
			return fmt.Errorf("template does not have components but components found in the payload")
		}
		return nil
	}

	dbParsedParameter, err := parseDBTemplateComponents(validationInput.dbComponents)
	if err != nil {
		return fmt.Errorf("failed to parse template components from database: %v", err)
	}

	if err := validateSendComponents(client, *validationInput.parameterFormat, validationInput.sendComponents, templateFields); err != nil {
		return err
	}

	sendParsedParameter, err := ws.ExtractSendComponentParameterValues(*validationInput.parameterFormat, validationInput.sendComponents)
	if err != nil {
		return err
	}

	if err := validateParameterMatch(dbParsedParameter, sendParsedParameter); err != nil {
		return err
	}
	return nil
}

func (ws *WhatsappBusiness) ExtractSendComponentParameterValues(parameterFormat string, sendComponents []map[string]any) (map[string]map[string]string, error) {
	parameterValues := map[string]map[string]string{
		"HEADER":  {},
		"BODY":    {},
		"FOOTER":  {},
		"BUTTONS": {},
	}
	if parameterFormat == "" {
		return parameterValues, fmt.Errorf("parameter format is required when components are present")
	}

	for _, component := range sendComponents {
		componentType, err := normalizeComponentType(component)
		if err != nil {
			return parameterValues, err
		}
		if _, exists := parameterValues[componentType]; !exists {
			return parameterValues, fmt.Errorf("unsupported component type: %s", componentType)
		}

		componentParametersAny, err := extractArrayField(component, "parameters")
		if err != nil {
			return parameterValues, fmt.Errorf("parameters field is required but missing or not an array in the payload for component type %s", componentType)
		}

		componentParameters := make([]map[string]any, 0, len(componentParametersAny))
		for _, p := range componentParametersAny {
			param, ok := p.(map[string]any)
			if !ok {
				return parameterValues, fmt.Errorf("invalid parameter format in the payload for component type %s, expected array of objects", componentType)
			}
			componentParameters = append(componentParameters, param)
		}

		switch componentType {
		case "HEADER":
			for i, p := range componentParameters {
				componentParameterType, err := extractStringField(p, "type")
				if err != nil {
					return parameterValues, fmt.Errorf("type field is required for parameters in %s but missing or not a string", componentType)
				}

				if componentParameterType == "text" {
					componentText, err := extractStringField(p, "text")
					if err != nil {
						return parameterValues, fmt.Errorf("text field is required for parameters in %s but missing or not a string", componentType)
					}
					if parameterFormat == "NAMED" {
						parameterName, err := extractStringField(p, "parameter_name")
						if err != nil {
							return parameterValues, fmt.Errorf("parameter_name is required for text component in %s with NAMED parameter format but missing or not a string", componentType)
						}
						parameterValues[componentType][parameterName] = componentText
					} else {
						parameterValues[componentType][fmt.Sprintf("%d", i+1)] = componentText
					}
					continue
				}

				componentParameterPayload, err := extractMapField(p, componentParameterType)
				if err != nil {
					return parameterValues, fmt.Errorf("%s field is required for parameters in %s but missing or invalid", componentParameterType, componentType)
				}
				if whatsapp_business.IsMediaMessageType(componentParameterType) {
					mediaID, ok := componentParameterPayload["id"].(string)
					if !ok {
						return parameterValues, fmt.Errorf("media id is required for media component in %s but missing or not a string", componentType)
					}
					parameterValues[componentType]["mediatype_db_"+componentParameterType] = mediaID
				}
			}
		case "BODY", "FOOTER":
			for i, p := range componentParameters {
				componentText, err := extractStringField(p, "text")
				if err != nil {
					return parameterValues, fmt.Errorf("text field is required for parameters in %s but missing or not a string", componentType)
				}
				if parameterFormat == "NAMED" {
					parameterName, err := extractStringField(p, "parameter_name")
					if err != nil {
						return parameterValues, fmt.Errorf("parameter_name is required for parameters in %s with NAMED parameter format but missing or not a string", componentType)
					}
					parameterValues[componentType][parameterName] = componentText
				} else {
					parameterValues[componentType][fmt.Sprintf("%d", i+1)] = componentText
				}
			}
		case "BUTTONS":
			continue
		}
	}

	return parameterValues, nil
}
