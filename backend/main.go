package main

import (
	"log"
	"sql-audit/config"
	"sql-audit/database"
	"sql-audit/handlers"
	"sql-audit/middleware"
	"sql-audit/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := database.Init(cfg); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	if err := database.Migrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	r := gin.Default()

	authService := services.NewAuthService(database.DB)
	authHandler := handlers.NewAuthHandler(authService)
	adminHandler := handlers.NewAdminHandler(database.DB)
	workOrderHandler := handlers.NewWorkOrderHandler(database.DB)

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

	r.Run(":" + cfg.ServerPort)
}
