package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Account struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID        uuid.UUID `gorm:"type:uuid;not null"`
	AccountNumber string    `gorm:"not null;unique"`
	Balance       int64     `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	User *User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
