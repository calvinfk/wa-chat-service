package http_v1

import (
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/filter_request"

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
		templateRoutes.Get("/get", h.getTemplates)
		templateRoutes.Get("/get-meta", h.getTemplatesMeta)
		templateRoutes.Post("/create", h.createTemplate)
		templateRoutes.Post("/sync", h.syncTemplate)
		templateRoutes.Delete("/delete", h.deleteTemplate)
	}
}

func (h *TemplateHandler) createTemplate(ctx fiber.Ctx) error {
	var inputData dto.TemplateCreateRequest
	if err := ctx.Bind().Body(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	data, serverError, err := h.templateUsecase.CreateTemplate(ctx, inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully created template", data)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *TemplateHandler) getTemplates(ctx fiber.Ctx) error {
	var inputData filter_request.FilterRequest[dto.TemplateGetByPhoneNumberID]
	if err := ctx.Bind().Query(&inputData.SpecificFilter); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	if err := ctx.Bind().Query(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	data, serverError, err := h.templateUsecase.GetFilteredByPhoneNumberID(ctx, inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully get templates", data)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *TemplateHandler) syncTemplate(ctx fiber.Ctx) error {
	var inputData dto.TemplateSyncRequest
	if err := ctx.Bind().Body(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	serverError, err := h.templateUsecase.SyncTemplate(ctx, inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully sync template", nil)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *TemplateHandler) getTemplatesMeta(ctx fiber.Ctx) error {
	var inputData dto.TemplateGetByPhoneNumberID
	if err := ctx.Bind().Query(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	data, serverError, err := h.templateUsecase.GetTemplatesMeta(ctx, inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully get templates meta", data)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *TemplateHandler) deleteTemplate(ctx fiber.Ctx) error {
	var inputData dto.TemplateDeleteRequest
	if err := ctx.Bind().Query(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	serverError, err := h.templateUsecase.DeleteTemplate(ctx, inputData)
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully delete template", nil)
	return ctx.Status(httpCode).JSON(apiResponse)
}
