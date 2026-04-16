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
	zslog          *zap.SugaredLogger
}

func NewChatUsecase(chatRepository repository.Chat, zslog *zap.SugaredLogger) *ChatUsecase {
	return &ChatUsecase{
		chatRepository: chatRepository,
		zslog:          zslog,
	}
}

func (uc *ChatUsecase) GetChatByPhoneNumberID(ctx context.Context, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse], bool, error) {
	response, err := uc.chatRepository.GetChatByPhoneNumberID(ctx, requestData)
	if err != nil {
		uc.zslog.Errorf("[GetChatByPhoneNumberID] error while getting chat by phone number id: %v", err)
		return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse]{}, true, err
	}
	return response, false, nil
}
