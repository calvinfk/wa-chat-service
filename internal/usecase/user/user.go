package user_usecase

import (
	"context"
	"fmt"
	"time"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

func (uc *UserUsecase) GetByID(ctx context.Context, tenantID string, requestData dto.UserGetByIDRequest) (dto.UserResponse, bool, error) {
	user, err := uc.userRepository.GetByID(ctx, requestData.ID)
	if err != nil {
		uc.zsLog.Errorf("[GetByID] error while getting user by id: %v", err)
		if err == errs.ErrGenericNotFound {
			return dto.UserResponse{}, false, err
		}
		return dto.UserResponse{}, true, err
	}
	return dto.UserResponse{}.FromModel(user), false, nil
}

func (uc *UserUsecase) Upsert(ctx context.Context, tenantID string, requestData dto.UserUpsertRequest) (dto.UserResponse, bool, error) {
	var err error
	user := model.User{
		TenantID:     tenantID,
		SupervisorID: requestData.SupervisorID,
		Role:         requestData.Role,
		Name:         requestData.Name,
		Email:        requestData.Email,
		CreatedAt:    time.Now(),
	}
	// Check if ID is provided, if yes then get the existing user and check if tenant ID matches, if not then return error
	if requestData.ID != nil {
		existingUser, err := uc.userRepository.GetByID(ctx, *requestData.ID)
		if err != nil {
			uc.zsLog.Errorf("[Upsert] error while getting user by id: %v", err)
			return dto.UserResponse{}, true, err
		}
		if existingUser.TenantID != tenantID {
			uc.zsLog.Errorf("[Upsert] user tenant id does not match request tenant id")
			return dto.UserResponse{}, false, errs.ErrGenericNotFound
		}
		user.Password = existingUser.Password
		user.DocumentID = *requestData.ID
		user.CreatedAt = existingUser.CreatedAt
	} else {
		userID, err := uuid.NewV7()
		if err != nil {
			uc.zsLog.Errorf("[Upsert] error while generating user ID: %v", err)
			return dto.UserResponse{}, true, err
		}
		user.DocumentID = userID.String()
		if requestData.Password != nil {
			pass, err := bcrypt.GenerateFromPassword([]byte(*requestData.Password), bcrypt.DefaultCost)
			if err != nil {
				uc.zsLog.Errorf("[Upsert] error while hashing password: %v", err)
				return dto.UserResponse{}, true, err
			}
			user.Password = string(pass)
		} else {
			uc.zsLog.Errorf("[Upsert] password is required for new user")
			return dto.UserResponse{}, false, fmt.Errorf("password is required for new user")
		}
	}
	// check if email exists for other user, if yes then return error
	existingUserByEmail, err := uc.userRepository.GetByEmail(ctx, requestData.Email)
	if err != nil && err != errs.ErrGenericNotFound {
		uc.zsLog.Errorf("[Upsert] error while getting user by email: %v", err)
		return dto.UserResponse{}, true, err
	}
	if err == nil && existingUserByEmail.DocumentID != user.DocumentID {
		uc.zsLog.Errorf("[Upsert] user with email %s already exists", requestData.Email)
		return dto.UserResponse{}, false, errs.ErrGenericAlreadyExists
	}
	upsertedUser, err := uc.userRepository.Upsert(ctx, nil, user)
	if err != nil {
		uc.zsLog.Errorf("[Upsert] error while upserting user: %v", err)
		return dto.UserResponse{}, true, err
	}
	return dto.UserResponse{}.FromModel(upsertedUser), false, nil
}
