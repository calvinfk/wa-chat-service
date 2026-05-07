package dto

import "time"

type (
	TicketCreateRequest struct {
		PhoneNumberId string `json:"phone_number_id" validate:"required"`
		RecipientId   string `json:"recipient_id" validate:"required"`
		RecipientName string `json:"recipient_name" validate:"required"`
	}
	TicketGetAnalyticsRequest struct {
		PhoneNumberIds *[]string `query:"phone_number_ids" validate:"omitempty,min=1,dive,required"` // if provided, must contain at least 1 phone number ID and each ID is required
		StartTime      time.Time `query:"start_time" validate:"required,lt"`
		EndTime        time.Time `query:"end_time" validate:"required,gtefield=StartTime"`
	}
	TicketCloseRequest struct {
		TicketID string `json:"ticket_id" validate:"required"`
	}
	TicketAssignAgentRequest struct {
		TicketID string `json:"ticket_id" validate:"required"`
		AgentID  string `json:"agent_id" validate:"required,uuid"`
	}
	TicketGetAnalyticsResponse struct {
		TotalCount                int `json:"total_count"`
		AverageResolutionMinutes  int `json:"average_resolution_minutes"`
		MedianResolutionMinutes   int `json:"median_resolution_minutes"`
		LongestResolutionMinutes  int `json:"longest_resolution_minutes"`
		ShortestResolutionMinutes int `json:"shortest_resolution_minutes"`
		OpenedCount               int `json:"opened_count"`
		InProgressCount           int `json:"in_progress_count"`
		ClosedCount               int `json:"closed_count"`
	}
)
