package model

import (
	"time"
)

type (
	// Represents a log entry for user activities in the system.
	// UserID is a foreign key that references the User model.
	// The ID is generated as a UUID.
	// The CreatedAt field is automatically managed by GORM to track when record is created.
	ActivityLog struct {
		ID          string    `json:"__name__" firestore:"id"`
		UserID      *string   `json:"user_id" firestore:"user_id"`
		Type        string    `json:"type" firestore:"type"`
		Description string    `json:"description" firestore:"description"`
		CreatedAt   time.Time `json:"created_at" firestore:"created_at"`
	}
)

func (ActivityLog) TableName() string {
	return "activity_logs"
}

func (ActivityLog) AllowedFilterFields() []string {
	return []string{"id", "user_id", "type", "description", "created_at"}
}

func (ActivityLog) AllowedSortFields() []string {
	return []string{"id", "user_id", "type", "description", "created_at"}
}
