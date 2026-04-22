package auth_usecase

import (
	"context"
	"fmt"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"

	"go.uber.org/zap"
)

type AuthUsecase struct {
	tenantRepository   repository.Tenant
	accessTokenService service.AccessToken
	encryptService     service.Encrypt
	zslog              *zap.SugaredLogger
}

func NewAuthUsecase(tenantRepository repository.Tenant, accessTokenService service.AccessToken, encryptService service.Encrypt, zslog *zap.SugaredLogger) *AuthUsecase {
	return &AuthUsecase{
		tenantRepository:   tenantRepository,
		accessTokenService: accessTokenService,
		encryptService:     encryptService,
		zslog:              zslog,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, req dto.AuthLoginRequest) (string, bool, error) {
	tenant, err := u.tenantRepository.GetByPhoneNumberID(ctx, req.PhoneNumberID)
	if err != nil {
		u.zslog.Errorf("[Login] tenantRepository.GetByPhoneNumberID error : %v", err)
		if err == errs.ErrGenericNotFound {
			return "", false, err
		}
		return "", true, err
	}
	if tenant.DocumentID != req.TenantID {
		u.zslog.Errorf("[Login] tenant.DocumentID != req.TenantID : %v", errs.ErrGenericNotFound)
		return "", false, errs.ErrGenericNotFound
	}
	sub := fmt.Sprintf("%s:%s", tenant.DocumentID, tenant.PhoneNumberID)
	accessToken, err := u.accessTokenService.GenerateAccessToken(sub)
	if err != nil {
		u.zslog.Errorf("[Login] accessTokenService.GenerateAccessToken error : %v", err)
		return "", true, err
	}
	encryptedToken, err := u.encryptService.Encrypt(accessToken)
	if err != nil {
		u.zslog.Errorf("[Login] encryptService.Encrypt error : %v", err)
		return "", true, err
	}
	return encryptedToken, false, nil
}
