package http_v1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"strconv"
	"strings"
	"time"
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
		storageMediaRouter.Get("/get", h.getMedia)
		storageMediaRouter.Delete("/delete", middleware.Protected(), h.deleteMedia)
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

type progressReader struct {
	r       io.Reader
	size    int64
	read    int64
	lastLog time.Time
	log     func(string, ...any)
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.read += int64(n)
		if time.Since(p.lastLog) >= time.Second {
			if p.size > 0 {
				p.log("[getMedia] stream progress: %d/%d bytes (%.1f%%)", p.read, p.size, float64(p.read)*100/float64(p.size))
			} else {
				p.log("[getMedia] stream progress: %d bytes", p.read)
			}
			p.lastLog = time.Now()
		}
	}
	return n, err
}

func isClientClosedStreamError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, io.ErrClosedPipe) {
		return true
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "response body closed") ||
		strings.Contains(errText, "stream closed") ||
		strings.Contains(errText, "broken pipe") ||
		strings.Contains(errText, "connection reset by peer")
}

func (h *StorageMediaHandler) getMedia(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaGetRequest
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewApiResponse(false, err, "", nil)
		return ctx.Status(code).JSON(response)
	}
	rangeHeader := ctx.Get(fiber.HeaderRange)
	mediaResponse, serverError, err := h.storageMediaUsecase.GetMedia(ctx.Context(), requestData, rangeHeader)
	if serverError || err != nil {
		code, response := api_response.NewApiResponse(serverError, err, "Failed to retrieve media", nil)
		return ctx.Status(code).JSON(response)
	}
	defer mediaResponse.Reader.Close()

	ctx.Set(fiber.HeaderContentType, mediaResponse.ContentType)
	ctx.Set(fiber.HeaderContentDisposition, mime.FormatMediaType("inline", map[string]string{"filename": mediaResponse.FileName}))
	ctx.Set(fiber.HeaderCacheControl, "private, max-age="+fmt.Sprintf("%d", int(mediaResponse.ExpiresIn.Seconds())))
	ctx.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	ctx.Set("Accept-Ranges", "bytes")
	if mediaResponse.ContentRange != "" {
		ctx.Set("Content-Range", mediaResponse.ContentRange)
	}
	if mediaResponse.Size > 0 {
		ctx.Set(fiber.HeaderContentLength, strconv.FormatInt(mediaResponse.Size, 10))
	}
	if mediaResponse.StatusCode != 0 {
		ctx.Status(mediaResponse.StatusCode)
	}

	// pr := &progressReader{
	// 	r:       mediaResponse.Reader,
	// 	size:    mediaResponse.Size,
	// 	lastLog: time.Now(),
	// 	log:     h.zslog.Infof,
	// }
	// _, err = io.Copy(ctx.Response().BodyWriter(), pr)

	_, err = io.Copy(ctx.Response().BodyWriter(), mediaResponse.Reader)
	if err != nil {
		if isClientClosedStreamError(err) || ctx.Context().Err() != nil {
			h.zslog.Infof("[getMedia] Client disconnected during stream")
			return nil
		}
		h.zslog.Error("[getMedia] Failed to stream file:", err)
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
