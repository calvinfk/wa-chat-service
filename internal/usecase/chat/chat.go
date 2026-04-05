package chat_usecase

import (
	"context"
	"log"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/repository"
	"wa_chat_service/pkg/filter_request"
)

type ChatUsecase struct {
	chatRepository repository.Chat
}

func NewChatUsecase(chatRepository repository.Chat) *ChatUsecase {
	return &ChatUsecase{
		chatRepository: chatRepository,
	}
}

func (uc *ChatUsecase) GetChatByPhoneNumberID(ctx context.Context, requestData filter_request.FilterRequest[dto.ChatGetByPhoneNumberIDRequest]) (filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse], bool, error) {
	response, err := uc.chatRepository.GetChatByPhoneNumberID(ctx, requestData)
	if err != nil {
		log.Println("[ERROR][internal/usecase/chat/chat.go][GetChatByPhoneNumberID] error while getting chat by phone number id: ", err)
		return filter_request.FilterResponse[dto.ChatGetByPhoneNumberIDResponse]{}, true, err
	}
	return response, false, nil
}
