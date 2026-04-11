package controllers

import (
	"encoding/json"
	"fmt"
	"myapp/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type ArticleRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// CreateArticle 发布文章
// @Summary 发布新文章
// @Description 当前登录用户发布一篇新文章
// @Tags 文章模块
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body CreateArticleRequest true "文章内容"
// @Success 200 {object} map[string]interface{}
// @Router /api/articles [post]
func CreateArticle(c *gin.Context) {
	var input ArticleRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 从 JWT 中间件存入的上下文中获取当前登录的 UserID
	uid, _ := c.Get("userID")

	article := models.Article{
		Title:   input.Title,
		Content: input.Content,
		UserID:  uid.(uint), // 类型断言为 uint
	}

	if err := models.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "发布失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "发布成功"})
}

// GetArticles 获取文章列表 (重点：分页与搜索)
// GetArticles 获取文章列表
// @Summary 分页获取文章
// @Description 分页获取系统的文章数据，支持标题模糊搜索
// @Tags 文章模块
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1"
// @Param size query int false "每页数量，默认10"
// @Param title query string false "文章标题"
// @Success 200 {object} map[string]interface{}
// @Router /api/articles [get]
func GetArticles(c *gin.Context) {
	// 1. 获取分页参数 (从 URL 参数拿，如 ?page=1&size=10)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1")) //strconv.Atoi 函数用于将字符串转换为整数。
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	title := c.Query("title") // 搜索关键词

	var articles []models.Article
	var total int64

	// 2. 构造查询构造器
	query := models.DB.Model(&models.Article{}) //SELECT * FROM articles

	// 3. 如果有搜索词，增加模糊查询
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
		//SELECT * FROM articles + WHERE title LIKE '%Go语言%'
	}

	// 4. 获取查询到的数据总数 (用于前端分页显示),赋值给 total 变量
	query.Count(&total)

	// 5. 执行分页查询
	// Offset: 跳过多少条数据，Limit: 限制返回多少条，达到分页效果。
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Preload("User").Find(&articles)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "查询成功",
		"data": gin.H{
			"list":  articles,
			"total": total,
		},
	})
}

// GetArticleDetail 获取文章详情
// @Summary 获取文章详情
// @Description 根据文章 ID 获取文章详情和作者信息
// @Tags 文章模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/articles/{id} [get]
// GetArticleDetail 获取单篇文章详情 (带 Redis 缓存)
func GetArticleDetail(c *gin.Context) {
	id := c.Param("id")

	// 1. 定义缓存的 Key，比如 "article:1"
	cacheKey := "article:" + id

	// 2. 【第一步】先去 Redis找数据
	val, err := models.RDB.Get(models.Ctx, cacheKey).Result()
	if err == nil {
		// 命中缓存,直接把 Redis 里的 JSON 字符串返回给前端

		var article models.Article
		json.Unmarshal([]byte(val), &article) //反序列化后结果给 article 变量

		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "查询成功(来自缓存)", "data": article})
		return
	} else if err != redis.Nil {
		// Redis 挂了或者发生了其他错误，除了缓存未命中以外的错误
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "缓存服务异常"})
		return
	}

	// 3. 【第二步】如果 Redis 里没有 (缓存未命中)，去 MySQL 查
	var article models.Article
	if err := models.DB.Preload("User").First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "文章未找到"})
		return
	}

	// 4. 【第三步】查到之后，把它序列化成 JSON，存到Redis，方便下一个人查
	articleJSON, _ := json.Marshal(article)
	// 设置 1 小时过期时间 (热点数据缓存策略)
	models.RDB.Set(models.Ctx, cacheKey, articleJSON, time.Hour) //context,key,value,expiration

	// 5. 返回给前端
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "查询成功(来自数据库)", "data": article})
}

// UpdateArticle 更新文章 (解决缓存一致性)
// @Summary 更新文章
// @Description 更新指定 ID 的文章，并自动清除 Redis 脏缓存
// @Tags 文章模块
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文章 ID"
// @Param data body UpdateArticleRequest true "更新内容"
// @Success 200 {object} map[string]interface{}
// @Router /api/articles/{id} [put]
func UpdateArticle(c *gin.Context) {
	id := c.Param("id")

	// 1. 校验数据格式
	var req ArticleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 2. 核心原则第一步：先更新 MySQL 数据库
	var article models.Article
	if err := models.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "文章不存在"})
		return
	}

	// 【安全防线】：防水平越权 (IDOR) - 确保当前登录的人是这篇文章的作者
	userID, _ := c.Get("userID")
	if article.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "越权操作：您只能修改自己的文章"})
		return
	}

	// 执行数据库更新
	models.DB.Model(&article).Updates(models.Article{
		Title:   req.Title,
		Content: req.Content,
	})

	// 3. 核心原则第二步：删除 Redis 里的旧缓存！
	// 拼接对应的 Cache Key
	cacheKey := "article:" + id
	err := models.RDB.Del(models.Ctx, cacheKey).Err()
	if err != nil {
		// 注意：真实生产环境中，如果删除缓存失败，通常会接入消息队列重试，
		// 这里为了简单，我们先记录日志（打印）
		fmt.Printf("警告: 删除缓存失败 %s: %v\n", cacheKey, err)
	}

	// 4. 返回成功
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功，缓存已刷新", "data": article})
}

// DeleteArticle 删除文章
// @Summary 删除文章
// @Description 根据文章 ID 删除文章，需具备 article:delete 权限
// @Tags 文章模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/articles/{id} [delete]
func DeleteArticle(c *gin.Context) {
	id := c.Param("id")

	// 根据 ID 查询文章
	var article models.Article
	if err := models.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "文章未找到"})
		return
	}

	//权限检查通过后删除文章
	if err := models.DB.Delete(&models.Article{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
