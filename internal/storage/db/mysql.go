package db

import (
	"IM/internal/config"
	"IM/internal/model"
	"errors"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg config.DatabaseConfig) error {
	if strings.TrimSpace(cfg.DSN) == "" {
		return errors.New("mysql dsn is required")
	}

	database, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := database.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		return err
	}

	if err := database.AutoMigrate(&model.User{}, &model.Message{}); err != nil {
		return err
	}

	DB = database
	return nil
}
