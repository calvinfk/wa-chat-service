package model

type PhoneNumber struct {
	ID          string `gorm:"primaryKey"`
	AccountID   string `gorm:"not null;unique"`
	PhoneNumber string `gorm:"not null;unique"`
}
