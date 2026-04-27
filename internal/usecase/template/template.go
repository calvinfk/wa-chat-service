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
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TemplateUsecase struct {
	templateRepository                repository.Template
	searchTemplateRepository          repository.SearchTemplate
	whatsappBusinessAccountRepository repository.WaBusinessAccount
	waBusinessAccountUsecase          usecase.WaBusinessAccount
	whatsappService                   service.WhatsappBusiness
	txManager                         *utils.TxManager
	zsLog                             *zap.SugaredLogger
}

func NewTemplateUsecase(templateRepository repository.Template, searchTemplateRepository repository.SearchTemplate, whatsappBusinessAccountRepository repository.WaBusinessAccount, waBusinessAccountUsecase usecase.WaBusinessAccount, whatsappService service.WhatsappBusiness, txManager *utils.TxManager, zsLog *zap.SugaredLogger) *TemplateUsecase {
	return &TemplateUsecase{
		templateRepository:                templateRepository,
		searchTemplateRepository:          searchTemplateRepository,
		whatsappBusinessAccountRepository: whatsappBusinessAccountRepository,
		waBusinessAccountUsecase:          waBusinessAccountUsecase,
		whatsappService:                   whatsappService,
		txManager:                         txManager,
		zsLog:                             zsLog,
	}
}

func (u *TemplateUsecase) CreateTemplate(ctx context.Context, tenantID string, inputData dto.TemplateCreateRequest) (any, bool, error) {
	waba, err := u.whatsappBusinessAccountRepository.GetByID(ctx, inputData.WaBusinessAccountID)
	if err != nil {
		u.zsLog.Errorf("[CreateTemplate] failed to get whatsapp business account: %v", err)
		if err == errs.ErrGenericNotFound {
			return nil, false, fmt.Errorf("whatsapp business account with ID %s not found", inputData.WaBusinessAccountID)
		}
		return nil, true, err
	}
	whatsappClient, _, err := u.waBusinessAccountUsecase.GetWhatsappClient(ctx, tenantID, waba.DocumentID)
	if err != nil {
		u.zsLog.Errorf("[CreateTemplate] failed to get whatsapp client: %v", err)
		return nil, true, err
	}
	response, httpCode, err := whatsappClient.CreateTemplate(whatsapp_business.TemplateCreateRequest{
		Name:            inputData.Name,
		Category:        inputData.Category,
		Language:        inputData.Language,
		ParameterFormat: inputData.ParameterFormat,
		Components:      inputData.Components,
	})
	if err != nil {
		u.zsLog.Errorf("[CreateTemplate] failed to create template: %v", err)
		return nil, httpCode >= http.StatusInternalServerError, err
	}
	componentsString, err := utils.AnyToJsonString(inputData.Components)
	if err != nil {
		u.zsLog.Errorf("[CreateTemplate] failed to marshal template components: %v", err)
		return nil, true, err
	}
	templateID, err := uuid.NewV7()
	if err != nil {
		u.zsLog.Errorf("[CreateTemplate] Failed to generate template ID: %v", err)
		return nil, true, err
	}
	newTemplate := model.Template{
		DocumentID:            templateID.String(),
		WaTemplateID:          response.ID,
		WaBusinessAccountID:   waba.DocumentID,
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
		u.zsLog.Errorf("[CreateTemplate] failed to upsert template: %v", err)
		return nil, true, err
	}
	if err := u.searchTemplateRepository.AddDocuments(ctx, []model.Template{newTemplate}); err != nil {
		u.zsLog.Errorf("[CreateTemplate] failed to add template document to meili repository: %v", err)
	}
	return response, false, nil
}

func (u *TemplateUsecase) GetFiltered(ctx context.Context, tenantID string, inputData filter_request.FilterRequest[dto.TemplateFilterRequest]) (filter_request.FilterResponse[dto.TemplateResponse], bool, error) {
	var emptyResponse filter_request.FilterResponse[dto.TemplateResponse]
	var err error
	var response filter_request.FilterResponse[dto.TemplateResponse]
	data, totalItems, paginate, err := u.searchTemplateRepository.GetFiltered(ctx, inputData)
	if err != nil {
		u.zsLog.Errorf("[GetFilteredByPhoneNumberID] failed to get filtered templates: %v", err)
		return emptyResponse, true, err
	}
	var dataResponse []dto.TemplateResponse
	for _, template := range data {
		dataResponse = append(dataResponse, dto.TemplateResponse{}.FromModel(template))
	}
	response = filter_request.NewFilterResponse(dataResponse, paginate, totalItems)
	return response, false, nil
}

func (u *TemplateUsecase) SyncTemplate(ctx context.Context, tenantID string, inputData dto.TemplateSyncRequest) (bool, error) {
	whatsappClient, waBusinessID, err := u.waBusinessAccountUsecase.GetWhatsappClientByWaBusinessAccountID(ctx, tenantID, inputData.WaBusinessAccountID)
	if err != nil {
		u.zsLog.Errorf("[SyncTemplate] failed to get whatsapp client: %v", err)
		return false, err
	}
	currentTemplates, err := u.templateRepository.GetAll(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[SyncTemplate] failed to get current templates from repository: %v", err)
		return false, err
	}
	currentTemplateMap := make(map[string]model.Template)
	for _, template := range currentTemplates {
		currentTemplateMap[template.WaTemplateID] = template
	}
	var savedTemplateIDs = make(map[string]bool)
	var nextCursor string
	errCount := 0
	processedCount := 0
	for {
		var queryParams string
		if nextCursor != "" {
			queryParams = "after=" + nextCursor
		}
		response, paging, httpCode, err := whatsappClient.GetTemplateList("limit=100", queryParams)
		if err != nil {
			u.zsLog.Errorf("[SyncTemplate] failed to get template list: %v", err)
			return httpCode >= http.StatusInternalServerError, err
		}
		if len(response) == 0 {
			break
		}
		var templates []model.Template
		for _, responseData := range response {
			processedCount++
			templateMeta := responseData.(whatsapp_business.TemplateResponse)
			savedTemplateIDs[templateMeta.ID] = true
			var template model.Template
			currentTemplate, exists := currentTemplateMap[templateMeta.ID]
			if exists {
				template = currentTemplate
			} else {
				componentString, err := utils.AnyToJsonString(templateMeta.Components)
				if err != nil {
					u.zsLog.Errorf("[SyncTemplate] failed to marshal template components for template ID %s: %v", templateMeta.ID, err)
					errCount++
					continue
				}
				templateID, err := uuid.NewV7()
				if err != nil {
					u.zsLog.Errorf("[SyncTemplate] Failed to generate template ID for template ID %s: %v", templateMeta.ID, err)
					errCount++
					continue
				}
				template = model.Template{
					DocumentID:                  templateID.String(),
					WaTemplateID:                templateMeta.ID,
					WaBusinessAccountID:         waBusinessID,
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
				u.zsLog.Errorf("[SyncTemplate] failed to upsert template: %v", err)
				errCount++
				continue
			}
			templates = append(templates, template)
		}
		if err := u.searchTemplateRepository.AddDocuments(ctx, templates); err != nil {
			u.zsLog.Errorf("[SyncTemplate] failed to add template document to meili repository: %v", err)
		}
		if paging.Next == "" {
			break
		} else {
			nextCursor = paging.Cursors.After
		}
	}
	var deletedID []string
	for templateID := range currentTemplateMap {
		if _, exists := savedTemplateIDs[templateID]; !exists {
			processedCount++
			err := u.templateRepository.DeleteByID(ctx, nil, tenantID, templateID)
			if err != nil {
				u.zsLog.Errorf("[SyncTemplate] failed to delete template with ID %s: %v", templateID, err)
				errCount++
				continue
			}
			deletedID = append(deletedID, templateID)
		}
	}
	if len(deletedID) > 0 {
		if err := u.searchTemplateRepository.DeleteDocuments(ctx, deletedID); err != nil {
			u.zsLog.Errorf("[SyncTemplate] failed to delete template document from meili repository: %v", err)
		}
	}
	if errCount > 0 {
		return false, fmt.Errorf("processed %d templates with %d errors", processedCount, errCount)
	}
	return false, nil
}

func (u *TemplateUsecase) DeleteTemplate(ctx context.Context, tenantID string, inputData dto.TemplateDeleteRequest) (bool, error) {
	whatsappClient, _, err := u.waBusinessAccountUsecase.GetWhatsappClientByWaBusinessAccountID(ctx, tenantID, inputData.WaBusinessAccountID)
	if err != nil {
		u.zsLog.Errorf("[DeleteTemplate] failed to get whatsapp client: %v", err)
		return false, err
	}
	template, err := u.templateRepository.GetByID(ctx, tenantID, inputData.ID)
	if err != nil {
		u.zsLog.Errorf("[DeleteTemplate] failed to get template from repository: %v", err)
		return false, err
	}
	_, httpCode, err := whatsappClient.DeleteTemplate(template.WaTemplateID, template.Name)
	if err != nil {
		u.zsLog.Errorf("[DeleteTemplate] failed to delete template: %v", err)
		return httpCode >= http.StatusInternalServerError, err
	}
	err = u.templateRepository.DeleteByID(ctx, nil, tenantID, template.DocumentID)
	if err != nil {
		u.zsLog.Errorf("[DeleteTemplate] failed to delete template from repository: %v", err)
		return true, err
	}
	err = u.searchTemplateRepository.DeleteDocuments(ctx, []string{template.DocumentID})
	if err != nil {
		u.zsLog.Errorf("[DeleteTemplate] failed to delete template document from meili repository: %v", err)
		return true, err
	}
	return false, nil
}

func (u *TemplateUsecase) UpdateTemplate(ctx context.Context, tenantID string, inputData dto.TemplateUpdateRequest) (bool, error) {
	whatsappClient, _, err := u.waBusinessAccountUsecase.GetWhatsappClientByWaBusinessAccountID(ctx, tenantID, inputData.WaBusinessAccountID)
	if err != nil {
		u.zsLog.Errorf("[UpdateTemplate] failed to get whatsapp client: %v", err)
		return true, err
	}
	currentTemplate, err := u.templateRepository.GetByID(ctx, tenantID, inputData.ID)
	if err != nil {
		u.zsLog.Errorf("[UpdateTemplate] failed to get current template from repository: %v", err)
		return true, err
	}
	if strings.ToUpper(currentTemplate.Status) != "REJECTED" {
		u.zsLog.Errorf("[UpdateTemplate] cannot update template with status %s", currentTemplate.Status)
		return false, fmt.Errorf("cannot update template with status %s", currentTemplate.Status)
	}
	componentsString, err := utils.AnyToJsonString(inputData.Components)
	if err != nil {
		u.zsLog.Errorf("[UpdateTemplate] failed to marshal template components: %v", err)
		return true, err
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
		u.zsLog.Errorf("[UpdateTemplate] failed to update template: %v", err)
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
		u.zsLog.Errorf("[UpdateTemplate] failed to upsert template: %v", err)
		return true, err
	}
	err = u.searchTemplateRepository.AddDocuments(ctx, []model.Template{currentTemplate})
	if err != nil {
		u.zsLog.Errorf("[UpdateTemplate] failed to add template document to meili repository: %v", err)
		return true, err
	}
	return false, nil
}
