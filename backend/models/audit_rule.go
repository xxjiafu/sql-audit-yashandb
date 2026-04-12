package models

import (
	"time"
)

type AuditRule struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	RuleType  string    `json:"rule_type" gorm:"size:20;not null"`
	Pattern   string    `json:"pattern" gorm:"type:text;not null"`
	Message   string    `json:"message" gorm:"size:500"`
	Severity  string    `json:"severity" gorm:"size:10;default:error"`
	IsEnabled bool      `json:"is_enabled" gorm:"default:true"`
	IsCustom  bool      `json:"is_custom" gorm:"default:false"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	RuleTypeSyntax      = "syntax"
	RuleTypePerformance = "performance"
	RuleTypeSecurity    = "security"
	RuleTypeConvention  = "convention"

	SeverityError   = "error"
	SeverityWarning = "warning"
)
