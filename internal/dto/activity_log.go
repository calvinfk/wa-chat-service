package dto

import (
	"time"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"

	"github.com/google/uuid"
)

type (
	// Data required to create a new activity log entry.
	ActivityLogCreateRequest struct {
		UserID      *uuid.UUID
		Type        string
		Description string
	}

	// Data used for filtering activity log entries when retrieving them.The user ID filter can also be used to filter for null values by using a specific UUID value (00000000-0000-0000-0000-000000000000).
	ActivityLogFilterRequest struct {
		ID     *filter_request.QueryFilterUUID   `json:"id" form:"id"`
		UserID *filter_request.QueryFilterUUID   `json:"user_id" form:"user_id"` // use 00000000-0000-0000-0000-000000000000 for filtering null user_id
		Type   *filter_request.QueryFilterString `json:"type" form:"type"`
	}

	// Data returned in the response when retrieving activity log entries.
	ActivityLogResponse struct {
		ID          uuid.UUID  `json:"id"`
		UserID      *uuid.UUID `json:"user_id"`
		UserName    *string    `json:"user_name"`
		UserRole    *string    `json:"user_role"`
		UserEmail   *string    `json:"user_email"`
		Type        string     `json:"type"`
		Description string     `json:"description"`
		CreatedAt   time.Time  `json:"created_at"`
	}
)

func (r ActivityLogFilterRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.ID != nil {
		if r.ID.IsEmpty() {
			errors["id"] = errs.ErrValidateEmptyField
		} else if !r.ID.IsValid() {
			errors["id"] = errs.ErrValidateInvalidUUID
		}
	}
	if r.UserID != nil {
		if r.UserID.IsEmpty() {
			errors["user_id"] = errs.ErrValidateEmptyField
		} else if !r.UserID.IsValid() {
			errors["user_id"] = errs.ErrValidateInvalidUUID
		}
	}
	if r.Type != nil && r.Type.IsEmpty() {
		errors["type"] = errs.ErrValidateEmptyField
	}
	return errors
}

func (r *ActivityLogResponse) FromModel(data model.ActivityLog) {
	r.ID = data.ID
	r.UserID = data.UserID
	r.Type = data.Type
	r.Description = data.Description
	r.CreatedAt = data.CreatedAt
}
