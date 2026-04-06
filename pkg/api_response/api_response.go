package api_response

import (
	"net/http"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/meta/whatsapp_business"

	"github.com/go-playground/validator/v10"
)

type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	// Errors  map[string]string `json:"errors"`
	Errors any `json:"errors"`
}

// NewApiResponse creates a standardized API response based on the provided parameters. It determines the appropriate HTTP status code and message based on whether there was a server error, a client error, or a successful operation, and formats the response accordingly.
func NewApiResponse(serverError bool, err any, successMessage string, data any) (int, ApiResponse) {
	var response ApiResponse
	if serverError {
		response.Code = http.StatusInternalServerError
		response.Message = "Internal server error"
		response.Data = nil
		response.Errors = nil
		return response.Code, response
	}
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := formatter.FormatErrors(validationErrors, data)
			response.Code = http.StatusBadRequest
			response.Data = nil
			response.Errors = errors
			response.Message = "Validation error"
		} else if waErrors, ok := err.(whatsapp_business.WhatsAppBusinessError); ok {
			response.Code = http.StatusBadRequest
			response.Data = nil
			response.Errors = waErrors
			response.Message = "WhatsApp Business API error"

		} else if errMap, ok := err.(map[string]string); ok {
			response.Code = http.StatusBadRequest
			response.Data = nil
			response.Errors = errMap
			response.Message = "Validation error"
		} else if errors, ok := err.(error); ok {
			switch errors {
			case errs.ErrGenericForbidden:
				response.Code = http.StatusForbidden
				response.Data = nil
			case errs.ErrGenericNotFound:
				response.Code = http.StatusNotFound
				response.Data = nil
			default:
				response.Code = http.StatusBadRequest
				response.Data = nil
				response.Errors = nil
			}
			response.Message = formatter.CapitalizeFirstLetter(errors.Error())
		} else {
			response.Code = http.StatusBadRequest
			response.Data = nil
			response.Errors = err
			response.Message = "Bad request"
		}
	} else {
		response.Code = http.StatusOK
		response.Message = successMessage
		response.Data = data
		response.Errors = nil
	}
	return response.Code, response
}
