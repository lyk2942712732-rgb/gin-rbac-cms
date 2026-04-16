package models

import (
	"log"
	"myapp/utils"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB 是全局数据库实例，首字母大写以便其他包调用
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	// 从环境变量 DSN 读取，如果读不到（比如本地开发时），就用默认的 localhost
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "root:675563@tcp(127.0.0.1:3306)/my_gin_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	var err error
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {

		// gorm.Open 函数用于打开数据库连接，第一个参数是数据库驱动，这里我们使用 mysql.Open 来指定 MySQL 数据库，第二个参数是 gorm.Config 配置对象。
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

		if err == nil {
			utils.Logger.Info("✅ 数据库连接成功！正在自动迁移表结构...")

			DB.AutoMigrate(&User{}, &Article{}, &Role{}, &Permission{})
			// GORM 扫描到 many2many 标签时，会自动在数据库里建出 user_roles 和 role_permissions 这两张中间表！
			// AutoMigrate 方法会根据 User 结构体自动创建或更新数据库表结构，确保数据库中的表与代码中的模型保持一致。

			return
		}

		log.Printf("⚠️ 数据库连接失败，正在进行第 %d/%d 次重试... (等待 3 秒)", i, maxRetries)
		time.Sleep(3 * time.Second) // 挂起 3 秒再试

	}

	// 如果 10 次（30秒）都没连上,这时候再 panic
	panic("❌ 数据库连接彻底失败: " + err.Error())

}
