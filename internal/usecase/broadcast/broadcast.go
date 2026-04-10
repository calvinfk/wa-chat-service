package broadcast_usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/meta/whatsapp_business/message_components"

	"github.com/google/uuid"
)

type BroadcastUsecase struct {
	templateRepository  repository.Template
	broadcastRepository repository.Broadcast
	tenantUsecase       usecase.Tenant
	googleTaskService   service.GoogleTask
}

func NewBroadcastUsecase(templateRepository repository.Template, broadcastRepository repository.Broadcast, tenantUsecase usecase.Tenant, googleTaskService service.GoogleTask) *BroadcastUsecase {
	return &BroadcastUsecase{
		templateRepository:  templateRepository,
		broadcastRepository: broadcastRepository,
		tenantUsecase:       tenantUsecase,
		googleTaskService:   googleTaskService,
	}
}

func (u *BroadcastUsecase) ScheduleBroadcast(ctx context.Context, inputData dto.BroadcastScheduleRequest) (bool, error) {
	_, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get whatsapp client: ", err)
		return true, err
	}
	template, err := u.templateRepository.GetByID(ctx, tenantID, inputData.TemplateID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get template: ", err)
		return true, err
	}
	// TODO: validate components before creating template component
	// var sendTemplate map[string]any
	payload := make(map[string]any)
	payload["template"] = map[string]any{
		"name":     template.Name,
		"language": map[string]any{"code": template.Language},
	}
	payload["template"].(map[string]any)["components"] = inputData.Components
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to marshal template payload: ", err)
		return true, err
	}
	templateComponent, err := whatsapp_business.NewTemplateComponent(payloadBytes)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to create template component: ", err)
		return true, err
	}
	broadcastID, err := uuid.NewV7()
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to generate broadcast ID: ", err)
		return true, err
	}
	payloadString, err := json.Marshal(templateComponent.GetPayload())
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to marshal template payload: ", err)
		return true, err
	}
	var sendingTime time.Time
	if inputData.SendAt == nil {
		sendingTime = time.Now().Add(time.Second * 10) // default to 10 seconds later if send_at is not provided
	} else {
		sendingTime = *inputData.SendAt
	}
	broadcast := model.Broadcast{
		DocumentID:    broadcastID.String(),
		Name:          inputData.Name,
		TemplateID:    inputData.TemplateID,
		PhoneNumberID: inputData.PhoneNumberID,
		Payload:       string(payloadString),
		Status:        string(dto.BroadcastScheduleScheduled),
		SendAt:        sendingTime,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = u.broadcastRepository.Insert(ctx, broadcast)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to insert broadcast: ", err)
		return true, err
	}
	for _, recipient := range inputData.Recipients {
		broadcastRecipient := model.BroadcastRecipient{
			DocumentID:    uuid.NewString(),
			BroadcastID:   broadcast.DocumentID,
			RecipientID:   recipient,
			RecipientType: "individual", // TODO: support group recipient type
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		err = u.broadcastRepository.InsertRecipient(ctx, broadcastRecipient)
	}
	err = u.googleTaskService.CreateBroadcastTask(broadcast.DocumentID, sendingTime)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to create broadcast task: ", err)
		return true, err
	}
	return false, nil
}

func (u *BroadcastUsecase) SendBroadcast(ctx context.Context, broadcastID string) (bool, error) {
	broadcast, err := u.broadcastRepository.GetByID(ctx, broadcastID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to get broadcast by ID: ", err)
		return true, err
	}
	whatsappClient, _, err := u.tenantUsecase.GetWhatsappClient(ctx, broadcast.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to get whatsapp client: ", err)
		return true, err
	}
	broadcastRecipients, err := u.broadcastRepository.GetRecipietsByBroadcastID(ctx, broadcastID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to list broadcast recipients: ", err)
		return true, err
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
	err = u.broadcastRepository.Update(ctx, broadcast)
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
	// limit to 10 concurrent workers = 10 message per second
	workers := make(chan broadcastSend, 10)
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for i := 0; i < cap(workers); i++ {
		wg.Go(func() {
			for req := range workers {
				err := u.sendMessageToRecipient(whatsappClient, req.Broadcast, req.Recipient)
				if err != nil {
					log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to send message to recipient: ", err)
				} else {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
				time.Sleep(time.Second * 1) // add delay between sending messages to avoid hitting rate limits
			}
		})
	}
	for _, recipient := range broadcastRecipients {
		workers <- broadcastSend{
			Broadcast: broadcast,
			Recipient: recipient,
		}
	}
	close(workers)
	wg.Wait()
	if successCount == len(broadcastRecipients) {
		broadcast.Status = string(dto.BroadcastScheduleSuccess)
	} else if successCount == 0 {
		broadcast.Status = string(dto.BroadcastScheduleFailed)
	} else {
		broadcast.Status = string(dto.BroadcastScheduleFailedPartially)
	}
	broadcast.UpdatedAt = time.Now()
	err = u.broadcastRepository.Update(ctx, broadcast)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to update broadcast status after sending: ", err)
		return true, err
	}
	return false, nil
}

func (u *BroadcastUsecase) sendMessageToRecipient(client *whatsapp_business.Client, broadcast model.Broadcast, recipient model.BroadcastRecipient) error {
	var recipientStatus string
	var template message_components.Template
	var err error
	defer func() {
		if recipientStatus == "" {
			recipientStatus = "delivered"
		}
		var errors *string
		if err != nil {
			errStr := err.Error()
			errors = &errStr
		}
		recipient.Status = recipientStatus
		recipient.Errors = errors
		recipient.UpdatedAt = time.Now()
		if err := u.broadcastRepository.UpdateRecipientStatus(context.Background(), recipient); err != nil {
			log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][sendMessageToRecipient] failed to update broadcast recipient status: ", err)
		}
	}()
	log.Println("[INFO][internal/usecase/broadcast/broadcast.go][sendMessageToRecipient] sending payload: ", broadcast.Payload)
	template, err = whatsapp_business.NewTemplateComponent([]byte(broadcast.Payload))
	if err != nil {
		recipientStatus = "failed"
		err = fmt.Errorf("failed to create template component: %v", err)
		return err
	}
	_, _, err = client.SendMessage(recipient.RecipientID, recipient.RecipientType, template)
	if err != nil {
		recipientStatus = "failed"
		err = fmt.Errorf("failed to send message to recipient: %v", err)
		return err
	}
	recipientStatus = "delivered"
	return nil
}
