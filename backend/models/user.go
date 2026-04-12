package models

import (
	"time"
)

type User struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Username    string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password    string    `json:"-" gorm:"size:255;not null"`
	Role        string    `json:"role" gorm:"size:20;default:developer"`
	ApplyRole   string    `json:"apply_role" gorm:"size:20"`
	ApplyStatus string    `json:"apply_status" gorm:"size:20"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	RoleDeveloper = "developer"
	RoleLeader    = "leader"
	RoleDBA       = "dba"
	RoleAdmin     = "admin"

	ApplyStatusPending  = "pending"
	ApplyStatusApproved = "approved"
	ApplyStatusRejected = "rejected"
)
