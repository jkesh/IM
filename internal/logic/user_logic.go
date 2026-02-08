package logic

import (
	"IM/internal/model"
	"IM/internal/storage/db"

	"golang.org/x/crypto/bcrypt"
)

func Register(username, password string) error {

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := model.User{
		Username: username,
		Password: string(hashedPassword),
	}

	// 2. 存入数据库
	return db.DB.Create(&user).Error
}
