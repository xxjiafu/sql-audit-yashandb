package main

import (
	"log"
	"sql-audit/config"
	"sql-audit/database"
	"sql-audit/handlers"
	"sql-audit/middleware"
	"sql-audit/models"
	"sql-audit/services"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	if err := database.Init(cfg); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	if err := database.Migrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建 DB_CONFIG 表
	database.DB.AutoMigrate(&handlers.DBConfig{})

	SeedDefaultRules(database.DB)
	SeedAdminUser(database.DB)

	r := gin.Default()

	// 确保所有JSON响应使用UTF-8编码
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	authService := services.NewAuthService(database.DB)
	authHandler := handlers.NewAuthHandler(authService)
	adminHandler := handlers.NewAdminHandler(database.DB)
	workOrderHandler := handlers.NewWorkOrderHandler(database.DB)
	ruleHandler := handlers.NewRuleHandler(database.DB)

	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	authorized := r.Group("/api")
	authorized.Use(middleware.JWTAuth())
	{
		authorized.GET("/auth/me", authHandler.GetCurrentUser)
		authorized.POST("/auth/apply-role", authHandler.ApplyRole)
	}

	admin := r.Group("/api/admin")
	admin.Use(middleware.JWTAuth())
	admin.Use(middleware.RequireAdminRole())
	{
		admin.POST("/users", adminHandler.CreateUser)
		admin.GET("/users", adminHandler.ListUsers)
		admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
		admin.PUT("/users/:id/password", adminHandler.UpdateUserPassword)
		admin.DELETE("/users/:id", adminHandler.DeleteUser)
		admin.PUT("/users/:id/apply", adminHandler.HandleApply)
	}

	adminWorkOrders := r.Group("/api/admin/workorders")
	adminWorkOrders.Use(middleware.JWTAuth())
	adminWorkOrders.Use(middleware.RequireAdminRole())
	{
		adminWorkOrders.GET("", adminHandler.ListAllWorkOrders)
		adminWorkOrders.GET("/:id", adminHandler.GetWorkOrder)
		adminWorkOrders.PUT("/:id/leader-approve", adminHandler.LeaderApprove)
		adminWorkOrders.PUT("/:id/dba-approve", adminHandler.DBAApprove)
		adminWorkOrders.PUT("/:id/reject", adminHandler.RejectWorkOrder)
		adminWorkOrders.PUT("/:id/execute", adminHandler.ExecuteWorkOrder)
		adminWorkOrders.PUT("/:id/schedule", adminHandler.ScheduleWorkOrder)
		adminWorkOrders.DELETE("/:id", adminHandler.DeleteWorkOrder)
	}

	// 所有登录用户都可以获取数据库配置列表（创建工单时选择）
	authorized.GET("/admin/db-config", adminHandler.GetDBConfig)

	// 数据库配置管理（只有admin可以修改）
	admin.POST("/db-config", adminHandler.SaveDBConfig)
	admin.DELETE("/db-config/:id", adminHandler.DeleteDBConfig)

	workorders := r.Group("/api/workorders")
	workorders.Use(middleware.JWTAuth())
	{
		workorders.POST("", workOrderHandler.Create)
		workorders.GET("", workOrderHandler.List)
		workorders.GET("/:id", workOrderHandler.Get)
		workorders.PUT("/:id", workOrderHandler.Update)
		workorders.DELETE("/:id", workOrderHandler.Delete)
	}

	rules := r.Group("/api/rules")
	rules.Use(middleware.JWTAuth())
	{
		rules.GET("", ruleHandler.List)
		rules.POST("", ruleHandler.Create)
		rules.PUT("/:id", ruleHandler.Update)
		rules.DELETE("/:id", ruleHandler.Delete)
	}

	// 启动定时任务：每分钟检查一次预约执行的工单，到点自动执行
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			// 查找所有DBA已通过、有预约时间、且预约时间已到、尚未执行的工单
			var workOrders []models.WorkOrder
			now := time.Now()
			database.DB.Where(
				"status = ? AND scheduled_time IS NOT NULL AND scheduled_time <= ?",
				models.StatusDBAAproved, now,
			).Find(&workOrders)

			// 逐个执行
			h := adminHandler
			for _, wo := range workOrders {
				if err := h.ExecuteScheduled(&wo); err != nil {
					log.Printf("自动执行预约工单失败 #%d: %v", wo.ID, err)
				}
			}
		}
	}()

	r.Run(":" + cfg.ServerPort)
}

func SeedDefaultRules(db *gorm.DB) {
	rules := []models.AuditRule{
		{Name: "禁止SELECT *", RuleType: models.RuleTypeSyntax, Pattern: `(?i)SELECT\s+\*`, Message: "禁止使用 SELECT *", Severity: models.SeverityError, IsEnabled: true},
		{Name: "必须使用LIMIT", RuleType: models.RuleTypeSyntax, Pattern: `\bSELECT\b.*\bLIMIT\b`, InvertMatch: true, Message: "SELECT 语句建议使用 LIMIT 限制结果集", Severity: models.SeverityWarning, IsEnabled: true},
		{Name: "禁止全表更新", RuleType: models.RuleTypePerformance, Pattern: `(?i)UPDATE\s+\w+\s+SET\s+\w+\s*=`, Message: "UPDATE 语句建议添加 WHERE 条件", Severity: models.SeverityWarning, IsEnabled: true},
		{Name: "避免LIKE通配符开头", RuleType: models.RuleTypePerformance, Pattern: `(?i)WHERE\s+\w+\s+LIKE\s+'%`, Message: "LIKE 以 % 开头无法使用索引", Severity: models.SeverityWarning, IsEnabled: true},
		{Name: "禁止明文密码", RuleType: models.RuleTypeSecurity, Pattern: `(?i)password\s*=\s*['"']`, Message: "禁止明文存储密码", Severity: models.SeverityError, IsEnabled: true},
	}
	for _, rule := range rules {
		var count int64
		db.Model(&models.AuditRule{}).Where("name = ?", rule.Name).Count(&count)
		if count == 0 {
			db.Create(&rule)
		} else {
			db.Model(&models.AuditRule{}).Where("name = ?", rule.Name).Updates(map[string]interface{}{
				"severity":    rule.Severity,
				"pattern":    rule.Pattern,
				"message":   rule.Message,
				"is_enabled": true,
			})
		}
	}
}

func SeedAdminUser(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		db.Create(&models.User{
			Username: "admin",
			Password: string(hashedPassword),
			Role:     "admin",
		})
	} else {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		db.Model(&models.User{}).Where("username = ?", "admin").Updates(map[string]interface{}{"password": string(hashedPassword), "role": "admin"})
	}
	var dbaCount int64
	db.Model(&models.User{}).Where("username = ?", "dba").Count(&dbaCount)
	if dbaCount == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("dba123"), bcrypt.DefaultCost)
		db.Create(&models.User{
			Username: "dba",
			Password: string(hashedPassword),
			Role:     "dba",
		})
	} else {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("dba123"), bcrypt.DefaultCost)
		db.Model(&models.User{}).Where("username = ?", "dba").Updates(map[string]interface{}{"password": string(hashedPassword), "role": "dba"})
	}
	var leaderCount int64
	db.Model(&models.User{}).Where("username = ?", "leader").Count(&leaderCount)
	if leaderCount == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("leader123"), bcrypt.DefaultCost)
		db.Create(&models.User{
			Username: "leader",
			Password: string(hashedPassword),
			Role:     "leader",
		})
	} else {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("leader123"), bcrypt.DefaultCost)
		db.Model(&models.User{}).Where("username = ?", "leader").Updates(map[string]interface{}{"password": string(hashedPassword), "role": "leader"})
	}
}