package wa_business_account_usecase

import (
	"context"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/meta/whatsapp_business"

	"go.uber.org/zap"
)

type WaBusinessAccountUsecase struct {
	waBusinessAccountRepository repository.WaBusinessAccount
	waPhoneRepository           repository.WaPhone
	encryptService              service.Encrypt
	zsLog                       *zap.SugaredLogger
}

func NewWaBusinessAccountUsecase(waBusinessAccountRepository repository.WaBusinessAccount, encryptService service.Encrypt, waPhoneRepository repository.WaPhone, zsLog *zap.SugaredLogger) *WaBusinessAccountUsecase {
	return &WaBusinessAccountUsecase{
		waBusinessAccountRepository: waBusinessAccountRepository,
		encryptService:              encryptService,
		waPhoneRepository:           waPhoneRepository,
		zsLog:                       zsLog,
	}
}

func (u *WaBusinessAccountUsecase) GetByPhoneNumberId(ctx context.Context, phoneNumberId string) (model.WaBusinessAccount, bool, error) {
	phone, err := u.waPhoneRepository.GetByPhoneNumberId(ctx, phoneNumberId)
	if err != nil {
		u.zsLog.Errorf("[GetByPhoneNumberId] Failed to get phone number data: %v", err)
		if err == errs.ErrGenericNotFound {
			return model.WaBusinessAccount{}, false, err
		}
		return model.WaBusinessAccount{}, true, err
	}
	account, err := u.waBusinessAccountRepository.GetByID(ctx, phone.WaBusinessAccountID)
	if err != nil {
		u.zsLog.Errorf("[GetByPhoneNumberId] Failed to get WhatsApp Business Account by phone number ID: %v", err)
		if err == errs.ErrGenericNotFound {
			return model.WaBusinessAccount{}, false, err
		}
		return model.WaBusinessAccount{}, true, err
	}
	return account, false, nil
}

func (u *WaBusinessAccountUsecase) GetWhatsappClient(ctx context.Context, tenantID string, phoneNumberId string) (*whatsapp_business.Client, string, error) {
	phone, err := u.waPhoneRepository.GetByPhoneNumberId(ctx, phoneNumberId)
	if err != nil {
		u.zsLog.Errorf("[getWhatsappClient] Failed to get phone number data: %v", err)
		return nil, "", err
	}
	waba, err := u.waBusinessAccountRepository.GetByID(ctx, phone.WaBusinessAccountID)
	if err != nil {
		u.zsLog.Errorf("[getWhatsappClient] Failed to get WhatsApp Business Account: %v", err)
		return nil, "", err
	}
	if waba.TenantID != tenantID {
		u.zsLog.Errorf("[getWhatsappClient] Tenant ID mismatch: expected %s, got %s", tenantID, waba.TenantID)
		return nil, "", errs.ErrGenericForbidden
	}
	whatsappClient, err := u.createClient(phoneNumberId, waba)
	if err != nil {
		u.zsLog.Errorf("[getWhatsappClient] Failed to create WhatsApp client: %v", err)
		return nil, "", err
	}
	return whatsappClient, waba.DocumentID, nil
}

func (u *WaBusinessAccountUsecase) GetWhatsappClientByWaBusinessAccountID(ctx context.Context, tenantID string, waBusinessAccountID string) (*whatsapp_business.Client, string, error) {
	waba, err := u.waBusinessAccountRepository.GetByID(ctx, waBusinessAccountID)
	if err != nil {
		u.zsLog.Errorf("[getWhatsappClientByWaBusinessAccountID] Failed to get WhatsApp Business Account: %v", err)
		return nil, "", err
	}
	if waba.TenantID != tenantID {
		u.zsLog.Errorf("[getWhatsappClientByWaBusinessAccountID] Tenant ID mismatch: expected %s, got %s", tenantID, waba.TenantID)
		return nil, "", errs.ErrGenericForbidden
	}
	whatsappClient, err := u.createClient("", waba)
	if err != nil {
		u.zsLog.Errorf("[getWhatsappClientByWaBusinessAccountID] Failed to create WhatsApp client: %v", err)
		return nil, "", err
	}
	return whatsappClient, waba.DocumentID, nil
}

func (u *WaBusinessAccountUsecase) createClient(phoneNumberId string, waba model.WaBusinessAccount) (*whatsapp_business.Client, error) {
	decyptedAccessToken, err := u.encryptService.Decrypt(waba.AccessToken)
	if err != nil {
		u.zsLog.Errorf("[getWhatsappClientByWaBusinessAccountID] Failed to decrypt access token: %v", err)
		return nil, err
	}
	whatsappClient := whatsapp_business.New(decyptedAccessToken, waba.WabaID, phoneNumberId)
	return whatsappClient, nil
}
