package whatsapp_service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/utils"
)

type ParsedParameter struct {
	Name  string
	Value string
}

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
	templateSendPayload := templateSend.GetPayload()
	if templateSendPayload["template"] == nil {
		return fmt.Errorf("template payload is required but missing")
	}
	sendPayload := templateSendPayload["template"].(map[string]any)

	// parse template components
	var templateDBComponents []map[string]any
	err = json.Unmarshal([]byte(templateDB.Components), &templateDBComponents)
	if err != nil {
		return fmt.Errorf("failed to unmarshal template components: %v", err)
	}
	sendComponents := make([]map[string]any, 0)
	sendComponentsAny := sendPayload["components"].([]any)
	for _, c := range sendComponentsAny {
		if component, ok := c.(map[string]any); ok {
			sendComponents = append(sendComponents, component)
		} else {
			return fmt.Errorf("invalid component format in send payload, expected array of objects")
		}
	}
	err = ws.validateTemplateParameter(client, templateDB.ParameterFormat, templateDBComponents, sendComponents)
	if err != nil {
		return fmt.Errorf("template parameter validation failed: %v", err)
	}
	return nil
}

func (ws *WhatsappBusiness) parseDBTemplateComponents(components []map[string]any) (map[string]map[string]any, error) {
	parsedParameter := map[string]map[string]any{
		"HEADER":  make(map[string]any),
		"BODY":    make(map[string]any),
		"FOOTER":  make(map[string]any),
		"BUTTONS": make(map[string]any),
	}
	for _, component := range components {
		componentType, ok := component["type"].(string)
		if !ok {
			return nil, fmt.Errorf("component type is required but missing in the payload")
		}
		// log.Println("[DEBUG][internal/service/whatsapp/whatsapp.go][parseDBTemplateComponents] Parsing component:", component)
		componentType = strings.ToUpper(componentType)
		switch componentType {
		case "HEADER", "BODY", "FOOTER":
			if component["text"] != nil {
				componentText := component["text"].(string)
				regexCountParam := regexp.MustCompile(`{{\s*([^{}\s]+)\s*}}`)
				matches := regexCountParam.FindAllStringSubmatch(componentText, -1)
				for _, match := range matches {
					parsedParameter[componentType][match[1]] = ""
				}
			}
		}
	}
	return parsedParameter, nil
}

// returns the number of parameters for each component type, e.g. {"header": 1, "body": 2, "button": 3}
func (ws *WhatsappBusiness) validateSendComponents(whatsappClient *whatsapp_business.Client, parameterFormat string, sendComponents []map[string]any) (map[string]map[string]any, error) {
	fillableParameterCount := map[string]map[string]any{
		"HEADER":  {},
		"BODY":    {},
		"FOOTER":  {},
		"BUTTONS": {},
	}
	if len(sendComponents) == 0 {
		log.Println("[DEBUG][internal/service/whatsapp/whatsapp.go][validateSendComponents] No components in send payload, skipping parameter validation")
		return fillableParameterCount, nil
	}
	if parameterFormat == "" {
		return fillableParameterCount, fmt.Errorf("parameter format is required when components are present")
	}
	maxNumOfComponents := map[string]int{
		"HEADER":  1,
		"BODY":    1,
		"FOOTER":  1,
		"BUTTONS": 10,
	}
	for _, component := range sendComponents {
		componentType, ok := component["type"].(string)
		if !ok {
			return fillableParameterCount, fmt.Errorf("component type is required but missing in the payload")
		}
		componentType = strings.ToUpper(componentType)
		if _, exists := fillableParameterCount[componentType]; !exists {
			return fillableParameterCount, fmt.Errorf("unsupported component type: %s", componentType)
		}
		componentParametersAny, ok := component["parameters"].([]any)
		if !ok {
			return fillableParameterCount, fmt.Errorf("parameters field is required but missing or not an array in the payload for component type %s", componentType)
		}
		componentParameters := make([]map[string]any, 0)
		for _, p := range componentParametersAny {
			if param, ok := p.(map[string]any); ok {
				componentParameters = append(componentParameters, param)
			} else {
				return fillableParameterCount, fmt.Errorf("invalid parameter format in the payload for component type %s, expected array of objects", componentType)
			}
		}
		switch componentType {
		case "HEADER":
			for _, p := range componentParameters {
				componentParameterType := p["type"].(string)
				var componentParameterPayload map[string]any
				if componentParameterType == "text" {
					componentParameterPayload = map[string]any{
						"body": p["text"],
					}
				} else {
					componentParameterPayload = p[componentParameterType].(map[string]any)
				}
				_, err := whatsapp_business.NewComponent(componentParameterType, componentParameterPayload)
				if err != nil {
					return fillableParameterCount, fmt.Errorf("failed to parse component for type %s.%s: %v", componentType, componentParameterType, err)
				}
				if whatsapp_business.IsMediaMessageType(componentParameterType) {
					mediaID, ok := componentParameterPayload["id"].(string)
					if !ok {
						return fillableParameterCount, fmt.Errorf("media id is required for media component in %s but missing or not a string", componentType)
					}
					httpCode, err := ws.validateMediaID(whatsappClient, mediaID)
					if err != nil || httpCode != http.StatusOK {
						return fillableParameterCount, fmt.Errorf("failed to validate media ID '%s' in %s: %v (HTTP code: %d)", mediaID, componentType, err, httpCode)
					}
				} else if componentParameterType == "text" {
					if parameterFormat == "NAMED" {
						parameterName, ok := p["parameter_name"].(string)
						if !ok {
							return fillableParameterCount, fmt.Errorf("parameter_name is required for text component in %s with NAMED parameter format but missing or not a string", componentType)
						}
						fillableParameterCount[componentType][parameterName] = ""
					} else {
						fillableParameterCount[componentType][fmt.Sprintf("%d", len(fillableParameterCount[componentType])+1)] = ""
					}
				}
			}
			maxNumOfComponents[componentType]--
		case "BODY", "FOOTER":
			for i, p := range componentParameters {
				componentText, ok := p["text"].(string)
				if !ok {
					return fillableParameterCount, fmt.Errorf("text field is required for parameters in %s but missing or not a string", componentType)
				}
				if parameterFormat == "NAMED" {
					fillableParameterCount[componentType][p["parameter_name"].(string)] = componentText
				} else {
					fillableParameterCount[componentType][fmt.Sprintf("%d", i+1)] = componentText
				}
			}
			maxNumOfComponents[componentType]--
		case "BUTTONS":
			maxNumOfComponents[componentType]--
		}
	}

	if maxNumOfComponents["BODY"] != 0 {
		return fillableParameterCount, fmt.Errorf("body component is required but missing in the payload")
	} else if maxNumOfComponents["HEADER"] < 0 {
		return fillableParameterCount, fmt.Errorf("too many header components in the payload, maximum allowed is 1")
	} else if maxNumOfComponents["FOOTER"] < 0 {
		return fillableParameterCount, fmt.Errorf("too many footer components in the payload, maximum allowed is 1")
	} else if maxNumOfComponents["BUTTONS"] < 0 {
		return fillableParameterCount, fmt.Errorf("too many button components in the payload, maximum allowed is 10")
	}

	return fillableParameterCount, nil
}

func (ws *WhatsappBusiness) validateTemplateParameter(whatsappClient *whatsapp_business.Client, parameterFormat *string, dbComponents, sendComponents []map[string]any) error {
	// Check parameter format
	if parameterFormat == nil {
		if len(sendComponents) > 0 {
			return fmt.Errorf("template does not have components but components found in the payload")
		}
		return nil // if parameter format is not defined, skip parameter validation
	}
	dbParsedParameter, err := ws.parseDBTemplateComponents(dbComponents)
	if err != nil {
		return fmt.Errorf("failed to parse template components from database: %v", err)
	}
	// Count parameters in send components
	data, err := ws.validateSendComponents(whatsappClient, *parameterFormat, sendComponents)
	if err != nil {
		return err
	}
	for componentType, parameters := range dbParsedParameter {
		if len(parameters) != len(data[componentType]) {
			return fmt.Errorf("parameter count mismatch for component type %s: expected %d, got %d", componentType, len(parameters), len(data[componentType]))
		}
		for paramName := range parameters {
			if _, exists := data[componentType][paramName]; !exists {
				return fmt.Errorf("missing parameter '%s' for component type %s in the payload", paramName, componentType)
			}
		}
	}
	// log.Printf("[DEBUG][internal/service/whatsapp/whatsapp.go][validateTemplateParameter] Fillable parameter count in send payload: %+v", data)
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
