package http_v1

import (
	"fmt"
	"io"
	"log"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"

	"github.com/gofiber/fiber/v3"
)

type StorageMediaHandler struct {
	storageMediaUsecase usecase.StorageMedia
}

func NewStorageMediaHandler(storageMediaUsecase usecase.StorageMedia) *StorageMediaHandler {
	return &StorageMediaHandler{
		storageMediaUsecase: storageMediaUsecase,
	}
}

func (h *StorageMediaHandler) RegisterRoutes(router fiber.Router) {
	storageMediaRouter := router.Group("/storage-media")
	{
		storageMediaRouter.Post("/upload", h.UploadMedia)
		storageMediaRouter.Get("/get", h.GetMedia)
		storageMediaRouter.Delete("/delete", h.DeleteMedia)
		storageMediaRouter.Post("/upload-media-id", h.UploadMediaUsingMediaID)
	}
}

func (h *StorageMediaHandler) UploadMedia(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaUploadRequest
	if err := ctx.Bind().Form(&requestData); err != nil {
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

func (h *StorageMediaHandler) GetMedia(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaGetRequest
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	// Stream the file from Google Cloud Storage
	rc, attrs, serverError, err := h.storageMediaUsecase.GetMedia(ctx.Context(), requestData)
	if serverError || err != nil {
		code, response := api_response.NewApiResponse(serverError, err, "Failed to retrieve media", nil)
		return ctx.Status(code).JSON(response)
	}
	defer rc.Close() // Ensure the reader is closed after streaming

	// 4. Set Headers to hide GCS and show your own info
	ctx.Set("Content-Type", attrs.ContentType)
	ctx.Set("Content-Disposition", "inline; filename="+attrs.Name)
	ctx.Set("Cache-Control", "private, max-age=3600")
	ctx.Set("Content-Length", fmt.Sprintf("%d", attrs.Size))

	if _, err := io.Copy(ctx.Response().BodyWriter(), rc); err != nil {
		log.Println("[ERROR][internal/handler/http/v1/storage_media_handler.go][GetMedia] Failed to stream file to response:", err)
		return err
	}
	return nil
}

func (h *StorageMediaHandler) DeleteMedia(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaDeleteRequest
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	deleted, err := h.storageMediaUsecase.DeleteMedia(ctx.Context(), requestData)
	code, response := api_response.NewApiResponse(false, err, "Media deleted successfully", deleted)
	return ctx.Status(code).JSON(response)
}

func (h *StorageMediaHandler) UploadMediaUsingMediaID(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaUploadUsingMediaIDRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		log.Println("[ERROR][internal/handler/http/v1/storage_media_handler.go][UploadMediaUsingMediaID] Failed to bind request data:", err)
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	id, serverError, err := h.storageMediaUsecase.UploadMediaUsingMediaID(ctx.Context(), requestData)
	code, response := api_response.NewApiResponse(serverError, err, "Media uploaded successfully", fiber.Map{"id": id})
	return ctx.Status(code).JSON(response)
}
