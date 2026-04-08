package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your_super_secret_key_123")

type Claims struct {
	UserID               uint `json:"user_id"`
	jwt.RegisteredClaims      //JWT 官方自带的基础信息，包含过期时间、发行人等
}

// GenerateToken 颁发 Token (首字母大写，暴露给 Controller 使用)
func GenerateToken(userID uint) (string, error) {
	claims := Claims{
		UserID: userID,
		//设置JWT的官方信息
		RegisteredClaims: jwt.RegisteredClaims{
			//ExpiresAt 是 JWT 的过期时间，使用 jwt.NewNumericDate 来设置一个具体的过期时间，这里设置为当前时间加 24 小时。
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "my_app",
		},
	}
	// NewWithClaims 方法创建一个新的 JWT 实例(未签名)，指定使用 HS256 签名算法，并传入我们定义的 Claims 结构体。
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// SignedString 方法使用jwtSecret将 JWT 进行签名，生成一个最终的 Token 字符串。
	return token.SignedString(jwtSecret)
}

// JWTAuthMiddleware 拦截并校验 Token
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		//这是前端发过来的，不用管怎么来的
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请求头中缺少 Authorization"})
			c.Abort()
			//Abort 方法会停止当前请求的处理流程，防止后续的处理函数被调用，并且返回一个 HTTP 401 Unauthorized 的响应，提示客户端缺少必要的认证信息。
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		//按空格把字符串切成 2 段。parts[0] 是 "Bearer"，parts[1] 是实际的 Token 字符串
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(
			tokenString,
			&Claims{},
			func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})
		// ParseWithClaims 方法用于解析和验证 JWT Token。它接受三个参数：要解析的 Token 字符串、一个空的 Claims 结构体实例（用于存储解析后的数据），以及一个回调函数（用于提供签名密钥）。
		// 如果解析成功且 Token 有效，它会返回一个包含解析结果的 token 对象；
		// 如果解析失败或 Token 无效，它会返回一个错误。
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		//将解析后的 Claims 转换为我们定义的 Claims 结构体类型。如果转换失败，说明 Token 的格式不正确或数据不完整，我们同样返回一个 HTTP 401 Unauthorized 的响应。
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 解析失败"})
			c.Abort()
			return
		}

		// 校验成功，将 userID 放入gin的上下文
		c.Set("userID", claims.UserID)
		//Set 方法用于将 userID 存储在 Gin 的上下文中，这样后续的处理函数就可以通过 c.Get("userID") 来获取当前请求的用户 ID。

		c.Next()
		//继续调用下一个函数，如果没有调用 Next()，请求就会在这里被中断，不会继续往下执行。
	}
}
