package model

import (
	"time"

	"github.com/google/uuid"
)

type (
	// Represents the  user data model for the application.
	// RoleID is a foreign key that references the Role model.
	// The PasswordHash field is not included in JSON responses for security reasons, and the ID is generated as a UUID.
	// The CreatedAt and UpdatedAt fields are automatically managed by GORM to track when records are created and updated.
	User struct {
		ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuidv7()"`
		RoleID       uuid.UUID  `json:"role_id" gorm:"type:uuid;not null"`
		Role         Role       `json:"role" gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE"`
		Name         string     `json:"name" gorm:"not null"`
		Email        string     `json:"email" gorm:"unique;not null"`
		PasswordHash string     `json:"-" gorm:"not null"`
		IsActive     int        `json:"is_active" gorm:"not null;default:1"` // 1 for active, 0 for inactive
		CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
		UpdatedAt    *time.Time `json:"updated_at" gorm:"autoUpdateTime"`
		DeletedAt    *time.Time `json:"deleted_at" gorm:"index"`
	}
)

func (User) AllowedFilterFields() []string {
	return []string{"id", "role_id", "name", "email", "is_active", "created_at", "updated_at", "deleted_at"}
}

func (User) AllowedSortFields() []string {
	return []string{"id", "role_id", "name", "email", "is_active", "created_at", "updated_at", "deleted_at"}
}
