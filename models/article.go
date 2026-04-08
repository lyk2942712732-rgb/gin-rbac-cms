package models

import "gorm.io/gorm"

type Article struct {
	gorm.Model
	Title   string `gorm:"type:varchar(200);not null" json:"title"`
	Content string `gorm:"type:longtext" json:"content"`
	// UserID 是外键，关联 User 表的 ID
	UserID uint `json:"user_id"`
	// User 结构体用于 GORM 的 Preload（预加载）功能，方便查询文章时直接带出作者信息
	//使用gorm:"foreignKey:UserID" 来指定 UserID 是 Article 表中的外键，关联到 User 表的 ID 字段。
	User User `gorm:"foreignKey:UserID" json:"author"`
}
