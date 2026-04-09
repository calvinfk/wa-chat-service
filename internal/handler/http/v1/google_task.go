package http_v1

import (
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"

	"github.com/gofiber/fiber/v3"
)

type GoogleTaskHandler struct {
	googleTaskUsecase usecase.GoogleTask
}

func NewGoogleTaskHandler(googleTaskUsecase usecase.GoogleTask) *GoogleTaskHandler {
	return &GoogleTaskHandler{
		googleTaskUsecase: googleTaskUsecase,
	}
}

func (h *GoogleTaskHandler) RegisterRoute(api fiber.Router) {
	googleTaskRoute := api.Group("/google-task")
	{
		googleTaskRoute.Post("/ping", h.createPingTask)
	}
}

func (h *GoogleTaskHandler) createPingTask(ctx fiber.Ctx) error {
	// Implementation for creating a Google Task
	serverError, err := h.googleTaskUsecase.CreatePingTask()
	httpCode, apiResponse := api_response.NewApiResponse(serverError, err, "Successfully created Google Task", nil)
	return ctx.Status(httpCode).JSON(apiResponse)
}
