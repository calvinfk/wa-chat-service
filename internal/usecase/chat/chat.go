package chat_usecase

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/repository"
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

func (uc *ChatUsecase) GetChatByPhoneNumberID(ctx context.Context, tenantID string, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse], bool, error) {
	// RODO: check if phone number belongs to tenant
	response, err := uc.chatRepository.GetChatByPhoneNumberID(ctx, requestData)
	if err != nil {
		uc.zsLog.Errorf("[GetChatByPhoneNumberID] error while getting chat by phone number id: %v", err)
		return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse]{}, true, err
	}
	return response, false, nil
}
