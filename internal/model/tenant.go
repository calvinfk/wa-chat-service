package model

import "github.com/google/uuid"

type Tenant struct {
	ID    uuid.UUID                 `gorm:"type:uuid;primaryKey;default:uuidv7()"`
	Name  string                    `gorm:";not null;unique"`
	WABAs []WhatsappBusinessAccount `gorm:"foreignKey:TenantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
