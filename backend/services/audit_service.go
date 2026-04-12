package services

import (
	"encoding/json"
	"gorm.io/gorm"
	"regexp"
	"sql-audit/models"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

type AuditResult struct {
	Passed   bool     `json:"passed"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

func (s *AuditService) Audit(sqlContent string) (*AuditResult, error) {
	result := &AuditResult{Passed: true}

	var rules []models.AuditRule
	s.db.Where("is_enabled = ?", true).Find(&rules)

	for _, rule := range rules {
		matched, err := regexp.MatchString(rule.Pattern, sqlContent)
		if err != nil {
			continue
		}

		if matched {
			if rule.Severity == models.SeverityError {
				result.Passed = false
				result.Errors = append(result.Errors, rule.Message)
			} else {
				result.Warnings = append(result.Warnings, rule.Message)
			}
		}
	}

	return result, nil
}

func (s *AuditService) AuditAndSave(workOrder *models.WorkOrder) error {
	if workOrder.SQLContent == "" {
		return nil
	}

	auditResult, err := s.Audit(workOrder.SQLContent)
	if err != nil {
		return err
	}

	resultJSON, _ := json.Marshal(auditResult)
	workOrder.AutoCheckResult = string(resultJSON)

	if !auditResult.Passed {
		workOrder.Status = models.StatusAutoRejected
	}

	return nil
}
