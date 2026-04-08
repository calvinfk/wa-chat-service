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
		storageMediaRouter.Post("/upload", h.uploadMedia)
		storageMediaRouter.Get("/get", h.getMedia)
		storageMediaRouter.Delete("/delete", h.deleteMedia)
		storageMediaRouter.Post("/save-media-id", h.saveMediaID)
	}
}

func (h *StorageMediaHandler) uploadMedia(ctx fiber.Ctx) error {
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
		log.Println("[ERROR][internal/handler/http/v1/storage_media_handler.go][getMedia] Failed to stream file to response:", err)
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
		log.Println("[ERROR][internal/handler/http/v1/storage_media_handler.go][saveMediaID] Failed to bind request data:", err)
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	data, serverError, err := h.storageMediaUsecase.SaveMediaID(ctx.Context(), requestData)
	code, response := api_response.NewApiResponse(serverError, err, "Media uploaded successfully", data)
	return ctx.Status(code).JSON(response)
}
