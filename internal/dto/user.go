package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	UserListRequest struct {
		Role *string `json:"role" query:"role" validate:"omitempty,oneof=admin agent supervisor"`
	}
	UserResponse struct {
		ID           string         `json:"id"`
		TenantID     string         `json:"tenant_id"` // foreign key to tenant
		SupervisorID *string        `json:"supervisor_id"`
		Role         model.UserRole `json:"role"`
		Name         string         `json:"name"`
		Email        string         `json:"email"`
		CreatedAt    time.Time      `json:"created_at"`
	}
	UserUpsertRequest struct {
		ID           *string        `json:"id,omitempty"`            // if ID is provided, it will update the existing user, otherwise it will create a new user
		SupervisorID *string        `json:"supervisor_id,omitempty"` // foreign key to supervisor, nullable
		Role         model.UserRole `json:"role" validate:"required,oneof=admin agent supervisor"`
		Name         string         `json:"name" validate:"required"`
		Email        string         `json:"email" validate:"required,email"`
		Password     *string        `json:"password" validate:"omitempty,min=8"`
	}

	UserGetByIDRequest struct {
		ID string `query:"id" validate:"required"`
	}
)

func (r UserListRequest) Validate() map[string]string {
	validator := utils.NewValidator()
	err := validator.Struct(r)
	if err != nil {
		return utils.GetValidatorErrorMessages(err)
	}
	return nil
}

func (UserResponse) FromModel(user model.User) UserResponse {
	return UserResponse{
		ID:           user.DocumentID,
		TenantID:     user.TenantID,
		SupervisorID: user.SupervisorID,
		Role:         user.Role,
		Name:         user.Name,
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
	}
}
