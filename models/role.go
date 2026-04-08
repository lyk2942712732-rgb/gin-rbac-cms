package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model        // gorm.Model 包含 ID、CreatedAt、UpdatedAt、DeletedAt 字段
	Name       string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`    // 例如: "管理员", "编辑"
	Keyword    string `gorm:"type:varchar(50);uniqueIndex;not null" json:"keyword"` // 例如: "admin", "editor"
	// 定义角色与权限的多对多关系
	//自动建出 role_permissions 中间表
	//因为逆向低频，不需要再 Role 结构体里定义 Users 字段了,其余多对多关系也同理
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}
