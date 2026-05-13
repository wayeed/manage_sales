package middleware

import (
	"time"

	"furniture-commission/configs"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 请求完成后记录日志
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		entry := logrus.WithFields(logrus.Fields{
			"status":     status,
			"method":     method,
			"path":       path,
			"query":      query,
			"ip":         clientIP,
			"latency":    latency.String(),
			"user_agent": c.Request.UserAgent(),
			"time":       time.Now().Format("2006-01-02 15:04:05"),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else if status >= 500 {
			entry.Error("Server error")
		} else if status >= 400 {
			entry.Warn("Client error")
		} else {
			entry.Info("Request")
		}
	}
}

// InitLogger 初始化日志配置
func InitLogger(cfg *configs.LogConfig) {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	switch cfg.Level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
