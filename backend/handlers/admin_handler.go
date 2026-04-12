package handlers

import (
	"net/http"
	"strconv"
	"time"

	"sql-audit/models"
	"sql-audit/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Role == "" {
		req.Role = models.RoleDeveloper
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := &models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     req.Role,
	}

	result := h.db.Create(user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "用户创建成功", "user_id": user.ID})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	var users []models.User
	h.db.Find(&users)
	c.JSON(http.StatusOK, users)
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	result := h.db.Model(&models.User{}).Where("id = ?", userID).Update("role", req.Role)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色分配成功"})
}

func (h *AdminHandler) HandleApply(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Approved bool   `json:"approved"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	updates := map[string]interface{}{
		"apply_status": "",
	}

	if req.Approved {
		updates["role"] = req.Role
	}

	result := h.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审批完成"})
}

func (h *AdminHandler) ListAllWorkOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	title := c.Query("title")
	username := c.Query("username")

	var workOrders []models.WorkOrder
	var total int64

	query := h.db.Model(&models.WorkOrder{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if username != "" {
		query = query.Joins("LEFT JOIN users ON users.id = work_orders.creator_id").Where("users.username LIKE ?", "%"+username+"%")
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("work_orders.created_at DESC").Find(&workOrders)

	c.JSON(http.StatusOK, gin.H{
		"data":      workOrders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AdminHandler) LeaderApprove(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var workOrder models.WorkOrder
	if err := h.db.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	if workOrder.Status != models.StatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "工单状态不正确"})
		return
	}

	workOrder.Status = models.StatusLeaderApproved
	h.db.Save(&workOrder)

	c.JSON(http.StatusOK, gin.H{"message": "审核通过"})
}

func (h *AdminHandler) DBAApprove(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var workOrder models.WorkOrder
	if err := h.db.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	if workOrder.Status != models.StatusLeaderApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "工单状态不正确，需要先通过开发组长审核"})
		return
	}

	workOrder.Status = models.StatusDBAAproved
	h.db.Save(&workOrder)

	c.JSON(http.StatusOK, gin.H{"message": "DBA审核通过"})
}

func (h *AdminHandler) RejectWorkOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	role := services.GetUserRole(c)

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	var workOrder models.WorkOrder
	if err := h.db.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	if role == models.RoleLeader {
		workOrder.Status = models.StatusLeaderRejected
	} else {
		workOrder.Status = models.StatusDBARejected
	}
	workOrder.RejectReason = req.Reason

	h.db.Save(&workOrder)

	c.JSON(http.StatusOK, gin.H{"message": "已驳回"})
}

func (h *AdminHandler) ExecuteWorkOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var workOrder models.WorkOrder
	if err := h.db.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	if workOrder.Status != models.StatusDBAAproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "工单未通过DBA审核"})
		return
	}

	workOrder.Status = models.StatusExecuted
	now := time.Now()
	workOrder.ExecutedAt = &now
	workOrder.ExecutionResult = "SQL执行成功"

	h.db.Save(&workOrder)

	c.JSON(http.StatusOK, gin.H{"message": "执行成功"})
}

func (h *AdminHandler) ScheduleWorkOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		ScheduledTime string `json:"scheduled_time" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	t, err := time.Parse("2006-01-02 15:04:05", req.ScheduledTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "时间格式错误"})
		return
	}

	var workOrder models.WorkOrder
	if err := h.db.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	workOrder.ScheduledTime = &t
	h.db.Save(&workOrder)

	c.JSON(http.StatusOK, gin.H{"message": "预约成功"})
}
