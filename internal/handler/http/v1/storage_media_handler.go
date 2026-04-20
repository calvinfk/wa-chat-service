package http_v1

import (
	"fmt"
	"io"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/handler/http/middleware"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/filter_request"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type StorageMediaHandler struct {
	storageMediaUsecase usecase.StorageMedia
	zslog               *zap.SugaredLogger
}

func NewStorageMediaHandler(storageMediaUsecase usecase.StorageMedia, zslog *zap.SugaredLogger) *StorageMediaHandler {
	return &StorageMediaHandler{
		storageMediaUsecase: storageMediaUsecase,
		zslog:               zslog,
	}
}

func (h *StorageMediaHandler) RegisterRoutes(router fiber.Router) {
	storageMediaRouter := router.Group("/storage-media")
	{
		storageMediaRouter.Post("/upload", middleware.Protected(), h.uploadMedia)
		storageMediaRouter.Get("/get", middleware.Protected(), h.getMedia)
		storageMediaRouter.Delete("/delete", middleware.Protected(), h.deleteMedia)
		// storageMediaRouter.Post("/save-media-id", middleware.Protected(), h.saveMediaID)
		// storageMediaRouter.Post("/resumable", middleware.Protected(), h.uploadResumableMedia)
		// storageMediaRouter.Post("/upload-meta", middleware.Protected(), h.uploadMediaMeta)
		storageMediaRouter.Get("/list", middleware.Protected(), h.getMediaList)
	}
}

func (h *StorageMediaHandler) uploadMedia(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaUploadRequest
	if err := ctx.Bind().All(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	if requestData.File == nil {
		code, response := api_response.NewApiResponse(false, fmt.Errorf("file is required"), "", nil)
		return ctx.Status(code).JSON(response)
	}
	media, serverError, err := h.storageMediaUsecase.UploadMedia(ctx.Context(), requestData)
	code, response := api_response.NewApiResponse(serverError, err, "Media uploaded successfully", media)
	return ctx.Status(code).JSON(response)
}

func (h *StorageMediaHandler) getMedia(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaGetRequest
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	// Stream the file from Google Cloud Storage
	mediaResponse, serverError, err := h.storageMediaUsecase.GetMedia(ctx.Context(), requestData)
	if serverError || err != nil {
		code, response := api_response.NewApiResponse(serverError, err, "Failed to retrieve media", nil)
		return ctx.Status(code).JSON(response)
	}
	defer mediaResponse.Reader.Close() // Ensure the reader is closed after streaming

	// 4. Set Headers to hide GCS and show your own info
	ctx.Set("Content-Type", mediaResponse.ContentType)
	ctx.Set("Content-Disposition", "inline; filename="+mediaResponse.FileName)
	ctx.Set("Cache-Control", "private, max-age=3600")
	ctx.Set("Content-Length", fmt.Sprintf("%d", mediaResponse.Size))

	if _, err := io.Copy(ctx.Response().BodyWriter(), mediaResponse.Reader); err != nil {
		h.zslog.Error("[getMedia] Failed to stream file to response:", err)
		return err
	}
	return nil
}

func (h *StorageMediaHandler) deleteMedia(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaDeleteRequest
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	deleted, err := h.storageMediaUsecase.DeleteMedia(ctx.Context(), requestData)
	code, response := api_response.NewApiResponse(false, err, "Media deleted successfully", deleted)
	return ctx.Status(code).JSON(response)
}

func (h *StorageMediaHandler) saveMediaID(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaSaveMediaIDRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	data, serverError, err := h.storageMediaUsecase.SaveMediaID(ctx.Context(), requestData)
	code, response := api_response.NewApiResponse(serverError, err, "Media uploaded successfully", data)
	return ctx.Status(code).JSON(response)
}

func (h *StorageMediaHandler) getMediaList(ctx fiber.Ctx) error {
	var requestData filter_request.FilterRequest[dto.StorageMediaGetListRequest]
	if err := ctx.Bind().Query(&requestData.SpecificFilter); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	response, serverError, err := h.storageMediaUsecase.GetFiltered(ctx.Context(), requestData)
	code, apiResponse := api_response.NewApiResponse(serverError, err, "Media list retrieved successfully", response)
	return ctx.Status(code).JSON(apiResponse)

}
