package broadcast_usecase

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

type BroadcastUsecase struct {
	templateRepository  repository.Template
	broadcastRepository repository.Broadcast
	tenantRepository    repository.Tenant
	messageUsecase      usecase.Message
	tenantUsecase       usecase.Tenant
	googleTaskService   service.GoogleTask
	whatsappService     service.WhatsappBusiness
	txManager           *utils.TxManager
}

func NewBroadcastUsecase(templateRepository repository.Template, broadcastRepository repository.Broadcast, tenantRepository repository.Tenant, messageUsecase usecase.Message, tenantUsecase usecase.Tenant, googleTaskService service.GoogleTask, whatsappService service.WhatsappBusiness, txManager *utils.TxManager) *BroadcastUsecase {
	return &BroadcastUsecase{
		templateRepository:  templateRepository,
		broadcastRepository: broadcastRepository,
		tenantRepository:    tenantRepository,
		messageUsecase:      messageUsecase,
		tenantUsecase:       tenantUsecase,
		googleTaskService:   googleTaskService,
		whatsappService:     whatsappService,
		txManager:           txManager,
	}
}

func (u *BroadcastUsecase) ScheduleBroadcast(ctx context.Context, inputData dto.BroadcastScheduleRequest) (bool, error) {
	broadcast, err := u.broadcastRepository.GetByID(ctx, inputData.ID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get broadcast by ID: ", err)
		return true, err
	}
	if broadcast.Status != string(dto.BroadcastScheduleDraft) {
		log.Printf("[INFO][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] broadcast with ID %s is not in draft status, cannot schedule", inputData.ID)
		return false, nil
	}
	broadcast.Status = string(dto.BroadcastScheduleScheduled)
	if inputData.SendAt != nil {
		broadcast.SendAt = *inputData.SendAt
	} else if broadcast.SendAt.IsZero() || broadcast.SendAt.Before(time.Now()) {
		broadcast.SendAt = time.Now().Add(time.Second * 10) // default to 10 seconds later if send_at is not provided or send_at is in the past
	}
	err = u.broadcastRepository.Update(ctx, nil, broadcast)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to update broadcast status: ", err)
		return true, err
	}
	template, err := u.templateRepository.GetByID(ctx, broadcast.TenantID, broadcast.TemplateID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get template: ", err)
		return true, err
	}
	serverError, err := u.createScheduleBroadcastTask(ctx, broadcast, template)
	return serverError, err
}

func (u *BroadcastUsecase) createScheduleBroadcastTask(ctx context.Context, broadcast model.Broadcast, template model.Template) (bool, error) {
	var broadcastPayload map[string]any
	err := json.Unmarshal([]byte(broadcast.Payload), &broadcastPayload)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to unmarshal template components: ", err)
		return true, err
	}
	templateSend, ok := broadcastPayload["template"].(map[string]any)
	if !ok {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to parse template payload: ", err)
		return true, err
	}
	var templateSendComponents []map[string]any
	if components, ok := templateSend["components"].([]any); ok {
		for _, component := range components {
			if componentMap, ok := component.(map[string]any); ok {
				templateSendComponents = append(templateSendComponents, componentMap)
			}
		}
	} else {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to parse template payload components: ", err)
		return true, err
	}
	quickReplyPayload, err := u.injectQuickReplyPayload(broadcast.DocumentID, template)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to inject quick reply payload: ", err)
		return true, err
	}
	for _, payload := range quickReplyPayload {
		templateSendComponents = append(templateSendComponents, payload)
	}
	templateSend["components"] = templateSendComponents
	broadcastPayload["template"] = templateSend
	payloadBytes, err := json.Marshal(broadcastPayload)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to marshal broadcast payload with quick reply: ", err)
		return true, err
	}
	broadcast.Payload = string(payloadBytes)
	err = u.broadcastRepository.Update(ctx, nil, broadcast)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to update broadcast with quick reply payload: ", err)
		return true, err
	}

	phoneNumbers, err := u.tenantRepository.GetContactByPhoneNumbers(ctx, broadcast.TenantID, broadcast.RecipientIDs)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get contacts by phone numbers: ", err)
		return true, err
	}
	for _, recipient := range broadcast.RecipientIDs {
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
			RecipientID:   recipient,
			RecipientName: recipientName,
			RecipientType: "individual", // TODO: support group recipient type
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		err = u.broadcastRepository.InsertRecipient(ctx, nil, broadcastRecipient)
		if err != nil {
			log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to insert broadcast recipient: ", err)
			return true, err
		}
	}

	err = u.googleTaskService.CreateBroadcastTask(broadcast.DocumentID, broadcast.SendAt)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to create broadcast task: ", err)
		return true, err
	}
	return false, nil
}

func (u *BroadcastUsecase) UpsertBroadcast(ctx context.Context, inputData dto.BroadcastUpsertRequest) (dto.BroadcastResponse, bool, error) {
	var emptyResponse dto.BroadcastResponse
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get whatsapp client: ", err)
		return emptyResponse, true, err
	}
	template, err := u.templateRepository.GetByID(ctx, tenantID, inputData.TemplateID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get template: ", err)
		return emptyResponse, true, err
	}
	broadcastID, err := uuid.NewV7()
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to generate broadcast ID: ", err)
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
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to marshal template payload: ", err)
		return emptyResponse, true, err
	}
	templateComponent, err := whatsapp_business.NewTemplateComponent(payloadBytes)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to create template component: ", err)
		return emptyResponse, true, err
	}
	err = u.whatsappService.ValidateTemplatePayload(whatsappClient, template, templateComponent)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to validate template payload: ", err)
		return emptyResponse, true, err
	}
	payloadString, err := json.Marshal(templateComponent.GetPayload())
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to marshal template payload: ", err)
		return emptyResponse, true, err
	}
	var broadcast model.Broadcast
	serverError, err := u.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
		var sendingTime time.Time
		if inputData.SendAt == nil {
			sendingTime = time.Now().Add(time.Second * 10) // default to 10 seconds later if send_at is not provided
		} else {
			sendingTime = *inputData.SendAt
		}
		broadcast = model.Broadcast{
			DocumentID:      broadcastID.String(),
			Name:            inputData.Name,
			TemplateID:      inputData.TemplateID,
			TenantID:        tenantID,
			RecipientIDs:    inputData.Recipients,
			PhoneNumberID:   inputData.PhoneNumberID,
			ParameterFormat: template.ParameterFormat,
			Payload:         string(payloadString),
			Status:          inputData.Status,
			SendAt:          sendingTime,
			CreatedBy:       inputData.PhoneNumberID,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		err = u.broadcastRepository.Insert(ctx, txFirestore, broadcast)
		if err != nil {
			log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to insert broadcast: ", err)
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return emptyResponse, serverError, err
	}
	if broadcast.Status == string(dto.BroadcastScheduleScheduled) {
		serverError, err := u.createScheduleBroadcastTask(ctx, broadcast, template)
		if err != nil {
			log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to schedule broadcast: ", err)
			return emptyResponse, serverError, err
		}
	}
	return dto.BroadcastResponse{}.FromModel(broadcast), serverError, err
}

func (u *BroadcastUsecase) injectQuickReplyPayload(broadcastID string, template model.Template) ([]map[string]any, error) {
	var templateDBComponents []map[string]any
	err := json.Unmarshal([]byte(template.Components), &templateDBComponents)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get template components: ", err)
		return nil, err
	}
	// broadcast_{{broadcast-id}}_{{contact-phone}}_{{text}}
	var quickReplyPayload []map[string]any
	for _, component := range templateDBComponents {
		if component["type"] == "BUTTONS" {
			for i, buttonComponent := range component["buttons"].([]any) {
				button, err := whatsapp_business.NewTemplateCreateButton(buttonComponent)
				if err != nil {
					log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][injectQuickReplyPayload] failed to create button component: ", err)
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
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to get broadcast by ID: ", err)
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
			log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to update broadcast status: ", err)
		}
	}()
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, broadcast.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to get whatsapp client: ", err)
		return true, err
	}
	broadcastRecipients, err := u.broadcastRepository.GetRecipientsByBroadcastID(ctx, broadcastID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to list broadcast recipients: ", err)
		return true, err
	}
	// get tenant's template fields to validate the payload before sending messages
	payload := make(map[string]any)
	err = json.Unmarshal([]byte(broadcast.Payload), &payload)
	if err != nil {
		finalStatus = dto.BroadcastScheduleFailed
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to unmarshal broadcast payload: ", err)
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
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to parse broadcast payload components: ", err)
		finalStatus = dto.BroadcastScheduleFailed
		return true, err
	}
	sendParameter, err := u.whatsappService.ExtractSendComponentParameterValues(*broadcast.ParameterFormat, sendComponents)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to extract parameter values from broadcast payload: ", err)
		finalStatus = dto.BroadcastScheduleFailed
		return true, err
	}
	templateFields, err := u.tenantRepository.GetTemplateFields(ctx, tenantID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to get template fields: ", err)
		finalStatus = dto.BroadcastScheduleFailed
		return true, err
	}
	// replace fields
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
			phoneNumbers = append(phoneNumbers, recipeint.RecipientID)
		}
		contacts, err = u.tenantRepository.GetContactByPhoneNumbers(ctx, tenantID, phoneNumbers)
		if err != nil {
			log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to get contacts by phone numbers: ", err)
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
		log.Printf("[INFO][internal/usecase/broadcast/broadcast.go][SendBroadcast] broadcast with ID %s is not in scheduled status, skipping sending", broadcastID)
		return false, nil
	}
	broadcast.Status = string(broadcastStatus)
	broadcast.UpdatedAt = time.Now()
	err = u.broadcastRepository.Update(ctx, nil, broadcast)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to update broadcast status: ", err)
		return true, err
	}
	if broadcastStatus == dto.BroadcastScheduleCancelled {
		log.Printf("[INFO][internal/usecase/broadcast/broadcast.go][SendBroadcast] broadcast with ID %s has no recipients, marking as cancelled", broadcastID)
		return false, nil
	}
	type broadcastSend struct {
		Broadcast model.Broadcast
		Recipient model.BroadcastRecipient
	}

	successCount := 0
	// limit to 100 message per second
	// TODO: consider to make the worker pool size configurable based on tenant's whatsapp message sending limit and use dynamic worker pool size to optimize the sending process
	workers := make(chan broadcastSend, 100)
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for i := 0; i < cap(workers); i++ {
		wg.Go(func() {
			for req := range workers {
				payloadMap := make(map[string]any)
				err := json.Unmarshal([]byte(req.Broadcast.Payload), &payloadMap)
				if err != nil {
					log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to unmarshal broadcast payload: ", err)
					continue
				}
				inputData := dto.MessageSendRequest{
					PhoneNumberID: req.Broadcast.PhoneNumberID,
					RecipientID:   req.Recipient.RecipientID,
					RecipientName: req.Recipient.RecipientName,
					SenderName:    "Broadcast: " + req.Broadcast.Name,
					Type:          "template",
					Payload:       payloadMap["template"],
				}
				_, _, err = u.messageUsecase.SendMessage(ctx, whatsappClient, tenantID, inputData)
				if err != nil {
					log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to send message to recipient: ", err)
					errStr := err.Error()
					req.Recipient.Status = string(dto.BroadcastScheduleFailed)
					req.Recipient.Errors = &errStr
				} else {
					req.Recipient.Status = string(dto.BroadcastScheduleSuccess)
					mu.Lock()
					successCount++
					mu.Unlock()
				}
				req.Recipient.UpdatedAt = time.Now()
				err = u.broadcastRepository.UpdateRecipientStatus(ctx, nil, req.Recipient)
				if err != nil {
					log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to update broadcast recipient status: ", err)
				}
				time.Sleep(time.Second * 1) // add delay between sending messages to avoid hitting rate limits
			}
		})
	}

	// TODO: consider to use pubsub to handle sending messages in case the number of recipients is too large and might cause memory issue if we load all recipients into memory at once. For now we assume the number of recipients is manageable and can be loaded into memory.
	// TODO: consider to implement retry mechanism for failed messages
	// TODO: change only the components that have parameters instead of replacing the whole payload to avoid potential issue of replacing non-parameter value that has the same value as parameter format
	for _, recipient := range broadcastRecipients {
		broadcastCopy := broadcast
		for key, value := range replaceVariable {
			broadcastCopy.Payload = strings.ReplaceAll(broadcastCopy.Payload, "{{"+key+"}}", contacts[recipient.RecipientID][value])
		}
		workers <- broadcastSend{
			Broadcast: broadcastCopy,
			Recipient: recipient,
		}
	}
	close(workers)
	wg.Wait()
	if successCount == len(broadcastRecipients) {
		finalStatus = dto.BroadcastScheduleSuccess
	} else if successCount == 0 {
		finalStatus = dto.BroadcastScheduleFailed
	} else {
		finalStatus = dto.BroadcastScheduleFailedPartially
	}
	return false, nil
}

func (u *BroadcastUsecase) CancelBroadcast(ctx context.Context, broadcastID string) (bool, error) {
	broadcast, err := u.broadcastRepository.GetByID(ctx, broadcastID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][CancelBroadcast] failed to get broadcast by ID: ", err)
		return true, err
	}
	if broadcast.Status != string(dto.BroadcastScheduleScheduled) {
		log.Printf("[INFO][internal/usecase/broadcast/broadcast.go][CancelBroadcast] broadcast with ID %s is not in scheduled status, cannot cancel", broadcastID)
		return false, nil
	}
	err = u.googleTaskService.DeleteBroadcastTask(broadcastID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][CancelBroadcast] failed to delete broadcast task: ", err)
		return true, err
	}
	broadcast.Status = string(dto.BroadcastScheduleCancelled)
	broadcast.UpdatedAt = time.Now()
	err = u.broadcastRepository.Update(ctx, nil, broadcast)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][CancelBroadcast] failed to update broadcast status: ", err)
		return true, err
	}
	return false, nil
}
