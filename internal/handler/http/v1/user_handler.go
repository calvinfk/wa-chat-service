package http_v1

import (
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/handler/http/middleware"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/filter_request"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userUsecase usecase.User
}

func NewUserHandler(userUsecase usecase.User) HandlerV1 {
	return &UserHandler{
		userUsecase: userUsecase,
	}
}

func (h *UserHandler) RegisterRoute(api fiber.Router) {
	userGroup := api.Group("/user")
	{
		userGroup.Get("/list", middleware.Protected(), middleware.Role("admin"), h.getUsersByTenantID)
		userGroup.Get("/get", middleware.Protected(), middleware.Role("admin"), h.getUserByID)
		userGroup.Post("/upsert", middleware.Protected(), middleware.Role("admin"), h.upsertUser)
	}
}

func (h *UserHandler) getUsersByTenantID(ctx fiber.Ctx) error {
	var requestData filter_request.FilterRequest[dto.UserListRequest]
	if err := ctx.Bind().Query(&requestData.SpecificFilter); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	tenantID := ctx.Locals("token_sub").(dto.AuthData).TenantID
	data, serverError, err := h.userUsecase.GetByTenantIDFiltered(ctx.Context(), tenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Users retrieved successfully", data)
	return ctx.Status(code).JSON(response)
}

func (h *UserHandler) getUserByID(ctx fiber.Ctx) error {
	var requestData dto.UserGetByIDRequest
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	tenantID := ctx.Locals("token_sub").(dto.AuthData).TenantID
	data, serverError, err := h.userUsecase.GetByID(ctx.Context(), tenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("User retrieved successfully", data)
	return ctx.Status(code).JSON(response)
}

func (h *UserHandler) upsertUser(ctx fiber.Ctx) error {
	var requestData dto.UserUpsertRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	tenantID := ctx.Locals("token_sub").(dto.AuthData).TenantID
	data, serverError, err := h.userUsecase.Upsert(ctx.Context(), tenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("User upserted successfully", data)
	return ctx.Status(code).JSON(response)
}
