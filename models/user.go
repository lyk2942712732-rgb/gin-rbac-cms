package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password string `gorm:"not null"`
	// 增加这一行，定义用户与角色的多对多关系,
	Roles []Role `gorm:"many2many:user_roles;" json:"roles"`
}
