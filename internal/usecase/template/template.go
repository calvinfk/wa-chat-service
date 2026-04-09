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
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/transaction"
)

type TemplateUsecase struct {
	templateRepository repository.Template
	tenantUsecase      usecase.Tenant
	whatsappService    service.WhatsappBusiness
	txManager          *transaction.TxManager
}

func NewTemplateUsecase(templateRepository repository.Template, tenantUsecase usecase.Tenant, whatsappService service.WhatsappBusiness, txManager *transaction.TxManager) *TemplateUsecase {
	return &TemplateUsecase{
		templateRepository: templateRepository,
		tenantUsecase:      tenantUsecase,
		whatsappService:    whatsappService,
		txManager:          txManager,
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
	currentTemplates, err := u.templateRepository.GetByTenantID(ctx, tenantID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/template/template.go][SyncTemplate] failed to get current templates from repository:", err)
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
			log.Println("[ERROR][internal/usecase/template/template.go][SyncTemplate] failed to get template list:", err)
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
				componentString, err := formatter.AnyToJsonString(templateMeta.Components)
				if err != nil {
					log.Printf("[ERROR][internal/usecase/template/template.go][SyncTemplate] failed to marshal template components for template ID %s: %v", templateMeta.ID, err)
					continue
				}
				template = model.Template{
					DocumentID:                  templateMeta.ID,
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
			if _, err := u.templateRepository.Upsert(ctx, nil, tenantID, template); err != nil {
				log.Println("[ERROR][internal/usecase/template/template.go][SyncTemplate] failed to upsert template:", err)
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
				log.Printf("[ERROR][internal/usecase/template/template.go][SyncTemplate] failed to delete template with ID %s: %v", templateID, err)
			}
		}
	}
	return false, nil
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
