package tenant_usecase

import (
	"context"
	"log"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/meta/whatsapp_business"
)

type TenantUsecase struct {
	tenantRepository repository.Tenant
	encryptService   service.Encrypt
}

func NewTenantUsecase(tenantRepository repository.Tenant, encryptService service.Encrypt) *TenantUsecase {
	return &TenantUsecase{
		tenantRepository: tenantRepository,
		encryptService:   encryptService,
	}
}

func (u *TenantUsecase) GetWhatsappClient(ctx context.Context, phoneNumberID string) (*whatsapp_business.Client, string, error) {
	tenant, err := u.tenantRepository.GetByPhoneNumberID(ctx, phoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][GetWhatsappClient] Failed to get phone number:", err)
		return nil, "", err
	}
	decyptedAccessToken, err := u.encryptService.Decrypt(tenant.AccessToken)
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][GetWhatsappClient] Failed to decrypt access token:", err)
		return nil, "", err
	}
	whatsappClient := whatsapp_business.New(decyptedAccessToken, tenant.WabaID, tenant.PhoneNumberID)
	return whatsappClient, tenant.DocumentID, nil
}
