package middleware

import (
	"furniture-commission/configs"
	"log"

	"github.com/gin-gonic/gin"
)

// getAllowedOrigins 获取允许的跨域域名列表
func getAllowedOrigins() []string {
	// 默认允许的域名列表（当配置不存在时使用）
	defaultOrigins := []string{
		"http://127.0.0.1:8080",
		"http://localhost:8080",
	}

	// 检查全局配置是否存在
	if configs.GlobalConfig == nil {
		log.Println("[CORS] 警告：全局配置未加载，使用默认域名列表")
		return defaultOrigins
	}

	// 检查CORS配置是否存在
	if len(configs.GlobalConfig.CORS.AllowedOrigins) == 0 {
		log.Println("[CORS] 警告：配置文件中未设置allowed_origins，使用默认域名列表")
		return defaultOrigins
	}

	return configs.GlobalConfig.CORS.AllowedOrigins
}

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	// 初始化时获取允许的域名列表
	allowedOrigins := getAllowedOrigins()

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否允许该域名
		allowOrigin := ""
		for _, o := range allowedOrigins {
			if origin == o {
				allowOrigin = origin
				break
			}
		}

		// 如果没有匹配，处理特殊情况
		if allowOrigin == "" {
			// 开发环境可以临时允许任意来源
			if configs.GlobalConfig != nil && configs.GlobalConfig.Server.Mode == "debug" {
				allowOrigin = origin
			} else {
				// 生产环境：允许无Origin头的请求（如移动端直接调用API）
				if origin == "" {
					allowOrigin = "*"
				} else {
					// 检查是否是公开接口（如登录），允许更多来源
					path := c.Request.URL.Path
					if isPublicPath(path) {
						// 公开接口允许任意来源
						allowOrigin = origin
					} else {
						log.Printf("[CORS] 拒绝未授权的来源: %s", origin)
						c.AbortWithStatus(403)
						return
					}
				}
			}
		}

		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-CSRF-Token, X-Requested-With, Accept")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if allowOrigin != "*" {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isPublicPath 检查路径是否为公开接口
func isPublicPath(path string) bool {
	publicPaths := []string{
		"/api/login",
		"/api/health",
		"/api/csrf-token",
		"/api/app-versions/latest",
	}
	for _, p := range publicPaths {
		if path == p {
			return true
		}
	}
	return false
}

// CSRF CSRF保护中间件
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过公开接口的CSRF检查（如登录、健康检查等）
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 仅对状态修改请求检查
		method := c.Request.Method
		if method != "GET" && method != "HEAD" && method != "OPTIONS" {
			// 优先从请求头获取token（支持移动端自定义header传递）
			token := c.GetHeader("X-CSRF-Token")

			// 如果header中没有，尝试从cookie获取（PC端浏览器）
			if token == "" {
				token, _ = c.Cookie("csrf_token")
			}

			if token == "" {
				log.Println("[CSRF] 缺少CSRF token")
				c.AbortWithStatusJSON(403, gin.H{
					"code":    403,
					"message": "CSRF token required",
				})
				return
			}

			// 验证token：从cookie中获取服务端存储的token进行对比
			sessionToken, _ := c.Cookie("csrf_token")
			// 如果cookie中没有session token，则使用请求中的token作为session token（首次验证时）
			if sessionToken == "" {
				// 移动端场景：token通过header传递，此时将token设置到cookie中用于后续验证
				c.SetCookie("csrf_token", token, 3600*24*7, "/", "", false, false)
				sessionToken = token
			}
			if token != sessionToken {
				log.Println("[CSRF] 无效的CSRF token")
				c.AbortWithStatusJSON(403, gin.H{
					"code":    403,
					"message": "Invalid CSRF token",
				})
				return
			}
		}

		c.Next()
	}
}
