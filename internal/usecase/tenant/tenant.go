package tenant_usecase

import (
	"context"
	"fmt"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/filter_request"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TenantUsecase struct {
	tenantRepository repository.Tenant
	encryptService   service.Encrypt
	zsLog            *zap.SugaredLogger
}

func NewTenantUsecase(tenantRepository repository.Tenant, encryptService service.Encrypt, zsLog *zap.SugaredLogger) *TenantUsecase {
	return &TenantUsecase{
		tenantRepository: tenantRepository,
		encryptService:   encryptService,
		zsLog:            zsLog,
	}
}

func (u *TenantUsecase) CreateContact(ctx context.Context, tenantID string, inputData dto.ContactCreateRequest) (bool, error) {
	tenant, err := u.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[CreateContact] Failed to get tenant by ID: %v", err)
		return true, err
	}
	contacts, err := u.tenantRepository.GetContactByPhoneNumbers(ctx, tenant.DocumentID, []string{inputData.PhoneNumber})
	if err != nil {
		u.zsLog.Errorf("[CreateContact] Failed to get contacts by phone number: %v", err)
		return true, err
	}
	if contacts != nil {
		u.zsLog.Errorf("[CreateContact] Contact with the same phone number already exists")
		return false, fmt.Errorf("contact with the same phone number already exists")
	}
	contactID, err := uuid.NewV7()
	if err != nil {
		u.zsLog.Errorf("[CreateContact] Failed to generate contact ID: %v", err)
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
	err = u.tenantRepository.UpsertContact(ctx, nil, contact)
	if err != nil {
		u.zsLog.Errorf("[CreateContact] Failed to create contact: %v", err)
		return true, err
	}
	return false, nil
}

func (u *TenantUsecase) GetContactsFiltered(ctx context.Context, tenantID string, inputData filter_request.FilterRequest[dto.ContactGetFilteredRequest]) (filter_request.FilterResponse[dto.ContactResponse], bool, error) {
	tenant, err := u.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[GetContactsFiltered] Failed to get tenant by ID: %v", err)
		return filter_request.FilterResponse[dto.ContactResponse]{}, true, err
	}
	contacts, err := u.tenantRepository.GetContactsFiltered(ctx, tenant.DocumentID, inputData)
	if err != nil {
		u.zsLog.Errorf("[GetContactsFiltered] Failed to get filtered contacts: %v", err)
		return filter_request.FilterResponse[dto.ContactResponse]{}, true, err
	}
	return contacts, false, nil
}

func (u *TenantUsecase) UpdateContact(ctx context.Context, tenantID string, inputData dto.ContactUpdateRequest) (bool, error) {
	tenant, err := u.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[UpdateContact] Failed to get tenant by ID: %v", err)
		return true, err
	}
	contacts, err := u.tenantRepository.GetContactByPhoneNumbers(ctx, tenant.DocumentID, []string{inputData.PhoneNumber})
	if err != nil {
		u.zsLog.Errorf("[UpdateContact] Failed to get contacts by phone number: %v", err)
		return true, err
	}
	if len(contacts) > 0 && contacts[inputData.PhoneNumber]["__name__"] != inputData.ID {
		u.zsLog.Errorf("[UpdateContact] Contact with the same phone number already exists")
		return false, fmt.Errorf("contact with the same phone number already exists")
	}
	contact, err := u.tenantRepository.GetContactByID(ctx, tenant.DocumentID, inputData.ID)
	if err != nil {
		u.zsLog.Errorf("[UpdateContact] Failed to get contact by ID: %v", err)
		return true, err
	}
	contact.Name = inputData.Name
	contact.PhoneNumber = inputData.PhoneNumber
	contact.UpdatedAt = time.Now()
	err = u.tenantRepository.UpsertContact(ctx, nil, contact)
	if err != nil {
		u.zsLog.Errorf("[UpdateContact] Failed to update contact: %v", err)
		return true, err
	}
	return false, nil
}

func (u *TenantUsecase) DeleteContact(ctx context.Context, tenantID string, inputData dto.ContactDeleteRequest) (bool, error) {
	tenant, err := u.tenantRepository.GetByID(ctx, tenantID)
	if err != nil {
		u.zsLog.Errorf("[DeleteContact] Failed to get tenant by ID: %v", err)
		return true, err
	}
	err = u.tenantRepository.DeleteContact(ctx, nil, tenant.DocumentID, inputData.ID)
	if err != nil {
		u.zsLog.Errorf("[DeleteContact] Failed to delete contact: %v", err)
		return true, err
	}
	return false, nil
}
