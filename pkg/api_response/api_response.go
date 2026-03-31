package api_response

import (
	"net/http"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/formatter"

	"github.com/go-playground/validator/v10"
)

type ApiResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    any               `json:"data"`
	Errors  map[string]string `json:"errors"`
}

// NewApiResponse creates a standardized API response based on the provided parameters. It determines the appropriate HTTP status code and message based on whether there was a server error, a client error, or a successful operation, and formats the response accordingly.
func NewApiResponse(serverError bool, err error, successMessage string, data any) (int, ApiResponse) {
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
			response.Code = http.StatusBadRequest
			response.Data = nil
			errors := make(map[string]string)
			for _, fieldErr := range validationErrors {
				errors[fieldErr.Field()] = fieldErr.Error()
			}
			response.Errors = errors
			response.Message = "Validation error"
		} else {
			switch err {
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
			response.Message = formatter.CapitalizeFirstLetter(err.Error())
		}
	} else {
		response.Code = http.StatusOK
		response.Message = successMessage
		response.Data = data
		response.Errors = nil
	}
	return response.Code, response
}
