package user_usecase

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/repository"
	"wa_chat_service/pkg/filter_request"

	"go.uber.org/zap"
)

type UserUsecase struct {
	userRepository repository.User
	zsLog          *zap.SugaredLogger
}

func NewUserUsecase(userRepository repository.User, zsLog *zap.SugaredLogger) *UserUsecase {
	return &UserUsecase{
		userRepository: userRepository,
		zsLog:          zsLog,
	}
}

func (uc *UserUsecase) GetByTenantIDFiltered(ctx context.Context, tenantID string, requestData filter_request.FilterRequest[dto.UserListRequest]) (filter_request.FilterResponse[dto.UserResponse], bool, error) {
	response, err := uc.userRepository.GetByTenantIDFiltered(ctx, tenantID, requestData)
	if err != nil {
		uc.zsLog.Errorf("[GetByTenantIDFiltered] error while getting users by tenant id: %v", err)
		return filter_request.FilterResponse[dto.UserResponse]{}, true, err
	}
	return response, false, nil
}
