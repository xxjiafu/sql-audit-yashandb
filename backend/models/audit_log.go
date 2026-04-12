package models

import (
	"time"
)

type AuditLog struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	UserID     int64     `json:"user_id" gorm:"index"`
	Action     string    `json:"action" gorm:"size:50;not null"`
	TargetType string    `json:"target_type" gorm:"size:20"`
	TargetID   int64     `json:"target_id"`
	Detail     string    `json:"detail" gorm:"type:json"`
	IPAddress  string    `json:"ip_address" gorm:"size:50"`
	CreatedAt  time.Time `json:"created_at"`
}
