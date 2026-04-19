package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"sql-audit/models"
	"sql-audit/services"
	"strconv"
	"time"
)

type WorkOrderHandler struct {
	db *gorm.DB
}

func NewWorkOrderHandler(db *gorm.DB) *WorkOrderHandler {
	return &WorkOrderHandler{db: db}
}

type CreateWorkOrderRequest struct {
	Title         string  `json:"title" binding:"required"`
	SQLContent    string  `json:"sql_content"`
	ScheduledTime *string `json:"scheduled_time"`
	TargetDBID    int64   `json:"target_db_id"`
	ExecutionUser string  `json:"execution_user"`
}

func (h *WorkOrderHandler) Create(c *gin.Context) {
	userID := services.GetUserID(c)

	var req CreateWorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.SQLContent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SQL内容不能为空"})
		return
	}

workOrder := &models.WorkOrder{
		Title:         req.Title,
		SQLContent:    req.SQLContent,
		CreatorID:     userID,
		TargetDBID:    req.TargetDBID,
		ExecutionUser: req.ExecutionUser,
		Status:       models.StatusPending,
	}

	if req.ScheduledTime != nil {
		if t, err := time.Parse("2006-01-02 15:04:05", *req.ScheduledTime); err == nil {
			workOrder.ScheduledTime = &t
		}
	}

	result := h.db.Create(workOrder)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	auditService := services.NewAuditService(h.db)
	if err := auditService.AuditAndSave(workOrder); err != nil {
		return
	}
	h.db.Save(workOrder)

	c.JSON(http.StatusCreated, gin.H{"message": "工单创建成功", "id": workOrder.ID})
}

func (h *WorkOrderHandler) List(c *gin.Context) {
	userID := services.GetUserID(c)
	role := services.GetUserRole(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var workOrders []models.WorkOrder
	var total int64

	query := h.db.Model(&models.WorkOrder{})

	if role != models.RoleDBA && role != models.RoleAdmin {
		query = query.Where("creator_id = ?", userID)
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&workOrders)

	c.JSON(http.StatusOK, gin.H{
		"data":      workOrders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *WorkOrderHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := services.GetUserID(c)
	role := services.GetUserRole(c)

	var workOrder models.WorkOrder
	result := h.db.First(&workOrder, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	if role != models.RoleDBA && role != models.RoleAdmin && workOrder.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限查看此工单"})
		return
	}

	// SQL内容区域只显示用户填写的内容
	// 即使SQL为空也不读取文件显示，文件只提供下载

	c.JSON(http.StatusOK, workOrder)
}

func (h *WorkOrderHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := services.GetUserID(c)

	var workOrder models.WorkOrder
	result := h.db.First(&workOrder, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	role := services.GetUserRole(c)
	// 创建者本人、管理员、DBA都可以修改
	if workOrder.CreatorID != userID && role != models.RoleAdmin && role != models.RoleDBA {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限修改此工单"})
		return
	}

	// 只要不是已经执行成功，都允许修改
	if workOrder.Status == models.StatusExecuted {
		c.JSON(http.StatusForbidden, gin.H{"error": "已执行成功的工单不能修改"})
		return
	}

	var req CreateWorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	updates := map[string]interface{}{
		"title":          req.Title,
		"sql_content":    req.SQLContent,
		"target_db_id":   req.TargetDBID,
		"execution_user": req.ExecutionUser,
	}

	if req.ScheduledTime != nil {
		if t, err := time.Parse("2006-01-02 15:04:05", *req.ScheduledTime); err == nil {
			updates["scheduled_time"] = t
		}
	}

	h.db.Model(&workOrder).Updates(updates)

	c.JSON(http.StatusOK, gin.H{"message": "工单更新成功"})
}

func (h *WorkOrderHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := services.GetUserID(c)

	var workOrder models.WorkOrder
	result := h.db.First(&workOrder, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	if workOrder.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除此工单"})
		return
	}

	if workOrder.Status != models.StatusPending {
		c.JSON(http.StatusForbidden, gin.H{"error": "只能删除待审核的工单"})
		return
	}

	h.db.Delete(&workOrder)

	c.JSON(http.StatusOK, gin.H{"message": "工单删除成功"})
}
