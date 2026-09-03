package models

import (
	"time"

	"gorm.io/gorm"
)

// Friendship represents a connection between two users
type Friendship struct {
	ID          uint           `gorm:"primaryKey"`
	RequesterID uint           `gorm:"not null;index"`
	AddresseeID uint           `gorm:"not null;index"`
	Status      string         `gorm:"type:varchar(20);not null;default:'pending'"` // 'pending', 'accepted', 'blocked'
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
