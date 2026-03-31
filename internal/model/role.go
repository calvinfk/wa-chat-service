package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type (
	// Represents the role data model for the application.
	// The ID is generated as a UUID.
	// The CreatedAt and UpdatedAt fields are automatically managed by GORM to track when records are created and updated.
	Role struct {
		ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuidv7()"`
		Name      string     `json:"name" gorm:"unique;not null"`
		CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
		UpdatedAt *time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	}
)

func (Role) TableName() string {
	return "roles"
}

func (Role) AllowedFilterFields() []string {
	return []string{"id", "name", "created_at", "updated_at"}
}

func (Role) AllowedSortFields() []string {
	return []string{"id", "name", "created_at", "updated_at"}
}

func (r Role) IsAdminRole() bool {
	return strings.ToLower(r.Name) == "admin"
}

func (r Role) IsAgentRole() bool {
	return strings.ToLower(r.Name) == "agent"
}

func (r Role) IsUserRole() bool {
	return strings.ToLower(r.Name) == "user"
}
