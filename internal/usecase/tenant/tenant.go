package tenant_usecase

import (
	"context"
	"fmt"
	"log"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/meta/whatsapp_business"

	"github.com/google/uuid"
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

func (u *TenantUsecase) CreateContact(ctx context.Context, inputData dto.ContactCreateRequest) (bool, error) {
	tenant, err := u.tenantRepository.GetByID(ctx, inputData.TenantID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][CreateContact] Failed to get tenant by ID:", err)
		return true, err
	}
	contacts, err := u.tenantRepository.GetContactByPhoneNumbers(ctx, tenant.DocumentID, []string{inputData.PhoneNumber})
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][CreateContact] Failed to get contacts by phone number:", err)
		return true, err
	}
	if contacts != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][CreateContact] Contact with the same phone number already exists")
		return false, fmt.Errorf("contact with the same phone number already exists")
	}
	contactID, err := uuid.NewV7()
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][CreateContact] Failed to generate contact ID:", err)
		return true, err
	}
	contact := model.Contact{
		DocumentID:  contactID.String(),
		TenantID:    tenant.DocumentID,
		Name:        inputData.Name,
		PhoneNumber: inputData.PhoneNumber,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = u.tenantRepository.InsertContact(ctx, contact)
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][CreateContact] Failed to create contact:", err)
		return true, err
	}
	return false, nil
}

func (u *TenantUsecase) GetContactsFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.ContactGetFilteredRequest]) (filter_request.FilterResponse[dto.ContactResponse], bool, error) {
	tenant, err := u.tenantRepository.GetByID(ctx, inputData.SpecificFilter.TenantID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][GetContactsFiltered] Failed to get tenant by ID:", err)
		return filter_request.FilterResponse[dto.ContactResponse]{}, true, err
	}
	contacts, err := u.tenantRepository.GetContactsFiltered(ctx, tenant.DocumentID, inputData)
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][GetContactsFiltered] Failed to get filtered contacts:", err)
		return filter_request.FilterResponse[dto.ContactResponse]{}, true, err
	}
	return contacts, false, nil
}

func (u *TenantUsecase) UpdateContact(ctx context.Context, inputData dto.ContactUpdateRequest) (bool, error) {
	tenant, err := u.tenantRepository.GetByID(ctx, inputData.TenantID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][UpdateContact] Failed to get tenant by ID:", err)
		return true, err
	}
	contacts, err := u.tenantRepository.GetContactByPhoneNumbers(ctx, tenant.DocumentID, []string{inputData.PhoneNumber})
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][UpdateContact] Failed to get contacts by phone number:", err)
		return true, err
	}
	if contacts != nil && contacts[inputData.PhoneNumber]["__name__"] != inputData.ID {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][UpdateContact] Contact with the same phone number already exists")
		return true, fmt.Errorf("contact with the same phone number already exists")
	}
	contact, err := u.tenantRepository.GetContactByID(ctx, tenant.DocumentID, inputData.ID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][UpdateContact] Failed to get contact by ID:", err)
		return true, err
	}
	contact.Name = inputData.Name
	contact.PhoneNumber = inputData.PhoneNumber
	contact.UpdatedAt = time.Now()
	err = u.tenantRepository.UpdateContact(ctx, contact)
	if err != nil {
		log.Println("[ERROR][internal/usecase/tenant/tenant.go][UpdateContact] Failed to update contact:", err)
		return true, err
	}
	return false, nil
}
