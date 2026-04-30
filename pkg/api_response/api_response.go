package api_response

import (
	"net/http"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/utils"

	"github.com/go-playground/validator/v10"
)

type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Errors  any    `json:"errors"`
}

// NewApiResponse creates a standardized API response based on the provided parameters.
// It takes a success message and data, and returns an HTTP status code along with the ApiResponse struct.
func NewApiResponse(successMessage string, data any) (int, ApiResponse) {
	var response ApiResponse
	response.Code = http.StatusOK
	response.Message = successMessage
	response.Data = data
	response.Errors = nil
	return response.Code, response
}

// TODO: Refactor this function to handle different types of errors more elegantly
// possibly by defining custom error types or using error wrapping to provide more context about the errors.
func NewErrorApiResponse(serverError bool, err any) (int, ApiResponse) {
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
			errors := utils.FormatErrors(validationErrors)
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
			case errs.ErrGenericGone:
				response.Code = http.StatusGone
				response.Data = nil
			case errs.ErrGenericRangeNotSatisfiable:
				response.Code = http.StatusRequestedRangeNotSatisfiable
				response.Data = nil
			default:
				response.Code = http.StatusBadRequest
				response.Data = nil
				response.Errors = nil
			}
			response.Message = utils.CapitalizeFirstLetter(errors.Error())
		} else {
			response.Code = http.StatusBadRequest
			response.Data = nil
			response.Errors = err
			response.Message = "Bad request"
		}
	}
	return response.Code, response
}
