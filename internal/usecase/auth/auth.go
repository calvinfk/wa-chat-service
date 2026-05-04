package auth_usecase

import (
	"context"
	"fmt"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/repository"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepository     repository.User
	tenantRepository   repository.Tenant
	accessTokenService service.AccessToken
	encryptService     service.Encrypt
	zsLog              *zap.SugaredLogger
}

func NewAuthUsecase(userRepository repository.User, tenantRepository repository.Tenant, accessTokenService service.AccessToken, encryptService service.Encrypt, zsLog *zap.SugaredLogger) *AuthUsecase {
	return &AuthUsecase{
		userRepository:     userRepository,
		tenantRepository:   tenantRepository,
		accessTokenService: accessTokenService,
		encryptService:     encryptService,
		zsLog:              zsLog,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, req dto.AuthLoginRequest) (string, bool, error) {
	user, err := u.userRepository.GetByEmail(ctx, req.Email)
	if err != nil {
		u.zsLog.Errorf("[Login] userRepository.GetByEmail error : %v", err)
		if err == errs.ErrGenericNotFound {
			return "", false, err
		}
		return "", true, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		u.zsLog.Errorf("[Login] bcrypt.CompareHashAndPassword error : %v", err)
		return "", true, errs.ErrGenericUnauthorized
	}
	sub := fmt.Sprintf("%s:%s:%s", user.TenantID, user.DocumentID, user.Role)
	accessToken := u.accessTokenService.GenerateAccessToken(sub)
	encryptedToken, err := u.encryptService.Encrypt(accessToken)
	if err != nil {
		u.zsLog.Errorf("[Login] encryptService.Encrypt error : %v", err)
		return "", true, err
	}
	return encryptedToken, false, nil
}
