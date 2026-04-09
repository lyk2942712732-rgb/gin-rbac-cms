package controllers

import (
	"myapp/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateArticle 发布文章
func CreateArticle(c *gin.Context) {
	var input struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
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

func GetArticleDetail(c *gin.Context) {
	id := c.Param("id")

	//根据id查询文章同时加载作者信息
	var article models.Article
	if err := models.DB.Preload("User").First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "文章未找到"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "查询成功", "article": article})

}

func UpdateArticle(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 根据 ID 查询文章
	var article models.Article
	if err := models.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "文章未找到"})
		return
	}

	// 验证当前用户是否是文章的作者或管理员（假设管理员 ID=1）
	if uid, exists := c.Get("userID"); exists {
		var userID = uid.(uint)
		if article.UserID != userID && userID != 1 {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权修改此文章"})
			return
		}
	}

	//权限检查通过后修改文章信息
	article.Title = input.Title
	article.Content = input.Content

	//保存修改后的文章信息到数据库
	if err := models.DB.Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

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
