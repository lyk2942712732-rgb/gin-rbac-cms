package controllers

import (
	"net/http"

	"myapp/middlewares" // 引入你的本地中间件包
	"myapp/models"      // 引入你的本地模型包

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// c *gin.Context 是 Gin 框架中用于处理 HTTP 请求和响应的上下文对象，包含了请求的所有信息以及用于构建响应的方法。
func Register(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		//结构体标签，告诉 Gin 这个字段是必填的，并且当结构体与json绑定时，会将 JSON 数据绑定到这个字段。
	}
	// ShouldBindJSON 方法会将请求体中的 JSON 数据绑定到 input 结构体中，如果绑定失败，它会返回一个错误。
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		// Gin 的 JSON 方法用于构建一个 JSON 响应，第一个参数是 HTTP 状态码，第二个参数是一个包含响应数据的 map。在这里，返回一个 400 Bad Request 错误，并且在响应体中包含一个 "error" 字段，值为 "参数错误"。
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	// bcrypt.GenerateFromPassword 函数用于将用户输入的密码进行哈希处理，生成一个安全的密码哈希。它接受两个参数：第一个是要哈希的密码（需要转换为字节切片），第二个是哈希的成本（bcrypt.DefaultCost 是一个预定义的常量，表示默认的哈希成本）。如果哈希过程失败，它会返回一个错误。
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	user := models.User{Username: input.Username, Password: string(hashedPassword)}
	// DB 是一个全局的数据库连接实例，已经在 models 包中初始化并连接到数据库。
	// 通过 models.DB，我们可以执行数据库操作，例如创建、查询、更新和删除记录。
	// 在这里，我们使用 models.DB.Create(&user) 来将新创建的用户记录保存到数据库中。
	if err := models.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在或创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

func Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var user models.User
	// Where 方法用于构建一个查询语句
	// 这里我们查询 username 字段等于 input.Username 的记录。First 方法用于将第一个结果加载到 user 变量中。
	// 如果查询失败（例如没有找到匹配的记录），它会返回一个错误。
	if err := models.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}
	//这里查询成功了,刚才定义的user实例被成功填充了数据库中的用户数据,接下来我们需要验证用户输入的密码是否正确。
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 调用 middlewares 包里的生成 Token 方法
	token, err := middlewares.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	//等效models.DB.Where("id = ?", userID).First(&user)
	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到用户信息"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "成功访问受保护的资源",
		"user_id":  user.ID,
		"username": user.Username,
	})
}
