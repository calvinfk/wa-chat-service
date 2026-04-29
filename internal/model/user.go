package model

import (
	"time"
)

type (
	// Represents the  user data model for the application.
	// Role is represented as a string, there's 3 roles: admin, supervisor, agent.
	// The PasswordHash field is not included in JSON responses for security reasons, and the ID is generated as a UUID.
	// The CreatedAt and UpdatedAt fields are automatically managed by GORM to track when records are created and updated.
	User struct {
		DocumentID   string    `json:"id" firestore:"-"`
		TenantID     string    `json:"tenant_id" firestore:"tenant_id"`                             // foreign key to tenant
		SupervisorID *string   `json:"supervisor_id,omitempty" firestore:"supervisor_id,omitempty"` // foreign key to supervisor, nullable
		Role         string    `json:"role" firestore:"role"`
		Name         string    `json:"name" firestore:"name"`
		Email        string    `json:"email" firestore:"email"`
		Password     string    `json:"password" firestore:"password"` // stored as a hash in the database, do not include in JSON responses
		CreatedAt    time.Time `json:"created_at" firestore:"created_at"`
	}
)

func (User) TableName() string {
	return "users"
}

func (User) AllowedFilterFields() []string {
	return []string{"id", "tenant_id", "role", "email", "supervisor_id"}
}

func (User) AllowedSortFields() []string {
	return []string{"created_at"}
}
