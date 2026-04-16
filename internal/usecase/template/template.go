package template_usecase

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/utils"

	"go.uber.org/zap"
)

type TemplateUsecase struct {
	templateRepository repository.Template
	tenantUsecase      usecase.Tenant
	whatsappService    service.WhatsappBusiness
	txManager          *utils.TxManager
	zslog              *zap.SugaredLogger
}

func NewTemplateUsecase(templateRepository repository.Template, tenantUsecase usecase.Tenant, whatsappService service.WhatsappBusiness, txManager *utils.TxManager, zslog *zap.SugaredLogger) *TemplateUsecase {
	return &TemplateUsecase{
		templateRepository: templateRepository,
		tenantUsecase:      tenantUsecase,
		whatsappService:    whatsappService,
		txManager:          txManager,
		zslog:              zslog,
	}
}

func (u *TemplateUsecase) CreateTemplate(ctx context.Context, inputData dto.TemplateCreateRequest) (any, bool, error) {
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[CreateTemplate] failed to get whatsapp client: %v", err)
		return nil, true, err
	}
	// log.Print(tenantID)
	response, httpCode, err := u.whatsappService.CreateTemplate(whatsappClient, inputData)
	if err != nil {
		u.zslog.Errorf("[CreateTemplate] failed to create template: %v", err)
		return nil, httpCode >= http.StatusInternalServerError, err
	}
	componentsString, err := utils.AnyToJsonString(inputData.Components)
	if err != nil {
		u.zslog.Errorf("[CreateTemplate] failed to marshal template components: %v", err)
		return nil, true, err
	}
	newTemplate := model.Template{
		DocumentID:            response.ID,
		TenantID:              tenantID,
		Name:                  inputData.Name,
		Category:              response.Category,
		Language:              inputData.Language,
		MessageSendTTLSeconds: 0,
		ParameterFormat:       inputData.ParameterFormat,
		Status:                response.Status,
		Components:            componentsString,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if _, err := u.templateRepository.Upsert(ctx, nil, newTemplate); err != nil {
		u.zslog.Errorf("[CreateTemplate] failed to upsert template: %v", err)
		return nil, true, err
	}
	return response, false, nil
}

func (u *TemplateUsecase) GetFilteredByPhoneNumberID(ctx context.Context, inputData filter_request.FilterRequest[dto.TemplateGetByPhoneNumberID]) (filter_request.FilterResponse[dto.TemplateGetByPhoneNumberIDResponse], bool, error) {
	var emptyResponse filter_request.FilterResponse[dto.TemplateGetByPhoneNumberIDResponse]
	var err error
	var response filter_request.FilterResponse[dto.TemplateGetByPhoneNumberIDResponse]
	response, err = u.templateRepository.GetFilteredByPhoneNumberID(ctx, inputData.SpecificFilter.TenantID, inputData)
	if err != nil {
		u.zslog.Errorf("[GetFilteredByPhoneNumberID] failed to get templates by phone number id: %v", err)
		return emptyResponse, true, err
	}
	return response, false, nil
}

func (u *TemplateUsecase) SyncTemplate(ctx context.Context, inputData dto.TemplateSyncRequest) (bool, error) {
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[SyncTemplate] failed to get whatsapp client: %v", err)
		return false, err
	}
	currentTemplates, err := u.templateRepository.GetAll(ctx, tenantID)
	if err != nil {
		u.zslog.Errorf("[SyncTemplate] failed to get current templates from repository: %v", err)
		return false, err
	}
	currentTemplateMap := make(map[string]model.Template)
	for _, template := range currentTemplates {
		currentTemplateMap[template.DocumentID] = template
	}
	var savedTemplateIDs = make(map[string]bool)
	var nextCursor string
	for {
		var queryParams string
		if nextCursor != "" {
			queryParams = "after=" + nextCursor
		}
		response, paging, httpCode, err := whatsappClient.GetTemplateList("limit=100", queryParams)
		if err != nil {
			u.zslog.Errorf("[SyncTemplate] failed to get template list: %v", err)
			return httpCode >= http.StatusInternalServerError, err
		}
		if len(response) == 0 {
			break
		}
		for _, responseData := range response {
			templateMeta := responseData.(whatsapp_business.TemplateResponse)
			savedTemplateIDs[templateMeta.ID] = true
			var template model.Template
			currentTemplate, exists := currentTemplateMap[templateMeta.ID]
			if exists {
				template = currentTemplate
			} else {
				componentString, err := utils.AnyToJsonString(templateMeta.Components)
				if err != nil {
					u.zslog.Errorf("[SyncTemplate] failed to marshal template components for template ID %s: %v", templateMeta.ID, err)
					continue
				}
				template = model.Template{
					DocumentID:                  templateMeta.ID,
					TenantID:                    tenantID,
					Name:                        templateMeta.Name,
					Category:                    templateMeta.Category,
					IsPrimaryDeviceDeliveryOnly: templateMeta.IsPrimaryDeviceDeliveryOnly,
					Language:                    templateMeta.Language,
					MessageSendTTLSeconds:       templateMeta.MessageSendTTLSeconds,
					ParameterFormat:             templateMeta.ParameterFormat,
					Status:                      templateMeta.Status,
					Components:                  componentString,
					CreatedAt:                   time.Now(),
					UpdatedAt:                   time.Now(),
				}
			}
			template.Status = templateMeta.Status
			template.Category = templateMeta.Category
			template.UpdatedAt = time.Now()
			if _, err := u.templateRepository.Upsert(ctx, nil, template); err != nil {
				u.zslog.Errorf("[SyncTemplate] failed to upsert template: %v", err)
			}
		}
		if paging.Next == "" {
			break
		} else {
			nextCursor = paging.Cursors.After
		}
	}
	for templateID := range currentTemplateMap {
		if _, exists := savedTemplateIDs[templateID]; !exists {
			err := u.templateRepository.DeleteByID(ctx, nil, tenantID, templateID)
			if err != nil {
				u.zslog.Errorf("[SyncTemplate] failed to delete template with ID %s: %v", templateID, err)
			}
		}
	}
	return false, nil
}

func (u *TemplateUsecase) DeleteTemplate(ctx context.Context, inputData dto.TemplateDeleteRequest) (bool, error) {
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[DeleteTemplate] failed to get whatsapp client: %v", err)
		return false, err
	}
	_, httpCode, err := whatsappClient.DeleteTemplate(inputData.ID, inputData.Name)
	if err != nil {
		u.zslog.Errorf("[DeleteTemplate] failed to delete template: %v", err)
		return httpCode >= http.StatusInternalServerError, err
	}
	if inputData.ID != "" {
		err = u.templateRepository.DeleteByID(ctx, nil, tenantID, inputData.ID)
		if err != nil {
			u.zslog.Errorf("[DeleteTemplate] failed to delete template from repository: %v", err)
			return true, err
		}
	} else if inputData.Name != "" {
		err = u.templateRepository.DeleteByName(ctx, nil, tenantID, inputData.Name)
		if err != nil {
			u.zslog.Errorf("[DeleteTemplate] failed to delete template from repository: %v", err)
			return true, err
		}
	}
	return false, nil
}

func (u *TemplateUsecase) UpdateTemplate(ctx context.Context, inputData dto.TemplateUpdateRequest) (bool, error) {
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[UpdateTemplate] failed to get whatsapp client: %v", err)
		return false, err
	}
	currentTemplate, err := u.templateRepository.GetByID(ctx, tenantID, inputData.ID)
	if err != nil {
		u.zslog.Errorf("[UpdateTemplate] failed to get current template from repository: %v", err)
		return false, err
	}
	if strings.ToUpper(currentTemplate.Status) != "REJECTED" {
		u.zslog.Errorf("[UpdateTemplate] cannot update template with status %s", currentTemplate.Status)
		return false, fmt.Errorf("cannot update template with status %s", currentTemplate.Status)
	}
	componentsString, err := utils.AnyToJsonString(inputData.Components)
	if err != nil {
		u.zslog.Errorf("[UpdateTemplate] failed to marshal template components: %v", err)
		return false, err
	}
	payload := whatsapp_business.TemplateCreateRequest{
		Name:            inputData.Name,
		Language:        inputData.Language,
		Category:        inputData.Category,
		ParameterFormat: inputData.ParameterFormat,
		Components:      inputData.Components,
	}
	response, httpCode, err := whatsappClient.UpdateTemplate(inputData.ID, payload)
	if err != nil {
		u.zslog.Errorf("[UpdateTemplate] failed to update template: %v", err)
		return httpCode >= http.StatusInternalServerError, err
	}
	currentTemplate.Category = response.Category
	currentTemplate.Status = response.Status
	currentTemplate.Name = payload.Name
	currentTemplate.Language = payload.Language
	currentTemplate.ParameterFormat = payload.ParameterFormat
	currentTemplate.Components = componentsString
	currentTemplate.UpdatedAt = time.Now()
	_, err = u.templateRepository.Upsert(ctx, nil, currentTemplate)
	if err != nil {
		u.zslog.Errorf("[UpdateTemplate] failed to upsert template: %v", err)
		return false, err
	}
	return false, nil
}
