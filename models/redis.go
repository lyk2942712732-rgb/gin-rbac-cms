package models

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

// RDB 是全局的 Redis 客户端
var RDB *redis.Client

// Ctx 是 Redis 操作必须的上下文
var Ctx = context.Background()

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379" // 本地开发时的默认地址
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // 默认没密码
		DB:       0,  // 使用默认的 0 号数据库
	})

	// 测试连接
	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		panic("连接 Redis 失败: " + err.Error())
	}
	fmt.Println("Redis 缓存引擎初始化成功！")
}
