package main

import (
	"myapp/controllers" // 替换为你的模块名
	"myapp/middlewares" // 替换为你的模块名
	"myapp/models"      // 替换为你的模块名

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化数据库
	models.InitDB()

	// 2. 注册 Gin 路由，r直接可以POST和GET
	r := gin.Default()

	// 开放路由组 (不需要 Token)
	//public :=gin.Default().Group("/api")
	public := r.Group("/api")
	{
		public.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
		public.POST("/register", controllers.Register) //调用 controllers 包里的 Register 方法
		public.POST("/login", controllers.Login)
	}

	// 保护路由组 (使用 JWT 拦截)
	protected := r.Group("/api")
	//这个路由组使用了 JWTAuthMiddleware 中间件，这意味着所有在这个路由组下定义的路由都会先运行middlewares.JWTAuthMiddleware()
	protected.Use(middlewares.JWTAuthMiddleware())
	{
		protected.GET("/profile", controllers.GetProfile)
		//等效于 protected.GET("/profile", middlewares.JWTAuthMiddleware(), controllers.GetProfile)

		// 文章相关接口
		protected.POST("/articles", controllers.CreateArticle) // 发布
		protected.GET("/articles", controllers.GetArticles)    // 列表 (含分页和搜索)
		protected.GET("/articles/:id", controllers.GetArticleDetail)
		// 删除文章（不仅需要登录，还需要 article:delete 权限）                                               // 详情
		protected.PUT("/articles/:id", middlewares.CheckPermission("article:update"), controllers.UpdateArticle)    // 更新
		protected.DELETE("/articles/:id", middlewares.CheckPermission("article:delete"), controllers.DeleteArticle) // 删除
	}

	// 3. 启动服务
	r.Run(":8080")
}
