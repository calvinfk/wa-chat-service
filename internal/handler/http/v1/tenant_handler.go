package http_v1

import (
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/handler/http/middleware"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/filter_request"

	"github.com/gofiber/fiber/v3"
)

type TenantHandler struct {
	tenantUsecase usecase.Tenant
}

func NewTenantHandler(tenantUsecase usecase.Tenant) *TenantHandler {
	return &TenantHandler{
		tenantUsecase: tenantUsecase,
	}
}

func (h *TenantHandler) RegisterRoute(router fiber.Router) {
	tenantRoutes := router.Group("/tenant")
	{
		tenantContactRoutes := tenantRoutes.Group("/contact")
		{
			tenantContactRoutes.Post("/create", middleware.Protected(), h.createContact)
			tenantContactRoutes.Get("/filter", middleware.Protected(), h.getFiltered)
			tenantContactRoutes.Put("/update", middleware.Protected(), h.updateContact)
			tenantContactRoutes.Delete("/delete", middleware.Protected(), h.deleteContact)
		}
	}
}

func (h *TenantHandler) createContact(ctx fiber.Ctx) error {
	var inputData dto.ContactCreateRequest
	if err := ctx.Bind().Body(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	serverError, err := h.tenantUsecase.CreateContact(ctx.Context(), authData.TenantID, inputData)
	if err != nil {
		httpCode, apiResponse := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	httpCode, apiResponse := api_response.NewApiResponse("Successfully created contact", nil)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *TenantHandler) getFiltered(ctx fiber.Ctx) error {
	var inputData filter_request.FilterRequest[dto.ContactGetFilteredRequest]
	if err := ctx.Bind().Query(&inputData.SpecificFilter); err != nil {
		httpCode, apiResponse := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	if err := ctx.Bind().Query(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	data, serverError, err := h.tenantUsecase.GetContactsFiltered(ctx.Context(), authData.TenantID, inputData)
	if err != nil {
		httpCode, apiResponse := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	httpCode, apiResponse := api_response.NewApiResponse("Successfully retrieved contacts", data)
	return ctx.Status(httpCode).JSON(apiResponse)

}

func (h *TenantHandler) updateContact(ctx fiber.Ctx) error {
	var inputData dto.ContactUpdateRequest
	if err := ctx.Bind().All(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	serverError, err := h.tenantUsecase.UpdateContact(ctx.Context(), authData.TenantID, inputData)
	if err != nil {
		httpCode, apiResponse := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	httpCode, apiResponse := api_response.NewApiResponse("Successfully updated contact", nil)
	return ctx.Status(httpCode).JSON(apiResponse)
}

func (h *TenantHandler) deleteContact(ctx fiber.Ctx) error {
	var inputData dto.ContactDeleteRequest
	if err := ctx.Bind().Query(&inputData); err != nil {
		httpCode, apiResponse := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	serverError, err := h.tenantUsecase.DeleteContact(ctx.Context(), authData.TenantID, inputData)
	if err != nil {
		httpCode, apiResponse := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(httpCode).JSON(apiResponse)
	}
	httpCode, apiResponse := api_response.NewApiResponse("Successfully deleted contact", nil)
	return ctx.Status(httpCode).JSON(apiResponse)
}
