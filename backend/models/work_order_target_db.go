package models

import "time"

type WorkOrderTargetDB struct {
	ID              int64      `json:"id" gorm:"primaryKey"`
	WorkOrderID     int64      `json:"work_order_id" gorm:"index"`
	DBConfigID      int64      `json:"db_config_id" gorm:"index"`
	DBHost          string     `json:"db_host" gorm:"size:200"`
	DBName          string     `json:"db_name" gorm:"size:100"`
	Status          string     `json:"status" gorm:"size:20;default:pending"`
	ExecutionResult string     `json:"execution_result" gorm:"type:text"`
	ExecutedAt      *time.Time `json:"executed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

const (
	TargetDBStatusPending    = "pending"
	TargetDBStatusExecuting  = "executing"
	TargetDBStatusSuccess    = "success"
	TargetDBStatusFailed     = "failed"
)