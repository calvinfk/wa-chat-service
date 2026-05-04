package http_v1

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net/url"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/handler/http/middleware"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/api_response"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type StorageMediaHandler struct {
	storageMediaUsecase usecase.StorageMedia
	encryptService      service.Encrypt
	zsLog               *zap.SugaredLogger
}

func NewStorageMediaHandler(storageMediaUsecase usecase.StorageMedia, encryptService service.Encrypt, zsLog *zap.SugaredLogger) *StorageMediaHandler {
	return &StorageMediaHandler{
		storageMediaUsecase: storageMediaUsecase,
		encryptService:      encryptService,
		zsLog:               zsLog,
	}
}

func (h *StorageMediaHandler) RegisterRoutes(router fiber.Router) {
	storageMediaRouter := router.Group("/storage-media")
	{
		storageMediaRouter.Post("/upload", middleware.Protected(), h.uploadMedia)
		storageMediaRouter.Get("/get", h.getMedia)
		storageMediaRouter.Post("/encrypt-link", h.encryptMediaLink)
		storageMediaRouter.Delete("/delete", middleware.Protected(), h.deleteMedia)
		storageMediaRouter.Get("/list", middleware.Protected(), h.getMediaList)
	}
}

func (h *StorageMediaHandler) uploadMedia(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaUploadRequest
	if err := ctx.Bind().All(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	if requestData.File == nil {
		code, response := api_response.NewErrorApiResponse(false, fmt.Errorf("file is required"))
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	media, serverError, err := h.storageMediaUsecase.UploadMedia(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Media uploaded successfully", media)
	return ctx.Status(code).JSON(response)
}

func (h *StorageMediaHandler) getMedia(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaGetRequest
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	rangeHeader := ctx.Get(fiber.HeaderRange)
	payload, serverError, err := h.storageMediaUsecase.ParseMediaToken(requestData.Media)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(false, errs.ErrGenericNotFound)
		return ctx.Status(code).JSON(response)
	}
	// Check if payload is a valid UUID (media ID) or a valid URL, and set the appropriate field in the requestData for the use case
	_, err = uuid.Parse(payload)
	if err == nil {
		requestData.StorageMediaID = &payload
	} else {
		// check if valid URL
		if _, err := url.ParseRequestURI(payload); err != nil {
			code, response := api_response.NewErrorApiResponse(false, errs.ErrGenericNotFound)
			return ctx.Status(code).JSON(response)
		}
		requestData.Url = &payload
	}
	mediaResponse, serverError, err := h.storageMediaUsecase.GetMedia(ctx.Context(), requestData, rangeHeader)
	if serverError || err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	ctx.Set(fiber.HeaderContentType, mediaResponse.ContentType)
	ctx.Set(fiber.HeaderContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": mediaResponse.FileName}))
	ctx.Set(fiber.HeaderCacheControl, fmt.Sprintf("private, max-age=%d, immutable", int(mediaResponse.ExpiresIn.Seconds())))
	ctx.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	if mediaResponse.ContentRange != "" {
		ctx.Set(fiber.HeaderContentRange, mediaResponse.ContentRange)
	}
	if mediaResponse.StatusCode != 0 {
		ctx.Status(mediaResponse.StatusCode)
	}
	if mediaResponse.Size <= 0 {
		mediaResponse.Size = -1 // unknown size, let fiber handle it with chunked encoding
	}

	pr := &utils.ProgressReader{
		Ctx:    ctx.Context(),
		Reader: mediaResponse.Reader,
		Size:   mediaResponse.Size,
		Log:    h.zsLog.Infof,
	}

	rawCtx := ctx.RequestCtx()
	rawCtx.Response.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer mediaResponse.Reader.Close()
		buf := make([]byte, 64*1024)
		for {
			if ctx.Context().Err() != nil {
				return
			}
			n, readErr := pr.Read(buf)
			if n > 0 {
				_, writeErr := w.Write(buf[:n])
				if writeErr != nil {
					if !utils.IsClientClosedStreamError(writeErr) {
						h.zsLog.Errorf("[getMedia] Write error: %v", writeErr)
					}
					return
				}
				if flushErr := w.Flush(); flushErr != nil {
					if !utils.IsClientClosedStreamError(flushErr) {
						h.zsLog.Errorf("[getMedia] Flush error: %v", flushErr)
					}
					return
				}
			}
			if readErr == io.EOF {
				return
			}
			if readErr != nil {
				if !utils.IsClientClosedStreamError(readErr) {
					h.zsLog.Errorf("[getMedia] Read error: %v", readErr)
				}
				return
			}
		}
	})
	rawCtx.Response.Header.SetContentLength(int(mediaResponse.Size))
	rawCtx.Response.Header.SetContentTypeBytes([]byte(mediaResponse.ContentType))
	return nil
}

func (h *StorageMediaHandler) deleteMedia(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaDeleteRequest
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	serverError, err := h.storageMediaUsecase.DeleteMedia(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, response := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(response)
	}
	code, response := api_response.NewApiResponse("Media deleted successfully", nil)
	return ctx.Status(code).JSON(response)
}

func (h *StorageMediaHandler) getMediaList(ctx fiber.Ctx) error {
	var requestData filter_request.FilterRequest[dto.StorageMediaGetListRequest]
	if err := ctx.Bind().Query(&requestData.SpecificFilter); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	if err := ctx.Bind().Query(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	authData := ctx.Locals("token_sub").(dto.AuthData)
	response, serverError, err := h.storageMediaUsecase.GetFilteredByTenantID(ctx.Context(), authData.TenantID, requestData)
	if err != nil {
		code, apiResponse := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(apiResponse)
	}
	code, apiResponse := api_response.NewApiResponse("Media list retrieved successfully", response)
	return ctx.Status(code).JSON(apiResponse)

}

func (h *StorageMediaHandler) encryptMediaLink(ctx fiber.Ctx) error {
	var requestData dto.StorageMediaEncryptLinkRequest
	if err := ctx.Bind().Body(&requestData); err != nil {
		code, response := api_response.NewErrorApiResponse(false, err)
		return ctx.Status(code).JSON(response)
	}
	response, serverError, err := h.storageMediaUsecase.GenerateEncryptedLink(ctx.Context(), requestData)
	if err != nil {
		code, apiResponse := api_response.NewErrorApiResponse(serverError, err)
		return ctx.Status(code).JSON(apiResponse)
	}
	code, apiResponse := api_response.NewApiResponse("Encrypted media link generated successfully", response)
	return ctx.Status(code).JSON(apiResponse)
}
