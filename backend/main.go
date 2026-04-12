package main

import (
	"log"
	"sql-audit/config"
	"sql-audit/database"

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
	r.Run(":" + cfg.ServerPort)
}
