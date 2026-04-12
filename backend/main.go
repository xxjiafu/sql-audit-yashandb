package main

import (
	"log"
	"sql-audit/config"
	"sql-audit/database"
	"sql-audit/handlers"
	"sql-audit/middleware"
	"sql-audit/models"
	"sql-audit/services"

	"github.com/gin-gonic/gin"
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

	SeedDefaultRules(database.DB)

	r := gin.Default()

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
	{
		admin.POST("/users", adminHandler.CreateUser)
		admin.GET("/users", adminHandler.ListUsers)
		admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
		admin.PUT("/users/:id/apply", adminHandler.HandleApply)
	}

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

	r.Run(":" + cfg.ServerPort)
}

func SeedDefaultRules(db *gorm.DB) {
	rules := []models.AuditRule{
		{Name: "禁止SELECT *", RuleType: models.RuleTypeSyntax, Pattern: `(?i)SELECT\s+\*`, Message: "不建议使用 SELECT *", Severity: models.SeverityWarning, IsEnabled: true},
		{Name: "必须使用LIMIT", RuleType: models.RuleTypeSyntax, Pattern: `(?i)SELECT\s+(?!.*LIMIT)`, Message: "SELECT 语句建议使用 LIMIT 限制结果集", Severity: models.SeverityWarning, IsEnabled: true},
		{Name: "禁止全表更新", RuleType: models.RuleTypePerformance, Pattern: `(?i)UPDATE\s+\w+\s+SET\s+\w+\s*=`, Message: "UPDATE 语句建议添加 WHERE 条件", Severity: models.SeverityWarning, IsEnabled: true},
		{Name: "避免LIKE通配符开头", RuleType: models.RuleTypePerformance, Pattern: `(?i)WHERE\s+\w+\s+LIKE\s+'%`, Message: "LIKE 以 % 开头无法使用索引", Severity: models.SeverityWarning, IsEnabled: true},
		{Name: "禁止明文密码", RuleType: models.RuleTypeSecurity, Pattern: `(?i)password\s*=\s*['"']`, Message: "禁止明文存储密码", Severity: models.SeverityError, IsEnabled: true},
	}
	for _, rule := range rules {
		db.Create(&rule)
	}
}
