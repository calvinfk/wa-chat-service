package whatsapp_service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"wa_chat_service/pkg/meta/whatsapp_business"
)

func normalizeComponentType(component map[string]any) (string, error) {
	componentType, ok := component["type"].(string)
	if !ok || componentType == "" {
		return "", fmt.Errorf("component type is required but missing in the payload")
	}
	return strings.ToUpper(componentType), nil
}

func extractMapField(data map[string]any, fieldName string) (map[string]any, error) {
	value, ok := data[fieldName]
	if !ok || value == nil {
		return nil, fmt.Errorf("%s field is required but missing", fieldName)
	}
	result, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s field is invalid, expected object", fieldName)
	}
	return result, nil
}

func extractArrayField(data map[string]any, fieldName string) ([]any, error) {
	value, ok := data[fieldName]
	if !ok || value == nil {
		return nil, fmt.Errorf("%s field is required but missing", fieldName)
	}
	result, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s field is required but missing or not an array", fieldName)
	}
	return result, nil
}

func extractStringField(data map[string]any, fieldName string) (string, error) {
	value, ok := data[fieldName]
	if !ok || value == nil {
		return "", fmt.Errorf("%s field is required but missing", fieldName)
	}
	result, ok := value.(string)
	if !ok || result == "" {
		return "", fmt.Errorf("%s field is required but missing or not a string", fieldName)
	}
	return result, nil
}

func parseTemplateComponentsJSON(raw string) ([]map[string]any, error) {
	var components []map[string]any
	if err := json.Unmarshal([]byte(raw), &components); err != nil {
		return nil, err
	}
	return components, nil
}

func parseDBTemplateComponents(components []map[string]any) (map[string]map[string]any, error) {
	parsedParameter := map[string]map[string]any{
		"HEADER":  {},
		"BODY":    {},
		"FOOTER":  {},
		"BUTTONS": {},
	}
	for _, component := range components {
		componentType, err := normalizeComponentType(component)
		if err != nil {
			return nil, err
		}

		switch componentType {
		case "HEADER", "BODY", "FOOTER":
			componentTextAny, exists := component["text"]
			if !exists || componentTextAny == nil {
				continue
			}
			componentText, ok := componentTextAny.(string)
			if !ok {
				return nil, fmt.Errorf("text field is invalid in %s component, expected string", componentType)
			}
			matches := templateParameterRegex.FindAllStringSubmatch(componentText, -1)
			for _, match := range matches {
				parsedParameter[componentType][match[1]] = ""
			}
		}
	}
	return parsedParameter, nil
}

func validateSendComponents(whatsappClient *whatsapp_business.Client, parameterFormat string, sendComponents []map[string]any) (map[string]map[string]any, error) {
	fillableParameterCount := map[string]map[string]any{
		"HEADER":  {},
		"BODY":    {},
		"FOOTER":  {},
		"BUTTONS": {},
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
	currentNumOfComponents := map[string]int{
		"HEADER":  0,
		"BODY":    0,
		"FOOTER":  0,
		"BUTTONS": 0,
	}

	for _, component := range sendComponents {
		componentType, err := normalizeComponentType(component)
		if err != nil {
			return fillableParameterCount, err
		}
		if _, exists := fillableParameterCount[componentType]; !exists {
			return fillableParameterCount, fmt.Errorf("unsupported component type: %s", componentType)
		}

		componentParametersAny, err := extractArrayField(component, "parameters")
		if err != nil {
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
				componentParameterType, err := extractStringField(p, "type")
				if err != nil {
					return fillableParameterCount, fmt.Errorf("type field is required for parameters in %s but missing or not a string", componentType)
				}

				var componentParameterPayload map[string]any
				if componentParameterType == "text" {
					componentText, err := extractStringField(p, "text")
					if err != nil {
						return fillableParameterCount, fmt.Errorf("text field is required for parameters in %s but missing or not a string", componentType)
					}
					componentParameterPayload = map[string]any{"body": componentText}
				} else {
					componentParameterPayload, err = extractMapField(p, componentParameterType)
					if err != nil {
						return fillableParameterCount, fmt.Errorf("%s field is required for parameters in %s but missing or invalid", componentParameterType, componentType)
					}
				}

				_, err = whatsapp_business.NewComponent(componentParameterType, componentParameterPayload)
				if err != nil {
					return fillableParameterCount, fmt.Errorf("failed to parse component for type %s.%s: %v", componentType, componentParameterType, err)
				}

				if whatsapp_business.IsMediaMessageType(componentParameterType) {
					mediaID, ok := componentParameterPayload["id"].(string)
					if !ok {
						return fillableParameterCount, fmt.Errorf("media id is required for media component in %s but missing or not a string", componentType)
					}
					httpCode, err := validateMediaID(whatsappClient, mediaID)
					if err != nil || httpCode != http.StatusOK {
						return fillableParameterCount, fmt.Errorf("failed to validate media ID '%s' in %s: %v (HTTP code: %d)", mediaID, componentType, err, httpCode)
					}
				} else if componentParameterType == "text" {
					if parameterFormat == "NAMED" {
						parameterName, err := extractStringField(p, "parameter_name")
						if err != nil {
							return fillableParameterCount, fmt.Errorf("parameter_name is required for text component in %s with NAMED parameter format but missing or not a string", componentType)
						}
						fillableParameterCount[componentType][parameterName] = ""
					} else {
						fillableParameterCount[componentType][fmt.Sprintf("%d", len(fillableParameterCount[componentType])+1)] = ""
					}
				}
			}
			currentNumOfComponents[componentType]++
		case "BODY", "FOOTER":
			for i, p := range componentParameters {
				componentText, err := extractStringField(p, "text")
				if err != nil {
					return fillableParameterCount, fmt.Errorf("text field is required for parameters in %s but missing or not a string", componentType)
				}
				if parameterFormat == "NAMED" {
					parameterName, err := extractStringField(p, "parameter_name")
					if err != nil {
						return fillableParameterCount, fmt.Errorf("parameter_name is required for parameters in %s with NAMED parameter format but missing or not a string", componentType)
					}
					fillableParameterCount[componentType][parameterName] = componentText
				} else {
					fillableParameterCount[componentType][fmt.Sprintf("%d", i+1)] = componentText
				}
			}
			currentNumOfComponents[componentType]++
		case "BUTTONS":
			currentNumOfComponents[componentType]++
		}
	}

	if currentNumOfComponents["BODY"] == 0 {
		return fillableParameterCount, fmt.Errorf("body component is required but missing in the payload")
	} else if currentNumOfComponents["HEADER"] > maxNumOfComponents["HEADER"] {
		return fillableParameterCount, fmt.Errorf("too many header components in the payload, maximum allowed is 1")
	} else if currentNumOfComponents["FOOTER"] > maxNumOfComponents["FOOTER"] {
		return fillableParameterCount, fmt.Errorf("too many footer components in the payload, maximum allowed is 1")
	} else if currentNumOfComponents["BUTTONS"] > maxNumOfComponents["BUTTONS"] {
		return fillableParameterCount, fmt.Errorf("too many button components in the payload, maximum allowed is 10")
	}

	return fillableParameterCount, nil
}

func extractSendTemplateComponents(templateSendPayload map[string]any) ([]map[string]any, error) {
	templatePayloadAny, ok := templateSendPayload["template"]
	if !ok || templatePayloadAny == nil {
		return nil, fmt.Errorf("template payload is required but missing")
	}
	templatePayload, ok := templatePayloadAny.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid template payload format, expected object")
	}

	componentsAny, err := extractArrayField(templatePayload, "components")
	if err != nil {
		return nil, fmt.Errorf("components field is required but missing or not an array")
	}

	components := make([]map[string]any, 0, len(componentsAny))
	for _, component := range componentsAny {
		componentMap, ok := component.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid component format in send payload, expected array of objects")
		}
		components = append(components, componentMap)
	}

	return components, nil
}

func validateMediaID(client *whatsapp_business.Client, mediaID string) (int, error) {
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

func validateParameterMatch(dbParsedParameter, sendParsedParameter map[string]map[string]any) error {
	for componentType, parameters := range dbParsedParameter {
		if len(parameters) != len(sendParsedParameter[componentType]) {
			return fmt.Errorf("parameter count mismatch for component type %s: expected %d, got %d", componentType, len(parameters), len(sendParsedParameter[componentType]))
		}
		for paramName := range parameters {
			if _, exists := sendParsedParameter[componentType][paramName]; !exists {
				return fmt.Errorf("missing parameter '%s' for component type %s in the payload", paramName, componentType)
			}
		}
	}
	return nil
}
