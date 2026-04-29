package dto

import "wa_chat_service/internal/model"

type (
	AuthLoginRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}
	AuthData struct {
		UserID   string         `json:"user_id"`
		TenantID string         `json:"tenant_id"`
		Role     model.UserRole `json:"role"`
	}
)
