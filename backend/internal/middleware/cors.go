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

		// 如果没有匹配，但 Origin 为空（如直接访问），也允许
		if allowOrigin == "" && origin != "" {
			// 生产环境可以设置为特定域名
			allowOrigin = origin
		}

		if allowOrigin == "" {
			allowOrigin = "*"
		}

		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With, Accept")
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
