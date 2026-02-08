package db

import (
	"IM/internal/model"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "root:jkesh1024@tcp(43.131.41.101:3306)/im_system?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 自动迁移：根据 User 结构体自动创建表
	db.AutoMigrate(&model.User{})
	DB = db
}
