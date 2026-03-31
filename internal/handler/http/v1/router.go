package http_v1

import (
	"net/http"
	"wa_chat_service/config"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/formatter"

	"github.com/gofiber/fiber/v3"
)

// ApiResponse defines the structure of the standardized API response that will be returned for all HTTP requests handled by this version of the API. It includes a status code, a message, optional data, and optional error details.
type ApiResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    any               `json:"data"`
	Errors  map[string]string `json:"errors"`
}

// NewApiResponse creates a standardized API response based on the provided parameters. It determines the appropriate HTTP status code and message based on whether there was a server error, a client error, or a successful operation, and formats the response accordingly.
func NewApiResponse(serverError bool, err error, successMessage string, data any, errors map[string]string) (int, ApiResponse) {
	var response ApiResponse
	if serverError {
		response.Code = http.StatusInternalServerError
		response.Message = "Internal server error"
		response.Data = nil
		response.Errors = nil
		return response.Code, response
	}
	if err != nil {
		if err == errs.ErrGenericForbidden {
			response.Code = http.StatusForbidden
			response.Data = nil
		} else {
			response.Code = http.StatusBadRequest
			response.Data = nil
		}
		if len(errors) > 0 {
			response.Errors = errors
		}
		response.Message = formatter.CapitalizeFirstLetter(err.Error())
	} else {
		response.Code = http.StatusOK
		response.Message = successMessage
		response.Data = data
		response.Errors = nil
	}
	return response.Code, response
}

type RouterHandlerV1 struct {
	ActivityLogUsecase    usecase.ActivityLog
	AccessTokenService    service.AccessToken
	EncryptService        service.Encrypt
	GoogleStorageService  service.GoogleStorage
	GoogleFirebaseService service.GoogleFirebase
}

type V1 struct {
	api          fiber.Router
	apiProtected fiber.Router
}

func NewApiV1Routes(api fiber.Router, routerHandler RouterHandlerV1, cfg *config.Config) {
	// v1 := &V1{
	// 	api:          api,
	// }
}
