package auth_usecase

import (
	"context"
	"log"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"
)

type AuthUsecase struct {
	tenantRepository   repository.Tenant
	accessTokenService service.AccessToken
	encryptService     service.Encrypt
}

func NewAuthUsecase(tenantRepository repository.Tenant, accessTokenService service.AccessToken, encryptService service.Encrypt) *AuthUsecase {
	return &AuthUsecase{
		tenantRepository:   tenantRepository,
		accessTokenService: accessTokenService,
		encryptService:     encryptService,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, req dto.AuthLoginRequest) (string, bool, error) {
	tenant, err := u.tenantRepository.GetByPhoneNumberID(ctx, req.PhoneNumberID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/auth/auth.go][Login] tenantRepository.GetByPhoneNumberID error:", err)
		return "", true, err
	}
	if tenant.DocumentID != req.TenantID {
		log.Println("[ERROR][internal/usecase/auth/auth.go][Login] tenant.DocumentID != req.TenantID")
		return "", false, errs.ErrUserNotFound
	}
	accessToken, err := u.accessTokenService.GenerateAccessToken(tenant.DocumentID)
	if err != nil {
		log.Println("[ERROR][internal/usecase/auth/auth.go][Login] accessTokenService.GenerateAccessToken error:", err)
		return "", true, err
	}
	encryptedToken, err := u.encryptService.Encrypt(accessToken)
	if err != nil {
		log.Println("[ERROR][internal/usecase/auth/auth.go][Login] encryptService.Encrypt error:", err)
		return "", true, err
	}
	return encryptedToken, false, nil
}
