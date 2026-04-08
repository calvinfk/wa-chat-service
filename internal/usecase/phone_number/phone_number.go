package phone_number_usecase

import (
	"context"
	"log"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/meta/whatsapp_business"
)

type PhoneNumberUsecase struct {
	phoneNumberRepository repository.PhoneNumber
	encryptService        service.Encrypt
}

func NewPhoneNumberUsecase(phoneNumberRepository repository.PhoneNumber, encryptService service.Encrypt) *PhoneNumberUsecase {
	return &PhoneNumberUsecase{
		phoneNumberRepository: phoneNumberRepository,
		encryptService:        encryptService,
	}
}

func (u *PhoneNumberUsecase) GetWhatsappClient(ctx context.Context, phoneNumberID string) (*whatsapp_business.Client, error) {
	phoneNumber, err := u.phoneNumberRepository.GetByPhoneNumberID(ctx, phoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to get phone number:", err)
		return nil, err
	}
	decyptedAccessToken, err := u.encryptService.Decrypt(phoneNumber.AccessToken)
	if err != nil {
		log.Println("[ERROR][internal/usecase/message/message.go][SendMessage] Failed to decrypt access token:", err)
		return nil, err
	}
	whatsappClient := whatsapp_business.New(decyptedAccessToken, phoneNumber.WabaID, phoneNumber.PhoneNumberID)
	return whatsappClient, nil
}
