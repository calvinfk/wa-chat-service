package template_usecase

import (
	"context"
	"net/http"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
)

type TemplateUsecase struct {
	phoneNumberUsecase usecase.PhoneNumber
	whatsappService    service.WhatsappBusiness
}

func NewTemplateUsecase(phoneNumberUsecase usecase.PhoneNumber, whatsappService service.WhatsappBusiness) *TemplateUsecase {
	return &TemplateUsecase{
		phoneNumberUsecase: phoneNumberUsecase,
		whatsappService:    whatsappService,
	}
}

func (u *TemplateUsecase) CreateTemplate(ctx context.Context, inputData dto.TemplateCreateRequest) (any, bool, error) {
	whatsappClient, err := u.phoneNumberUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		return nil, false, err
	}
	response, httpCode, err := u.whatsappService.CreateTemplate(whatsappClient, inputData)
	if err != nil {
		return nil, httpCode >= http.StatusInternalServerError, err
	}
	return response, true, nil
}
