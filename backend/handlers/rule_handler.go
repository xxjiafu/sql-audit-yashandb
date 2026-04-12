package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"sql-audit/models"
	"strconv"
)

type RuleHandler struct {
	db *gorm.DB
}

func NewRuleHandler(db *gorm.DB) *RuleHandler {
	return &RuleHandler{db: db}
}

type CreateRuleRequest struct {
	Name      string `json:"name" binding:"required"`
	RuleType  string `json:"rule_type" binding:"required"`
	Pattern   string `json:"pattern" binding:"required"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	IsEnabled bool   `json:"is_enabled"`
}

func (h *RuleHandler) List(c *gin.Context) {
	var rules []models.AuditRule
	h.db.Find(&rules)
	c.JSON(http.StatusOK, rules)
}

func (h *RuleHandler) Create(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Severity == "" {
		req.Severity = models.SeverityError
	}

	rule := &models.AuditRule{
		Name:      req.Name,
		RuleType:  req.RuleType,
		Pattern:   req.Pattern,
		Message:   req.Message,
		Severity:  req.Severity,
		IsEnabled: req.IsEnabled,
		IsCustom:  true,
	}

	result := h.db.Create(rule)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "规则创建成功", "id": rule.ID})
}

func (h *RuleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	result := h.db.Model(&models.AuditRule{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":       req.Name,
		"rule_type":  req.RuleType,
		"pattern":    req.Pattern,
		"message":    req.Message,
		"severity":   req.Severity,
		"is_enabled": req.IsEnabled,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "规则更新成功"})
}

func (h *RuleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	result := h.db.Delete(&models.AuditRule{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "规则删除成功"})
}
