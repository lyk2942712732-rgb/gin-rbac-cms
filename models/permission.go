package models

import "gorm.io/gorm"

type Permission struct {
	gorm.Model
	Name string `gorm:"type:varchar(50);not null" json:"name"`             // 例如: "删除文章"
	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"` // 例如: "article:delete"
}
