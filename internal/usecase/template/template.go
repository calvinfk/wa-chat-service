package template_usecase

import (
	"context"
	"log"
	"net/http"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/formatter"
)

type TemplateUsecase struct {
	templateRepository repository.Template
	tenantUsecase      usecase.Tenant
	whatsappService    service.WhatsappBusiness
}

func NewTemplateUsecase(templateRepository repository.Template, tenantUsecase usecase.Tenant, whatsappService service.WhatsappBusiness) *TemplateUsecase {
	return &TemplateUsecase{
		templateRepository: templateRepository,
		tenantUsecase:      tenantUsecase,
		whatsappService:    whatsappService,
	}
}

func (u *TemplateUsecase) CreateTemplate(ctx context.Context, inputData dto.TemplateCreateRequest) (any, bool, error) {
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][CreateTemplate] failed to get whatsapp client:", err)
		return nil, true, err
	}
	// log.Print(tenantID)
	response, httpCode, err := u.whatsappService.CreateTemplate(whatsappClient, inputData)
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][CreateTemplate] failed to create template:", err)
		return nil, httpCode >= http.StatusInternalServerError, err
	}
	componentsString, err := formatter.AnyToJsonString(inputData.Components)
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][CreateTemplate] failed to marshal template components:", err)
		return nil, true, err
	}
	newTemplate := model.Template{
		DocumentID:            response.ID,
		Name:                  inputData.Name,
		Category:              response.Category,
		Language:              inputData.Language,
		MessageSendTTLSeconds: 0,
		ParameterFormat:       inputData.ParameterFormat,
		Status:                response.Status,
		Components:            componentsString,
		CreatedAt:             time.Now(),
	}

	if _, err := u.templateRepository.Upsert(ctx, nil, tenantID, newTemplate); err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][CreateTemplate] failed to upsert template:", err)
		return nil, true, err
	}
	return response, false, nil
}

func (u *TemplateUsecase) GetFilteredByPhoneNumberID(ctx context.Context, inputData filter_request.FilterRequest[dto.TemplateGetByPhoneNumberID]) (filter_request.FilterResponse[dto.TemplateGetByPhoneNumberIDResponse], bool, error) {
	var emptyResponse filter_request.FilterResponse[dto.TemplateGetByPhoneNumberIDResponse]
	var err error
	_, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.SpecificFilter.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][GetFilteredByPhoneNumberID] failed to get whatsapp client:", err)
		return emptyResponse, true, err
	}
	var response filter_request.FilterResponse[dto.TemplateGetByPhoneNumberIDResponse]
	response, err = u.templateRepository.GetFilteredByPhoneNumberID(ctx, tenantID, inputData)
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][GetFilteredByPhoneNumberID] failed to get templates by phone number id:", err)
		return emptyResponse, true, err
	}
	return response, false, nil
}

func (u *TemplateUsecase) SyncTemplate(ctx context.Context, inputData dto.TemplateSyncRequest) (bool, error) {
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][SyncTemplate] failed to get whatsapp client:", err)
		return false, err
	}
	response, httpCode, err := whatsappClient.GetTemplateList()
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][SyncTemplate] failed to get template list:", err)
		return httpCode >= http.StatusInternalServerError, err
	}
	for _, template := range response {
		componentsString, err := formatter.AnyToJsonString(template.Components)
		if err != nil {
			log.Println("[ERROR][internal/usecase/template/template.go][SyncTemplate] failed to marshal template components:", err)
			continue
		}
		newTemplate := model.Template{
			DocumentID:                  template.ID,
			Name:                        template.Name,
			Category:                    template.Category,
			IsPrimaryDeviceDeliveryOnly: template.IsPrimaryDeviceDeliveryOnly,
			Language:                    template.Language,
			MessageSendTTLSeconds:       template.MessageSendTTLSeconds,
			ParameterFormat:             template.ParameterFormat,
			Status:                      template.Status,
			Components:                  componentsString,
			CreatedAt:                   time.Now(),
		}
		if _, err := u.templateRepository.Upsert(ctx, nil, tenantID, newTemplate); err != nil {
			log.Println("[ERROR][internal/usecase/template/template.go][SyncTemplate] failed to upsert template:", err)
			continue
		}
	}
	return false, nil
}

func (u *TemplateUsecase) GetTemplatesMeta(ctx context.Context, inputData dto.TemplateGetByPhoneNumberID) (any, bool, error) {
	whatsappClient, _, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][GetTemplatesMeta] failed to get whatsapp client:", err)
		return nil, true, err
	}
	metaResponse, httpCode, err := whatsappClient.GetTemplateList()
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][GetTemplatesMeta] failed to get templates meta:", err)
		return nil, httpCode >= http.StatusInternalServerError, err
	}
	return metaResponse, false, nil
}

func (u *TemplateUsecase) DeleteTemplate(ctx context.Context, inputData dto.TemplateDeleteRequest) (bool, error) {
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][DeleteTemplate] failed to get whatsapp client:", err)
		return false, err
	}
	_, httpCode, err := whatsappClient.DeleteTemplate(inputData.ID, inputData.Name)
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][DeleteTemplate] failed to delete template:", err)
		return httpCode >= http.StatusInternalServerError, err
	}
	if inputData.ID != "" {
		err = u.templateRepository.DeleteByID(ctx, nil, tenantID, inputData.ID)
		if err != nil {
			log.Println("[ERROR][internal/usecase/template/template.go][DeleteTemplate] failed to delete template from repository:", err)
			return true, err
		}
	} else if inputData.Name != "" {
		err = u.templateRepository.DeleteByName(ctx, nil, tenantID, inputData.Name)
		if err != nil {
			log.Println("[ERROR][internal/usecase/template/template.go][DeleteTemplate] failed to delete template from repository:", err)
			return true, err
		}
	}
	return false, nil
}
