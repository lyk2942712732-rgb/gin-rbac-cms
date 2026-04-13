package controllers

import (
	"net/http"

	"myapp/middlewares" // 引入你的本地中间件包
	"myapp/models"      // 引入你的本地模型包

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"myapp/utils" // 引入日志工具包 (注意替换为你的真实模块名)

	"go.uber.org/zap" // 引入 zap 核心包
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 账号注册
// @Summary 用户注册
// @Description 接收前端传来的账号密码，使用 Bcrypt 加密后存入数据库
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param data body RegisterRequest true "注册信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/register [post]
func Register(c *gin.Context) {
	var input RegisterRequest
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
	//初始化角色
	user := models.User{Username: input.Username, Password: string(hashedPassword)}

	// 这里我们假设新注册的用户默认都是 "editor" (内容编辑) 角色
	var defaultRole models.Role
	// 去角色表里找 keyword 为 "editor" 的角色
	if err := models.DB.Where("keyword = ?", "editor").First(&defaultRole).Error; err == nil {
		// 如果找到了，就把这个角色分配给用户
		// GORM 在接下来 Create 的时候，会自动把这个关联写入到 user_roles 中间表里
		user.Roles = []models.Role{defaultRole}
	} else {
		// 严谨起见：如果连默认角色都没查到，可能数据库还没初始化
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统默认角色缺失，请联系管理员"})
		return
	}

	if err := models.DB.Create(&user).Error; err != nil {
		utils.Logger.Error("注册失败:用户名已存在或写入失败", zap.String("username", user.Username), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在或创建失败"})
		return
	}

	// 【常规情报】：记录新用户
	utils.Logger.Info("新用户注册成功", zap.String("username", user.Username))

	c.JSON(http.StatusOK, gin.H{"message": "注册成功，已自动分配基础权限"})
}

// Login 账号登录
// @Summary 用户登录
// @Description 验证账号密码，成功后返回 JWT Token
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param data body LoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/login [post]
func Login(c *gin.Context) {
	var input LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var user models.User
	// Where 方法用于构建一个查询语句
	// 这里我们查询 username 字段等于 input.Username 的记录。First 方法用于将第一个结果加载到 user 变量中。
	// 如果查询失败（例如没有找到匹配的记录），它会返回一个错误。
	if err := models.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		utils.Logger.Warn("登录失败：用户不存在", zap.String("attempt_username", input.Username))
		//记录不存在的用户名尝试登录
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}
	//这里查询成功了,刚才定义的user实例被成功填充了数据库中的用户数据,接下来我们需要验证用户输入的密码是否正确。
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		utils.Logger.Warn("登录失败：密码错误", zap.String("attempt_username", input.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 调用 middlewares 包里的生成 Token 方法
	token, err := middlewares.GenerateToken(user.ID)
	if err != nil {
		utils.Logger.Error("生成 Token 失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
		return
	}

	utils.Logger.Info("用户登录成功", zap.String("username", user.Username), zap.Uint("userID", user.ID))

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetProfile 获取当前登录用户信息
// @Summary 获取当前登录用户信息
// @Description 获取当前登录用户的基础资料
// @Tags 用户模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/profile [get]
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
