package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transactions struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AccountID uuid.UUID `gorm:"type:uuid"`
	Amount    int64     `gorm:"not null"`
	// "TOPUP", "WITHDRAW", "TRANSFER_IN", "TRANSFER_OUT"
	Type        string     `gorm:"type:varchar(12);not null"`
	Description *string    `gorm:"type:text"`
	RelatedID   *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Account *Account `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
