package model

import (
	"time"

	"github.com/google/uuid"
)

type WhatsappBusinessAccount struct {
	ID           string        `gorm:"primaryKey"`
	TenantID     uuid.UUID     `gorm:"type:uuid;not null"`
	CreatedAt    time.Time     `gorm:"not null"`
	UpdatedAt    time.Time     `gorm:"not null"`
	PhoneNumbers []PhoneNumber `gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
