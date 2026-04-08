package http_v1

import (
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"

	"github.com/gofiber/fiber/v3"
)

type TemplateHandler struct {
	templateUsecase usecase.Template
}

func NewTemplateHandler(templateUsecase usecase.Template) *TemplateHandler {
	return &TemplateHandler{
		templateUsecase: templateUsecase,
	}
}

func (h *TemplateHandler) RegisterRoute(router fiber.Router) {
	templateRoutes := router.Group("/template")
	{
		templateRoutes.Post("/create", h.createTemplate)
	}
}

func (h *TemplateHandler) createTemplate(ctx fiber.Ctx) error {
	var inputData dto.TemplateCreateRequest
	if err := ctx.Bind().Body(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "Failed to parse request body", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	data, serverError, err := h.templateUsecase.CreateTemplate(ctx, inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully created template", data)
	return ctx.Status(httpCode).JSON(apiResponse)
}
