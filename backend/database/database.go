package database

import (
	"fmt"
	"sql-audit/config"
	"sql-audit/models"

	yasdb "git.yasdb.com/go/gorm-yasdb"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *config.Config) error {
	dsn := fmt.Sprintf("sqlaudit/sqlaudit@192.168.23.87:1688")

	db, err := gorm.Open(yasdb.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	DB = db
	return nil
}

func Migrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.WorkOrder{},
		&models.AuditRule{},
		&models.AuditLog{},
	)
}
