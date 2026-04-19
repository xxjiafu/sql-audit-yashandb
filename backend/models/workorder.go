package models

import (
	"time"
)

type WorkOrder struct {
	ID              int64      `json:"id" gorm:"primaryKey"`
	Title           string     `json:"title" gorm:"size:200;not null"`
	SQLContent      string     `json:"sql_content" gorm:"type:text"`
	FileURL         string     `json:"file_url" gorm:"size:500"`
	ScheduledTime   *time.Time `json:"scheduled_time"`
	CreatorID       int64      `json:"creator_id" gorm:"index"`
	TargetDBID      int64      `json:"target_db_id" gorm:"index"`
	ExecutionUser   string     `json:"execution_user" gorm:"size:100"`
	Status          string     `json:"status" gorm:"size:20;default:pending"`
	RejectReason    string     `json:"reject_reason" gorm:"size:500"`
	AutoCheckResult string     `json:"auto_check_result" gorm:"type:json"`
	ExecutedAt      *time.Time `json:"executed_at"`
	ExecutionResult string     `json:"execution_result" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

const (
	StatusPending        = "pending"
	StatusAutoRejected   = "auto_rejected"
	StatusLeaderRejected = "leader_rejected"
	StatusDBARejected    = "dba_rejected"
	StatusLeaderApproved = "leader_approved"
	StatusDBAAproved     = "dba_approved"
	StatusExecuting      = "executing"
	StatusExecuted       = "executed"
	StatusFailed         = "failed"
)
