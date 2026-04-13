package models

import (
	"myapp/utils"

	"golang.org/x/crypto/bcrypt"
)

// InitSeedData 初始化系统基础数据（角色、权限、默认账号）
func InitSeedData() {
	var count int64
	DB.Model(&Role{}).Count(&count)

	// 如果角色表里已经有数据了，说明不是第一次启动，直接跳过
	if count > 0 {
		return
	}

	utils.Logger.Info("🚀 系统第一次启动，正在播种初始数据...")

	// 1. 创建基础权限 (严格对应 rbac.go 里的 perm.Code)
	pCreate := Permission{Name: "发布文章", Code: "article:create"}
	pUpdate := Permission{Name: "更新文章", Code: "article:update"}
	pDelete := Permission{Name: "删除文章", Code: "article:delete"}
	DB.Create(&pCreate)
	DB.Create(&pUpdate)
	DB.Create(&pDelete)

	// 2. 创建基础角色 (严格对应 rbac.go 里的 role.Keyword == "admin")
	adminRole := Role{
		Name:        "超级管理员",
		Keyword:     "admin",
		Permissions: []Permission{pCreate, pUpdate, pDelete},
	}
	editorRole := Role{
		Name:    "内容编辑",
		Keyword: "editor",
		// 编辑只有发布和更新权限，没有删除权限
		Permissions: []Permission{pCreate, pUpdate},
	}
	DB.Create(&adminRole)
	DB.Create(&editorRole)

	// 3. 创建默认测试账号 (密码统一为: 123456)
	hashPwd, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	adminUser := User{
		Username: "admin",
		Password: string(hashPwd),
		Roles:    []Role{adminRole}, // 分配超级管理员角色
	}
	editorUser := User{
		Username: "editor",
		Password: string(hashPwd),
		Roles:    []Role{editorRole}, // 分配普通编辑角色
	}

	DB.Create(&adminUser)
	DB.Create(&editorUser)

	utils.Logger.Info("✅ 初始数据播种完成！默认账号: admin/123456 (超级管理员), editor/123456 (内容编辑)")
}
