package chat_usecase

import (
	"context"
	"fmt"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"

	"go.uber.org/zap"
)

type ChatUsecase struct {
	chatRepository repository.Chat
	zsLog          *zap.SugaredLogger
}

func NewChatUsecase(chatRepository repository.Chat, zsLog *zap.SugaredLogger) *ChatUsecase {
	return &ChatUsecase{
		chatRepository: chatRepository,
		zsLog:          zsLog,
	}
}

func (uc *ChatUsecase) GetChatByPhoneNumberID(ctx context.Context, tenantID string, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIdRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse], bool, error) {
	// TODO: check if phone number belongs to tenant
	response, err := uc.chatRepository.GetChatByPhoneNumberID(ctx, requestData)
	if err != nil {
		uc.zsLog.Errorf("[GetChatByPhoneNumberID] error while getting chat by phone number id: %v", err)
		return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIdResponse]{}, true, err
	}
	return response, false, nil
}

func (uc *ChatUsecase) CloseTicket(ctx context.Context, requestData dto.ChatCloseTicketRequest) (bool, error) {
	chat, err := uc.chatRepository.GetByID(ctx, requestData.ChatID)
	if err != nil {
		uc.zsLog.Errorf("[CloseTicket] error while getting chat by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return false, err
		}
		return true, err
	}
	if chat.ChatStatus == model.ChatStatusClosed {
		return false, nil
	}
	chat.ChatStatus = model.ChatStatusClosed
	chat.AgentID = nil
	chat.UpdatedAt = time.Now()
	_, _, err = uc.chatRepository.Upsert(ctx, nil, chat)
	if err != nil {
		uc.zsLog.Errorf("[CloseTicket] error while upserting chat: %v", err)
		return true, err
	}
	return false, nil
}

func (uc *ChatUsecase) AssignAgent(ctx context.Context, requestData dto.ChatAssignAgentRequest) (bool, error) {
	chat, err := uc.chatRepository.GetByID(ctx, requestData.ChatID)
	if err != nil {
		uc.zsLog.Errorf("[AssignAgent] error while getting chat by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return false, err
		}
		return true, err
	}
	if chat.ChatStatus == model.ChatStatusClosed {
		return false, fmt.Errorf("cannot assign agent to closed chat")
	}
	chat.AgentID = &requestData.AgentID
	chat.UpdatedAt = time.Now()
	_, _, err = uc.chatRepository.Upsert(ctx, nil, chat)
	if err != nil {
		uc.zsLog.Errorf("[AssignAgent] error while upserting chat: %v", err)
		return true, err
	}
	return false, nil
}

func (uc *ChatUsecase) GetByID(ctx context.Context, chatID string) (model.Chat, bool, error) {
	chat, err := uc.chatRepository.GetByID(ctx, chatID)
	if err != nil {
		uc.zsLog.Errorf("[GetByID] error while getting chat by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return model.Chat{}, false, err
		}
		return model.Chat{}, true, err
	}
	return chat, false, nil
}
