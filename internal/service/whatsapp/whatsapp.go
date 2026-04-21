package whatsapp_service

import (
	"fmt"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/meta/whatsapp_business/template_components"

	"go.uber.org/zap"
)

type WhatsappBusiness struct {
	zslog *zap.SugaredLogger
}

func NewWhatsappService(zslog *zap.SugaredLogger) *WhatsappBusiness {
	return &WhatsappBusiness{
		zslog: zslog,
	}
}

func (s *WhatsappBusiness) ValidateTemplatePayload(client *whatsapp_business.Client, templateDB model.Template, templateSend whatsapp_business.MessageComponent) error {
	validationInput, err := parseTemplateValidationInput(templateDB, templateSend)
	if err != nil {
		return err
	}

	if validationInput.parameterFormat == nil {
		defaultParameter := "POSITIONAL"
		validationInput.parameterFormat = &defaultParameter
	}

	dbParsedParameter, err := parseDBTemplateComponents(validationInput.dbComponents)
	if err != nil {
		s.zslog.Errorf("[ValidateTemplatePayload] Failed to parse template components from database: %v", err)
		return fmt.Errorf("failed to parse template components from database: %v", err)
	}

	if err := validateSendComponents(client, *validationInput.parameterFormat, validationInput.sendComponents); err != nil {
		if waErr, ok := err.(whatsapp_business.WhatsAppBusinessError); ok {
			s.zslog.Errorf("[ValidateTemplatePayload] WhatsApp Business API error: %s (code: %d, subcode: %d)", waErr.ErrorData.Message, waErr.ErrorData.Code, waErr.ErrorData.ErrorSubcode)
			return waErr
		}
		s.zslog.Errorf("[ValidateTemplatePayload] Failed to validate send components: %v", err)
		return err
	}

	sendParsedParameter, err := s.ExtractSendComponentParameterValues(*validationInput.parameterFormat, validationInput.sendComponents)
	if err != nil {
		return err
	}

	if err := validateParameterMatch(dbParsedParameter, sendParsedParameter); err != nil {
		s.zslog.Errorf("[ValidateTemplatePayload] Template parameter validation failed: %v", err)
		return err
	}
	return nil
}

func (s *WhatsappBusiness) ExtractSendComponentParameterValues(parameterFormat string, sendComponents []map[string]any) (map[string]map[string]string, error) {
	parameterValues := map[string]map[string]string{
		"HEADER": {},
		"BODY":   {},
		"FOOTER": {},
		"BUTTON": {},
	}
	if parameterFormat == "" {
		s.zslog.Errorf("[ExtractSendComponentParameterValues] parameter format is required when components are present")
		return parameterValues, fmt.Errorf("parameter format is required when components are present")
	}

	for _, component := range sendComponents {
		componentType, err := normalizeComponentType(component)
		if err != nil {
			s.zslog.Errorf("[ExtractSendComponentParameterValues] Failed to normalize component type: %v", err)
			return parameterValues, err
		}
		if _, exists := parameterValues[componentType]; !exists {
			s.zslog.Errorf("[ExtractSendComponentParameterValues] Unsupported component type: %s", componentType)
			return parameterValues, fmt.Errorf("unsupported component type: %s", componentType)
		}
		componentParametersAny, err := extractArrayField(component, "parameters")
		if err != nil {
			s.zslog.Errorf("[ExtractSendComponentParameterValues] parameters field is required but missing or not an array in the payload for component type %s", componentType)
			return parameterValues, fmt.Errorf("parameters field is required but missing or not an array in the payload for component type %s", componentType)
		}

		componentParameters := make([]map[string]any, 0, len(componentParametersAny))
		for _, p := range componentParametersAny {
			param, ok := p.(map[string]any)
			if !ok {
				s.zslog.Errorf("[ExtractSendComponentParameterValues] invalid parameter format in the payload for component type %s, expected array of objects", componentType)
				return parameterValues, fmt.Errorf("invalid parameter format in the payload for component type %s, expected array of objects", componentType)
			}
			componentParameters = append(componentParameters, param)
		}

		switch componentType {
		case "HEADER":
			for i, p := range componentParameters {
				componentParameterType, err := extractStringField(p, "type")
				if err != nil {
					s.zslog.Errorf("[ExtractSendComponentParameterValues] type field is required for parameters in %s but missing or not a string", componentType)
					return parameterValues, fmt.Errorf("type field is required for parameters in %s but missing or not a string", componentType)
				}

				if componentParameterType == "text" {
					componentText, err := extractStringField(p, "text")
					if err != nil {
						s.zslog.Errorf("[ExtractSendComponentParameterValues] text field is required for parameters in %s but missing or not a string", componentType)
						return parameterValues, fmt.Errorf("text field is required for parameters in %s but missing or not a string", componentType)
					}
					if parameterFormat == "NAMED" {
						parameterName, err := extractStringField(p, "parameter_name")
						if err != nil {
							s.zslog.Errorf("[ExtractSendComponentParameterValues] parameter_name is required for text component in %s with NAMED parameter format but missing or not a string", componentType)
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
					s.zslog.Errorf("[ExtractSendComponentParameterValues] %s field is required for parameters in %s but missing or invalid", componentParameterType, componentType)
					return parameterValues, fmt.Errorf("%s field is required for parameters in %s but missing or invalid", componentParameterType, componentType)
				}
				if whatsapp_business.IsMediaMessageType(componentParameterType) {
					mediaID, ok := componentParameterPayload["id"].(string)
					if !ok {
						s.zslog.Errorf("[ExtractSendComponentParameterValues] media id is required for media component in %s but missing or not a string", componentType)
						return parameterValues, fmt.Errorf("media id is required for media component in %s but missing or not a string", componentType)
					}
					parameterValues[componentType]["mediatype_db_"+componentParameterType] = mediaID
				}
			}
		case "BODY", "FOOTER":
			for i, p := range componentParameters {
				componentText, err := extractStringField(p, "text")
				if err != nil {
					s.zslog.Errorf("[ExtractSendComponentParameterValues] text field is required for parameters in %s but missing or not a string", componentType)
					return parameterValues, fmt.Errorf("text field is required for parameters in %s but missing or not a string", componentType)
				}
				if parameterFormat == "NAMED" {
					parameterName, err := extractStringField(p, "parameter_name")
					if err != nil {
						s.zslog.Errorf("[ExtractSendComponentParameterValues] parameter_name is required for parameters in %s with NAMED parameter format but missing or not a string", componentType)
						return parameterValues, fmt.Errorf("parameter_name is required for parameters in %s with NAMED parameter format but missing or not a string", componentType)
					}
					parameterValues[componentType][parameterName] = componentText
				} else {
					parameterValues[componentType][fmt.Sprintf("%d", i+1)] = componentText
				}
			}
		case "BUTTON":
			buttonComponent, err := whatsapp_business.NewTemplateSendButton(component)
			if err != nil {
				s.zslog.Errorf("[ExtractSendComponentParameterValues] Failed to parse button component: %v", err)
				return parameterValues, err
			}
			if buttonComponent.GetSubType() == "QUICK_REPLY" {
				quickReplyButton, ok := buttonComponent.(*template_components.SendQuickReplyButton)
				if !ok {
					s.zslog.Errorf("[ExtractSendComponentParameterValues] Failed to assert button component as QuickReplyButton for component: %v", component)
					return parameterValues, fmt.Errorf("failed to assert button component as QuickReplyButton")
				}
				parameterValues[componentType][quickReplyButton.Index] = quickReplyButton.Parameters[0].Payload
			}
		}
	}

	return parameterValues, nil
}

func (s *WhatsappBusiness) ParseTemplateComponentParameter(value string) string {
	matches := templateParameterRegex.FindStringSubmatch(value)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}
