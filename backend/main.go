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

	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	authorized := r.Group("/api")
	authorized.Use(middleware.JWTAuth())
	{
		authorized.GET("/auth/me", authHandler.GetCurrentUser)
	}

	r.Run(":" + cfg.ServerPort)
}
