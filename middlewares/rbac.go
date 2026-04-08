package middlewares

import (
	"myapp/models" // 替换为你的模块名
	"net/http"

	"github.com/gin-gonic/gin"
)

// CheckPermission 是一个闭包，传入需要的权限代码（如 "article:delete"）
func CheckPermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 JWT 中间件获取当前用户 ID
		uid, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
			c.Abort()
			return
		}

		var user models.User
		// 2. 【核心魔法】嵌套预加载：一次性查出用户 -> 用户的角色 -> 角色的权限
		if err := models.DB.Preload("Roles.Permissions").First(&user, uid).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户不存在"})
			c.Abort()
			return
		}

		// 3. 遍历用户的角色和权限，进行比对
		hasPermission := false
		for _, role := range user.Roles {
			// 保留你之前的优秀思路：如果是超级管理员，直接拥有所有权限放行
			if role.Keyword == "admin" {
				hasPermission = true
				break
			}

			// 如果不是超级管理员，就挨个检查权限代码是否匹配
			for _, perm := range role.Permissions {
				if perm.Code == requiredPermission {
					hasPermission = true
					break
				}
			}
		}

		// 4. 判断结果
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "越权操作：您没有 [" + requiredPermission + "] 的权限",
			})
			c.Abort() // 拦截请求，不再往下执行 Controller 的业务代码
			return
		}

		// 权限校验通过，放行
		c.Next()
	}
}
