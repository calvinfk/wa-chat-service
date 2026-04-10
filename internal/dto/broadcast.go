package dto

import "time"

type (
	BroadcastScheduleRequest struct {
		PhoneNumberID string           `json:"phone_number_id" validate:"required"`
		TemplateID    string           `json:"template_id" validate:"required"`
		Name          string           `json:"name" validate:"required"`
		SendAt        *time.Time       `json:"send_at" validate:"omitempty,gt"`
		Recipients    []string         `json:"recipients" validate:"required,min=1,dive,required"`
		Components    []map[string]any `json:"components" validate:"required,min=1,dive"`
	}
)

type BroadcastScheduleStatus string

const (
	BroadcastScheduleFailed          BroadcastScheduleStatus = "failed"
	BroadcastScheduleFailedPartially BroadcastScheduleStatus = "failed_partially"
	BroadcastScheduleCancelled       BroadcastScheduleStatus = "cancelled"
	BroadcastScheduleSuccess         BroadcastScheduleStatus = "success"
	BroadcastScheduleSending         BroadcastScheduleStatus = "sending"
	BroadcastScheduleScheduled       BroadcastScheduleStatus = "scheduled"
)
