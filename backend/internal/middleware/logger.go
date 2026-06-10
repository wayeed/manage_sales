package middleware

import (
	"io"
	"os"
	"path/filepath"
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
// 设置日志级别、格式，并同时输出到控制台和文件
func InitLogger(cfg *configs.LogConfig) {
	// 设置日志格式（JSON 格式便于日志采集）
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 设置日志级别
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

	// 打开日志文件并设置同时输出到文件和控制台
	if cfg.File != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.File)
		if logDir != "" && logDir != "." {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				logrus.Warnf("创建日志目录失败: %v", err)
			}
		}

		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logrus.Warnf("无法打开日志文件 %s: %v，仅输出到控制台", cfg.File, err)
		} else {
			// 同时输出到控制台和文件
			multiWriter := io.MultiWriter(os.Stdout, file)
			logrus.SetOutput(multiWriter)
			logrus.Infof("日志文件已配置: %s", cfg.File)
		}
	}
}
