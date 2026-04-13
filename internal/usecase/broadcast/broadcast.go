package broadcast_usecase

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/internal/usecase"
	"wa_chat_service/pkg/meta/whatsapp_business"
	"wa_chat_service/pkg/transaction"

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
	txManager           *transaction.TxManager
}

func NewBroadcastUsecase(templateRepository repository.Template, broadcastRepository repository.Broadcast, tenantRepository repository.Tenant, messageUsecase usecase.Message, tenantUsecase usecase.Tenant, googleTaskService service.GoogleTask, whatsappService service.WhatsappBusiness, txManager *transaction.TxManager) *BroadcastUsecase {
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
	whatsappClient, tenantID, err := u.tenantUsecase.GetWhatsappClient(ctx, inputData.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get whatsapp client: ", err)
		return true, err
	}
	template, err := u.templateRepository.GetByID(ctx, tenantID, inputData.TemplateID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get template: ", err)
		return true, err
	}
	phoneNumbers, err := u.tenantRepository.GetContactByPhoneNumbers(ctx, tenantID, inputData.Recipients)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to get contacts by phone numbers: ", err)
		return true, err
	}
	// TODO: validate components before creating template component
	// var sendTemplate map[string]any
	payload := make(map[string]any)
	payload["template"] = map[string]any{
		"name":       template.Name,
		"language":   map[string]any{"code": template.Language},
		"components": inputData.Components,
	}
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
	err = u.whatsappService.ValidateTemplatePayload(whatsappClient, template, templateComponent)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to validate template payload: ", err)
		return false, err
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
	serverError, err := u.txManager.DoFirestore(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error) {
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
		err = u.broadcastRepository.Insert(ctx, txFirestore, broadcast)
		if err != nil {
			log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to insert broadcast: ", err)
			return true, err
		}
		for _, recipient := range inputData.Recipients {
			var recipientName string
			contact, exists := phoneNumbers[recipient]
			if exists {
				recipientName = contact.Name
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
			err = u.broadcastRepository.InsertRecipient(ctx, txFirestore, broadcastRecipient)
			if err != nil {
				log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to insert broadcast recipient: ", err)
				return true, err
			}
		}
		// err = fmt.Errorf("testing")
		// if err != nil {
		// 	return false, err
		// }
		err = u.googleTaskService.CreateBroadcastTask(broadcast.DocumentID, sendingTime)
		if err != nil {
			log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] failed to create broadcast task: ", err)
			return true, err
		}
		return false, nil
	})
	return serverError, err
}

func (u *BroadcastUsecase) SendBroadcast(ctx context.Context, broadcastID string) (bool, error) {
	broadcast, err := u.broadcastRepository.GetByID(ctx, broadcastID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to get broadcast by ID: ", err)
		return true, err
	}
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
	// limit to 10 concurrent workers = 10 message per second
	workers := make(chan broadcastSend, 10)
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
	err = u.broadcastRepository.Update(ctx, nil, broadcast)
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][SendBroadcast] failed to update broadcast status after sending: ", err)
		return true, err
	}
	return false, nil
}
