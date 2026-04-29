package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"
)

type (
	UserListRequest struct {
		Role *string `json:"role" query:"role" validate:"omitempty,oneof=admin agent"`
	}
	UserResponse struct {
		ID        string    `json:"id"`
		TenantID  string    `json:"tenant_id"` // foreign key to tenant
		Role      string    `json:"role"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
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
		ID:        user.DocumentID,
		TenantID:  user.TenantID,
		Role:      user.Role,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
