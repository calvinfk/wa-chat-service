package broadcast_usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BroadcastUsecase struct {
	templateRepository       repository.Template
	broadcastRepository      repository.Broadcast
	tenantRepository         repository.Tenant
	messageUsecase           usecase.Message
	waBusinessAccountUsecase usecase.WaBusinessAccount
	googleTaskService        service.GoogleTask
	whatsappService          service.WhatsappBusiness
	txManager                *utils.TxManager
	zsLog                    *zap.SugaredLogger
}

func NewBroadcastUsecase(
	templateRepository repository.Template,
	broadcastRepository repository.Broadcast,
	tenantRepository repository.Tenant,
	messageUsecase usecase.Message,
	waBusinessAccountUsecase usecase.WaBusinessAccount,
	googleTaskService service.GoogleTask,
	whatsappService service.WhatsappBusiness,
	txManager *utils.TxManager,
	zsLog *zap.SugaredLogger,
) *BroadcastUsecase {
	return &BroadcastUsecase{
		templateRepository:       templateRepository,
		broadcastRepository:      broadcastRepository,
		tenantRepository:         tenantRepository,
		messageUsecase:           messageUsecase,
		waBusinessAccountUsecase: waBusinessAccountUsecase,
		googleTaskService:        googleTaskService,
		whatsappService:          whatsappService,
		txManager:                txManager,
		zsLog:                    zsLog,
	}
}

func (u *BroadcastUsecase) ScheduleBroadcast(ctx context.Context, tenantID string, inputData dto.BroadcastScheduleRequest) (bool, error) {
	broadcast, err := u.broadcastRepository.GetByID(ctx, inputData.ID)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to get broadcast by ID: %v", err)
		return true, err
	}
	if broadcast.Status != string(dto.BroadcastScheduleDraft) {
		u.zsLog.Errorf("[ScheduleBroadcast] broadcast with ID %s is not in draft status, cannot schedule", inputData.ID)
		return false, nil
	}
	if broadcast.TenantID != tenantID {
		u.zsLog.Errorf("[ScheduleBroadcast] broadcast with ID %s does not belong to tenant with ID %s, cannot schedule", inputData.ID, tenantID)
		return false, errs.ErrGenericForbidden
	}
	template, err := u.templateRepository.GetByID(ctx, broadcast.TenantID, broadcast.TemplateID)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to get template: %v", err)
		return true, err
	}
	serverError, err := u.txManager.DoFirestore(ctx, func(ctx context.Context, tx *firestore.Transaction) (bool, error) {
		broadcast.Status = string(dto.BroadcastScheduleScheduled)
		if inputData.SendAt != nil {
			broadcast.SendAt = *inputData.SendAt
		} else if broadcast.SendAt.IsZero() || broadcast.SendAt.Before(time.Now()) {
			broadcast.SendAt = time.Now().Add(time.Second * 10) // default to 10 seconds later if send_at is not provided or send_at is in the past
		}
		err = u.broadcastRepository.Update(ctx, tx, broadcast)
		if err != nil {
			u.zsLog.Errorf("[ScheduleBroadcast] failed to update broadcast status: %v", err)
			return true, err
		}
		serverError, err := u.createScheduleBroadcastTask(ctx, tx, broadcast, template)
		if err != nil {
			return serverError, err
		}
		return false, nil
	})
	return serverError, err
}

func (u *BroadcastUsecase) createScheduleBroadcastTask(ctx context.Context, tx *firestore.Transaction, broadcast model.Broadcast, template model.Template) (bool, error) {
	var broadcastPayload map[string]any
	err := json.Unmarshal([]byte(broadcast.Payload), &broadcastPayload)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to unmarshal template components: %v", err)
		return true, err
	}
	templateSend, ok := broadcastPayload["template"].(map[string]any)
	if !ok {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to parse template payload: %v", err)
		return true, err
	}
	var templateSendComponents []map[string]any
	if components, ok := templateSend["components"].([]any); ok {
		for _, component := range components {
			if componentMap, ok := component.(map[string]any); ok {
				templateSendComponents = append(templateSendComponents, componentMap)
			}
		}
	}
	quickReplyPayload, err := u.injectQuickReplyPayload(broadcast.DocumentID, template)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to inject quick reply payload: %v", err)
		return true, err
	}
	for _, payload := range quickReplyPayload {
		templateSendComponents = append(templateSendComponents, payload)
	}
	templateSend["components"] = templateSendComponents
	broadcastPayload["template"] = templateSend
	payloadBytes, err := json.Marshal(broadcastPayload)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to marshal broadcast payload with quick reply: %v", err)
		return true, err
	}
	broadcast.Payload = string(payloadBytes)
	err = u.broadcastRepository.Update(ctx, tx, broadcast)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to update broadcast with quick reply payload: %v", err)
		return true, err
	}

	phoneNumbers, err := u.tenantRepository.GetContactByPhoneNumbers(ctx, broadcast.TenantID, broadcast.RecipientIds)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to get contacts by phone numbers: %v", err)
		return true, err
	}
	for _, recipient := range broadcast.RecipientIds {
		var recipientName string
		contact, exists := phoneNumbers[recipient]
		if exists {
			recipientName = contact["name"]
		} else {
			recipientName = recipient
		}
		broadcastRecipient := model.BroadcastRecipient{
			DocumentID:    uuid.NewString(),
			BroadcastID:   broadcast.DocumentID,
			RecipientId:   recipient,
			RecipientName: recipientName,
			RecipientType: "individual", // TODO: support group recipient type
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		err = u.broadcastRepository.InsertRecipient(ctx, tx, broadcastRecipient)
		if err != nil {
			u.zsLog.Errorf("[ScheduleBroadcast] failed to insert broadcast recipient: %v", err)
			return true, err
		}
	}

	err = u.googleTaskService.CreateBroadcastTask(broadcast.DocumentID, broadcast.SendAt)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to create broadcast task: %v", err)
		return true, err
	}
	return false, nil
}

func (u *BroadcastUsecase) UpsertBroadcast(ctx context.Context, tenantID string, inputData dto.BroadcastUpsertRequest) (dto.BroadcastResponse, bool, error) {
	var err error
	var broadcast model.Broadcast
	var emptyResponse dto.BroadcastResponse
	if inputData.ID != nil {
		broadcast, err = u.broadcastRepository.GetByID(ctx, *inputData.ID)
		if err != nil {
			u.zsLog.Errorf("[ScheduleBroadcast] failed to get broadcast by ID: %v", err)
			return emptyResponse, true, err
		}
		if broadcast.Status != string(dto.BroadcastScheduleDraft) {
			u.zsLog.Errorf("[ScheduleBroadcast] broadcast with ID %s is not in draft status, cannot update", *inputData.ID)
			return emptyResponse, false, fmt.Errorf("broadcast currently in %s status, only broadcast in draft status can be updated", broadcast.Status)
		}
		if broadcast.TenantID != tenantID {
			u.zsLog.Errorf("[ScheduleBroadcast] broadcast with ID %s does not belong to tenant %s", *inputData.ID, tenantID)
			return emptyResponse, false, errs.ErrGenericForbidden
		}
	}
	// remove duplicate recipient IDs
	recipientIDMap := make(map[string]bool)
	var uniqueRecipientIds []string
	for _, recipientID := range inputData.Recipients {
		if _, exists := recipientIDMap[recipientID]; !exists {
			recipientIDMap[recipientID] = true
			uniqueRecipientIds = append(uniqueRecipientIds, recipientID)
		}
	}
	inputData.Recipients = uniqueRecipientIds
	whatsappClient, wabaID, err := u.waBusinessAccountUsecase.GetWhatsappClient(ctx, tenantID, inputData.PhoneNumberId)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to get whatsapp client: %v", err)
		return emptyResponse, true, err
	}
	template, err := u.templateRepository.GetByID(ctx, wabaID, inputData.TemplateID)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to get template: %v", err)
		if err == errs.ErrGenericNotFound {
			return emptyResponse, false, errs.ErrGenericNotFound
		}
		return emptyResponse, true, err
	}
	broadcastID, err := uuid.NewV7()
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to generate broadcast ID: %v", err)
		return emptyResponse, true, err
	}
	payload := make(map[string]any)
	payload["template"] = map[string]any{
		"name":       template.Name,
		"language":   map[string]any{"code": template.Language},
		"components": inputData.Components,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to marshal template payload: %v", err)
		return emptyResponse, true, err
	}
	templateComponent, err := whatsapp_business.NewTemplateComponent(payloadBytes)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to create template component: %v", err)
		return emptyResponse, true, err
	}
	err = u.whatsappService.ValidateTemplatePayload(whatsappClient, template, templateComponent)
	if err != nil {
		return emptyResponse, false, err
	}
	payloadString, err := json.Marshal(templateComponent.GetPayload())
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to marshal template payload: %v", err)
		return emptyResponse, true, err
	}
	var sendingTime time.Time
	if inputData.SendAt == nil {
		sendingTime = time.Now().Add(time.Second * 10) // default to 10 seconds later if send_at is not provided
	} else {
		sendingTime = *inputData.SendAt
	}
	serverError, err := u.txManager.DoFirestore(ctx, func(ctx context.Context, tx *firestore.Transaction) (bool, error) {
		if inputData.ID == nil {
			broadcast = model.Broadcast{
				DocumentID:      broadcastID.String(),
				TenantID:        tenantID,
				ParameterFormat: template.ParameterFormat,
				CreatedBy:       inputData.PhoneNumberId,
				CreatedAt:       time.Now(),
			}
		}
		broadcast.Name = inputData.Name
		broadcast.TemplateID = inputData.TemplateID
		broadcast.RecipientIds = inputData.Recipients
		broadcast.PhoneNumberId = inputData.PhoneNumberId
		broadcast.Payload = string(payloadString)
		broadcast.Status = inputData.Status
		broadcast.SendAt = sendingTime
		broadcast.UpdatedAt = time.Now()
		err = u.broadcastRepository.Insert(ctx, tx, broadcast)
		if err != nil {
			u.zsLog.Errorf("[ScheduleBroadcast] failed to insert broadcast: %v", err)
			return true, err
		}
		if broadcast.Status == string(dto.BroadcastScheduleScheduled) {
			serverError, err := u.createScheduleBroadcastTask(ctx, tx, broadcast, template)
			if err != nil {
				u.zsLog.Errorf("[ScheduleBroadcast] failed to schedule broadcast: %v", err)
				broadcast.Status = string(dto.BroadcastScheduleDraft)
				return serverError, err
			}
		}
		return false, nil
	})
	if err != nil {
		return emptyResponse, serverError, err
	}
	return dto.BroadcastResponse{}.FromModel(broadcast), false, err
}

func (u *BroadcastUsecase) injectQuickReplyPayload(broadcastID string, template model.Template) ([]map[string]any, error) {
	var templateDBComponents []map[string]any
	err := json.Unmarshal([]byte(template.Components), &templateDBComponents)
	if err != nil {
		u.zsLog.Errorf("[ScheduleBroadcast] failed to get template components: %v", err)
		return nil, err
	}
	// broadcast_{{broadcast-id}}_{{contact-phone}}_{{text}}
	var quickReplyPayload []map[string]any
	for _, component := range templateDBComponents {
		if component["type"] == "BUTTONS" {
			for i, buttonComponent := range component["buttons"].([]any) {
				button, err := whatsapp_business.NewTemplateCreateButton(buttonComponent)
				if err != nil {
					u.zsLog.Errorf("[injectQuickReplyPayload] failed to create button component: %v", err)
					return nil, err
				}
				iStr := strconv.Itoa(i)
				if button.GetType() == "QUICK_REPLY" {
					quickReplyPayload = append(quickReplyPayload, map[string]any{
						"type":     "button",
						"sub_type": "QUICK_REPLY",
						"index":    iStr,
						"parameters": []map[string]any{{
							"type":    "payload",
							"payload": "broadcast_" + broadcastID + "_" + iStr + "_" + button.GetText(),
						}},
					})
				}
			}
		}
	}

	return quickReplyPayload, nil

}

func (u *BroadcastUsecase) SendBroadcast(ctx context.Context, broadcastID string) (bool, error) {
	broadcast, err := u.broadcastRepository.GetByID(ctx, broadcastID)
	if err != nil {
		u.zsLog.Errorf("[SendBroadcast] failed to get broadcast by ID: %v", err)
		return true, err
	}
	var finalStatus dto.BroadcastScheduleStatus
	defer func() {
		if broadcast.Status == string(dto.BroadcastScheduleSuccess) {
			return
		}
		broadcast.Status = string(finalStatus)
		broadcast.UpdatedAt = time.Now()
		err = u.broadcastRepository.Update(ctx, nil, broadcast)
		if err != nil {
			u.zsLog.Errorf("[SendBroadcast] failed to update broadcast status: %v", err)
		}
	}()
	whatsappClient, _, err := u.waBusinessAccountUsecase.GetWhatsappClient(ctx, broadcast.TenantID, broadcast.PhoneNumberId)
	if err != nil {
		u.zsLog.Errorf("[SendBroadcast] failed to get whatsapp client: %v", err)
		return true, err
	}
	broadcastRecipients, err := u.broadcastRepository.GetRecipientsByBroadcastID(ctx, broadcastID)
	if err != nil {
		u.zsLog.Errorf("[SendBroadcast] failed to list broadcast recipients: %v", err)
		return true, err
	}
	// get tenant's template fields to validate the payload before sending messages
	payload := make(map[string]any)
	err = json.Unmarshal([]byte(broadcast.Payload), &payload)
	if err != nil {
		finalStatus = dto.BroadcastScheduleFailed
		u.zsLog.Errorf("[SendBroadcast] failed to unmarshal broadcast payload: %v", err)
		return true, err
	}
	sendComponents := make([]map[string]any, 0)
	if components, ok := payload["template"].(map[string]any)["components"].([]any); ok {
		for _, component := range components {
			if componentMap, ok := component.(map[string]any); ok {
				sendComponents = append(sendComponents, componentMap)
			}
		}
	} else {
		u.zsLog.Errorf("[SendBroadcast] failed to parse broadcast payload components: %v", err)
		finalStatus = dto.BroadcastScheduleFailed
		return true, err
	}
	sendParameter, err := u.whatsappService.ExtractSendComponentParameterValues(*broadcast.ParameterFormat, sendComponents)
	if err != nil {
		u.zsLog.Errorf("[SendBroadcast] failed to extract parameter values from broadcast payload: %v", err)
		finalStatus = dto.BroadcastScheduleFailed
		return true, err
	}
	templateFields, err := u.tenantRepository.GetTemplateFields(ctx, broadcast.TenantID)
	if err != nil {
		u.zsLog.Errorf("[SendBroadcast] failed to get template fields: %v", err)
		finalStatus = dto.BroadcastScheduleFailed
		return true, err
	}
	// replace fields
	// TODO: maybe make a default data for replacement if the field is not found in contact
	replaceVariable := make(map[string]string)
	for _, parameters := range sendParameter {
		for _, parameterValue := range parameters {
			parameterValue = u.whatsappService.ParseTemplateComponentParameter(parameterValue)
			// replace if its a format from template fields
			if field, ok := templateFields[parameterValue]; ok {
				replaceVariable[parameterValue] = field.Field
			}
		}
	}
	var contacts map[string]map[string]string
	if len(replaceVariable) > 0 {
		var phoneNumbers []string
		for _, recipeint := range broadcastRecipients {
			phoneNumbers = append(phoneNumbers, recipeint.RecipientId)
		}
		contacts, err = u.tenantRepository.GetContactByPhoneNumbers(ctx, broadcast.TenantID, phoneNumbers)
		if err != nil {
			u.zsLog.Errorf("[SendBroadcast] failed to get contacts by phone numbers: %v", err)
			finalStatus = dto.BroadcastScheduleFailed
			return true, err
		}
	}
	var broadcastStatus dto.BroadcastScheduleStatus
	if len(broadcastRecipients) == 0 {
		broadcastStatus = dto.BroadcastScheduleCancelled
	} else {
		broadcastStatus = dto.BroadcastScheduleSending
	}
	if broadcast.Status != string(dto.BroadcastScheduleScheduled) {
		u.zsLog.Errorf("[SendBroadcast] broadcast with ID %s is not in scheduled status, skipping sending", broadcastID)
		return false, nil
	}
	broadcast.Status = string(broadcastStatus)
	broadcast.UpdatedAt = time.Now()
	err = u.broadcastRepository.Update(ctx, nil, broadcast)
	if err != nil {
		u.zsLog.Errorf("[SendBroadcast] failed to update broadcast status: %v", err)
		return true, err
	}
	if broadcastStatus == dto.BroadcastScheduleCancelled {
		u.zsLog.Errorf("[SendBroadcast] broadcast with ID %s has no recipients, marking as cancelled", broadcastID)
		return false, nil
	}
	type broadcastSend struct {
		Broadcast model.Broadcast
		Recipient model.BroadcastRecipient
	}

	var hasError atomic.Bool
	var hasSuccess atomic.Bool
	// limit to log(n) workers with max of 100 workers
	workerSize := int(math.Max(1, math.Min(math.Log(float64(len(broadcastRecipients))), 100)))
	workers := make(chan broadcastSend, workerSize)
	recipientStatusUpdates := make(chan model.BroadcastRecipient, workerSize)
	workerWg := sync.WaitGroup{}
	statusWg := sync.WaitGroup{}
	statusWg.Go(func() {
		for recipient := range recipientStatusUpdates {
			err := u.broadcastRepository.UpdateRecipientStatus(ctx, nil, recipient)
			if err != nil {
				u.zsLog.Errorf("[SendBroadcast] failed to update recipient status: %v", err)
			}
		}
	})
	// TODO: implement retry mechanism for failed messages
	for i := 0; i < cap(workers); i++ {
		workerWg.Go(func() {
			for req := range workers {
				payloadMap := make(map[string]any)
				err := json.Unmarshal([]byte(req.Broadcast.Payload), &payloadMap)
				if err != nil {
					u.zsLog.Errorf("[SendBroadcast] failed to unmarshal broadcast payload: %v", err)
					continue
				}
				inputData := dto.MessageSendRequest{
					ChatID:     req.Recipient.RecipientId + "-" + req.Broadcast.PhoneNumberId, // TODO: support group recipient type
					SenderName: "Broadcast: " + req.Broadcast.Name,
					Type:       "template",
					Payload:    payloadMap["template"],
				}
				_, _, err = u.messageUsecase.SendMessage(ctx, whatsappClient, req.Broadcast.TenantID, inputData)
				if err != nil {
					u.zsLog.Errorf("[SendBroadcast] failed to send message to recipient: %v", err)
					errStr := err.Error()
					req.Recipient.Status = string(dto.BroadcastScheduleFailed)
					req.Recipient.Errors = &errStr
					hasError.Store(true)
				} else {
					req.Recipient.Status = string(dto.BroadcastScheduleSuccess)
					req.Recipient.Errors = nil
					hasSuccess.Store(true)
				}
				req.Recipient.UpdatedAt = time.Now()
				recipientStatusUpdates <- req.Recipient
				time.Sleep(time.Second * 1) // add delay between sending messages to avoid hitting rate limits
			}
		})
	}
	for _, recipient := range broadcastRecipients {
		broadcastCopy := broadcast
		var errStr string
		for key, value := range replaceVariable {
			broadcastPayload := make(map[string]any)
			err := json.Unmarshal([]byte(broadcastCopy.Payload), &broadcastPayload)
			if err != nil {
				u.zsLog.Errorf("[SendBroadcast] failed to unmarshal broadcast payload for recipient: %v", err)
				errStr = err.Error()
				break
			}
			if templatePayload, ok := broadcastPayload["template"].(map[string]any); ok {
				componentsBytes, err := json.Marshal(templatePayload["components"])
				if err != nil {
					u.zsLog.Errorf("[SendBroadcast] failed to marshal broadcast components for recipient: %v", err)
					errStr = err.Error()
					break
				}
				componentsStr := string(componentsBytes)
				replaceWith, exists := contacts[recipient.RecipientId][value]
				if !exists || replaceWith == "" {
					replaceWith = "-"
				}
				componentsStr = strings.ReplaceAll(componentsStr, "{{"+key+"}}", replaceWith)
				var components []map[string]any
				err = json.Unmarshal([]byte(componentsStr), &components)
				if err != nil {
					u.zsLog.Errorf("[SendBroadcast] failed to unmarshal broadcast components after replacement for recipient: %v", err)
					errStr = err.Error()
					break
				}
				templatePayload["components"] = components
				broadcastPayload["template"] = templatePayload
				payloadBytes, err := json.Marshal(broadcastPayload)
				if err != nil {
					u.zsLog.Errorf("[SendBroadcast] failed to marshal broadcast payload after replacement for recipient: %v", err)
					errStr = err.Error()
					break
				}
				broadcastCopy.Payload = string(payloadBytes)
			} else {
				u.zsLog.Errorf("[SendBroadcast] failed to parse broadcast payload for recipient: %v", err)
				errStr = "failed to parse broadcast payload for recipient"
				break
			}
		}
		if errStr != "" {
			recipient.Status = string(dto.BroadcastScheduleFailed)
			recipient.Errors = &errStr
			recipient.UpdatedAt = time.Now()
			hasError.Store(true)
			recipientStatusUpdates <- recipient
			continue
		}
		workers <- broadcastSend{
			Broadcast: broadcastCopy,
			Recipient: recipient,
		}
	}
	close(workers)
	workerWg.Wait()
	close(recipientStatusUpdates)
	statusWg.Wait()
	if hasError.Load() {
		if hasSuccess.Load() {
			finalStatus = dto.BroadcastScheduleFailedPartially
		} else {
			finalStatus = dto.BroadcastScheduleFailed
		}
	} else {
		finalStatus = dto.BroadcastScheduleSuccess
	}
	return false, nil
}

func (u *BroadcastUsecase) CancelBroadcast(ctx context.Context, tenantID string, inputData dto.BroadcastCancelRequest) (bool, error) {
	broadcast, err := u.broadcastRepository.GetByID(ctx, inputData.ID)
	if err != nil {
		u.zsLog.Errorf("[CancelBroadcast] failed to get broadcast by ID: %v", err)
		return true, err
	}
	if broadcast.Status != string(dto.BroadcastScheduleScheduled) {
		u.zsLog.Errorf("[CancelBroadcast] broadcast with ID %s is not in scheduled status, cannot cancel", inputData.ID)
		return false, fmt.Errorf("broadcast currently in %s status, only broadcast in scheduled status can be cancelled", broadcast.Status)
	}
	if broadcast.TenantID != tenantID {
		u.zsLog.Errorf("[CancelBroadcast] broadcast with ID %s does not belong to tenant with ID %s, cannot cancel", inputData.ID, tenantID)
		return false, errs.ErrGenericForbidden
	}
	err = u.googleTaskService.DeleteBroadcastTask(inputData.ID)
	if err != nil {
		u.zsLog.Errorf("[CancelBroadcast] failed to delete broadcast task: %v", err)
		return true, err
	}
	broadcast.Status = string(dto.BroadcastScheduleCancelled)
	broadcast.UpdatedAt = time.Now()
	err = u.broadcastRepository.Update(ctx, nil, broadcast)
	if err != nil {
		u.zsLog.Errorf("[CancelBroadcast] failed to update broadcast status: %v", err)
		return true, err
	}
	return false, nil
}

func (u *BroadcastUsecase) GetFilteredBroadcast(ctx context.Context, tenantID string, inputData filter_request.FilterRequest[dto.BroadcastGetFilteredRequest]) (filter_request.FilterResponse[dto.BroadcastResponse], bool, error) {
	var emptyResponse filter_request.FilterResponse[dto.BroadcastResponse]
	broadcasts, err := u.broadcastRepository.GetFilteredByTenantID(ctx, tenantID, inputData)
	if err != nil {
		u.zsLog.Errorf("[GetFilteredBroadcast] failed to get filtered broadcasts: %v", err)
		return emptyResponse, true, err
	}
	return broadcasts, false, nil
}

func (u *BroadcastUsecase) GetFilteredBroadcastRecipients(ctx context.Context, tenantID string, inputData filter_request.FilterRequest[dto.BroadcastGetRecipientsFilteredRequest]) (filter_request.FilterResponse[dto.BroadcastRecipientResponse], bool, error) {
	var emptyResponse filter_request.FilterResponse[dto.BroadcastRecipientResponse]
	broadcast, err := u.broadcastRepository.GetByID(ctx, inputData.SpecificFilter.BroadcastID)
	if err != nil {
		u.zsLog.Errorf("[GetFilteredBroadcastRecipients] failed to get broadcast by ID: %v", err)
		return emptyResponse, true, err
	}
	if broadcast.TenantID != tenantID {
		u.zsLog.Infof("[GetFilteredBroadcastRecipients] broadcast with ID %s does not belong to tenant with ID %s, cannot retrieve recipients", inputData.SpecificFilter.BroadcastID, tenantID)
		return emptyResponse, true, errs.ErrGenericForbidden
	}
	recipients, err := u.broadcastRepository.GetRecipientsFiltered(ctx, inputData)
	if err != nil {
		u.zsLog.Errorf("[GetFilteredBroadcastRecipients] failed to get filtered broadcast recipients: %v", err)
		return emptyResponse, true, err
	}
	return recipients, false, nil
}
