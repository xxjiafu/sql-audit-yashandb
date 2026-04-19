package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sql-audit/models"
	"sql-audit/services"

	yasdb "git.yasdb.com/go/gorm-yasdb"
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

	c.JSON(http.StatusOK, gin.H{"message": "角色更新成功"})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&models.User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *AdminHandler) UpdateUserPassword(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，密码至少6个字符"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	result := h.db.Model(&models.User{}).Where("id = ?", userID).Update("password", string(hashedPassword))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
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

	// 填充目标数据库信息
	type WorkOrderWithDB struct {
		models.WorkOrder
		TargetDB string `json:"target_db"`
	}
	var result []WorkOrderWithDB
	for _, wo := range workOrders {
		tdb := ""
		if wo.TargetDBID > 0 {
			var dbConfig DBConfig
			if err := h.db.First(&dbConfig, wo.TargetDBID).Error; err == nil {
				tdb = dbConfig.Name + "(" + dbConfig.Host + ":" + fmt.Sprint(dbConfig.Port) + ")"
			}
		}
		result = append(result, WorkOrderWithDB{WorkOrder: wo, TargetDB: tdb})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AdminHandler) GetWorkOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var workOrder models.WorkOrder
	if err := h.db.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	// SQL内容区域只显示用户填写的内容
	// 即使SQL为空也不读取文件显示，文件只提供下载

	type WorkOrderWithDB struct {
		models.WorkOrder
		TargetDB string `json:"target_db"`
	}
	tdb := ""
	if workOrder.TargetDBID > 0 {
		var dbConfig DBConfig
		if err := h.db.First(&dbConfig, workOrder.TargetDBID).Error; err == nil {
			tdb = dbConfig.Name + "(" + dbConfig.Host + ":" + fmt.Sprint(dbConfig.Port) + ")"
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": WorkOrderWithDB{WorkOrder: workOrder, TargetDB: tdb}})
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

	role := services.GetUserRole(c)
	// 只有DBA和管理员可以执行SQL
	if role != models.RoleDBA && role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有DBA才能执行SQL"})
		c.Abort()
		return
	}

	var workOrder models.WorkOrder
	if err := h.db.First(&workOrder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
		return
	}

	if workOrder.Status == models.StatusExecuted || workOrder.Status == models.StatusExecuting {
		c.JSON(http.StatusBadRequest, gin.H{"error": "工单已执行成功或正在执行中，不能重复执行"})
		return
	}

	// 真正执行SQL - 在配置的数据库上执行
	sqlContent := workOrder.SQLContent

	if sqlContent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SQL内容为空，请输入SQL"})
		return
	}

	// 获取工单指定的目标数据库配置
	var dbConfig DBConfig
	targetDBID := workOrder.TargetDBID
	if targetDBID == 0 {
		// 如果没有指定，使用激活的数据库配置
		if err := h.db.Where("is_active = ?", true).First(&dbConfig).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置目标数据库"})
			return
		}
	} else {
		if err := h.db.First(&dbConfig, targetDBID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "目标数据库配置不存在"})
			return
		}
	}

	// 如果工单指定了执行用户，覆盖数据库配置中的用户名
	if workOrder.ExecutionUser != "" {
		dbConfig.Username = workOrder.ExecutionUser
	}

	// 连接目标数据库并执行
	targetDB, err := gorm.Open(yasdb.Open(dbConfig.BuildDSN()), &gorm.Config{})
	if err != nil {
		workOrder.Status = models.StatusFailed
		workOrder.ExecutionResult = "连接目标数据库失败: " + err.Error()
		h.db.Save(&workOrder)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "连接目标数据库失败: " + err.Error()})
		return
	}

	// 将SQL按分号分割成多条语句逐条执行（YashanDB不支持一次执行多条语句）
	statements := splitSQL(sqlContent)
	var totalRowsAffected int64

	for _, stmt := range statements {
		if stmt == "" {
			continue
		}
		result := targetDB.Exec(stmt)
		if result.Error != nil {
			workOrder.Status = models.StatusFailed
			workOrder.ExecutionResult = fmt.Sprintf("执行失败: %s\n在语句: %s", result.Error.Error(), stmt)
			h.db.Save(&workOrder)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "SQL执行失败", "detail": result.Error.Error(), "statement": stmt})
			return
		}
		totalRowsAffected += result.RowsAffected
	}

	if totalRowsAffected == 0 && isDDLStatement(sqlContent) {
		totalRowsAffected = 1
	}

	workOrder.Status = models.StatusExecuted
	now := time.Now()
	workOrder.ExecutedAt = &now
	workOrder.ExecutionResult = fmt.Sprintf("执行成功，影响行数: %d", totalRowsAffected)

	h.db.Save(&workOrder)

	c.JSON(http.StatusOK, gin.H{"message": "执行成功", "rows_affected": totalRowsAffected})
}

// ExecuteScheduled 自动执行预约工单（供定时任务调用）
func (h *AdminHandler) ExecuteScheduled(workOrder *models.WorkOrder) error {
	// 已经预约时间检查过了，直接执行
	sqlContent := workOrder.SQLContent

	if sqlContent == "" {
		err := fmt.Errorf("SQL内容为空，请输入SQL")
		workOrder.Status = models.StatusFailed
		workOrder.ExecutionResult = err.Error()
		h.db.Save(workOrder)
		return err
	}

	// 获取工单指定的目标数据库配置
	var dbConfig DBConfig
	targetDBID := workOrder.TargetDBID
	if targetDBID == 0 {
		// 如果没有指定，使用激活的数据库配置
		if err := h.db.Where("is_active = ?", true).First(&dbConfig).Error; err != nil {
			err := fmt.Errorf("请先配置目标数据库")
			workOrder.Status = models.StatusFailed
			workOrder.ExecutionResult = err.Error()
			h.db.Save(workOrder)
			return err
		}
	} else {
		if err := h.db.First(&dbConfig, targetDBID).Error; err != nil {
			err := fmt.Errorf("目标数据库配置不存在: %v", err)
			workOrder.Status = models.StatusFailed
			workOrder.ExecutionResult = err.Error()
			h.db.Save(workOrder)
			return err
		}
	}

	// 如果工单指定了执行用户，覆盖数据库配置中的用户名
	if workOrder.ExecutionUser != "" {
		dbConfig.Username = workOrder.ExecutionUser
	}

	// 连接目标数据库并执行
	targetDB, err := gorm.Open(yasdb.Open(dbConfig.BuildDSN()), &gorm.Config{})
	if err != nil {
		workOrder.Status = models.StatusFailed
		workOrder.ExecutionResult = fmt.Sprintf("连接目标数据库失败: %v", err)
		h.db.Save(workOrder)
		return err
	}

	// 将SQL按分号分割成多条语句逐条执行（YashanDB不支持一次执行多条语句）
	statements := splitSQL(sqlContent)
	var totalRowsAffected int64

	for _, stmt := range statements {
		if stmt == "" {
			continue
		}
		result := targetDB.Exec(stmt)
		if result.Error != nil {
			workOrder.Status = models.StatusFailed
			workOrder.ExecutionResult = fmt.Sprintf("执行失败: %v\n在语句: %s", result.Error, stmt)
			h.db.Save(workOrder)
			return result.Error
		}
		totalRowsAffected += result.RowsAffected
	}

	workOrder.Status = models.StatusExecuted
	now := time.Now()
	workOrder.ExecutedAt = &now
	workOrder.ExecutionResult = fmt.Sprintf("执行成功，影响行数: %d", totalRowsAffected)

	h.db.Save(workOrder)
	return nil
}

type DBConfig struct {
	ID       int64  `json:"id"`
	Host     string `json:"host" gorm:"size:200"`
	Port     int    `json:"port"`
	Instance string `json:"instance" gorm:"size:100;default:''"`
	Username string `json:"username" gorm:"size:100"`
	Password string `json:"password" gorm:"size:200"`
	Name     string `json:"name" gorm:"size:100"`
	IsActive bool   `json:"is_active" gorm:"default:false"`
}

func (c *DBConfig) BuildDSN() string {
	// YashanDB驱动DSN格式：用户名:密码@主机:端口?参数
	return fmt.Sprintf("%s/%s@%s:%d?autocommit=true&number_as_string=true", c.Username, c.Password, c.Host, c.Port)
}

var targetDB *gorm.DB

func (h *AdminHandler) SaveDBConfig(c *gin.Context) {
	var req DBConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.ID == 0 {
		// 新建默认激活，不禁用其他配置
		req.IsActive = true
		h.db.Create(&req)
	} else {
		h.db.Save(&req)
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置保存成功"})
}

func (h *AdminHandler) GetDBConfig(c *gin.Context) {
	var configs []DBConfig
	h.db.Find(&configs)
	c.JSON(http.StatusOK, configs)
}

func (h *AdminHandler) DeleteDBConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Delete(&DBConfig{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *AdminHandler) ScheduleWorkOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	role := services.GetUserRole(c)
	// 只有DBA和管理员可以预约执行
	if role != models.RoleDBA && role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有DBA才能预约执行SQL"})
		c.Abort()
		return
	}

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

func (h *AdminHandler) DeleteWorkOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&models.WorkOrder{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// UploadFile 上传SQL文件
// splitSQL 将SQL内容按分号分割成多条语句
// 忽略空语句，处理换行
func splitSQL(sql string) []string {
	var statements []string
	var current strings.Builder
	inQuotes := false

	for _, r := range sql {
		switch r {
		case ';':
			if !inQuotes {
				stmt := strings.TrimSpace(current.String())
				if stmt != "" {
					statements = append(statements, stmt)
				}
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		case '\'', '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		default:
			current.WriteRune(r)
		}
	}

	// 处理最后一条语句（如果没有分号结尾）
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

func isDDLStatement(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	ddlKeywords := []string{"CREATE", "ALTER", "DROP", "TRUNCATE"}
	for _, kw := range ddlKeywords {
		if strings.HasPrefix(upper, kw) {
			return true
		}
	}
	return false
}
