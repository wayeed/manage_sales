package middleware

import (
	"bytes"
	"io"
	"time"

	"furniture-commission/internal/models"
	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type responseRecorder struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// OperationLog 操作日志中间件
func OperationLog(logService *service.OperationLogService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只记录需要认证的写操作接口（POST/PUT/DELETE）
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// 读取请求体
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 使用自定义ResponseWriter记录响应
		recorder := &responseRecorder{ResponseWriter: c.Writer, body: bytes.NewBufferString("")}
		c.Writer = recorder

		// 处理请求
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		// 在 goroutine 外提取所有需要的值，避免访问已回收的 gin.Context
		userID := GetUserID(c)
		var username string
		var user models.User
		if db != nil && userID > 0 {
			db.Select("username, real_name").First(&user, userID)
			if user.RealName != "" {
				username = user.RealName
			} else {
				username = user.Username
			}
		}

		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		now := time.Now()

		// 同步写入数据库
		log := &models.OperationLog{
			UserID:    userID,
			Username:  username,
			Action:    method + " " + path,
			Detail:    requestBody,
			IPAddress: clientIP,
			UserAgent: userAgent,
			CreatedAt: now,
		}

		if err := logService.Create(log); err != nil {
			logrus.WithError(err).Warn("写入操作日志失败")
		}

		// 同时记录到 logrus
		entry := logrus.WithFields(logrus.Fields{
			"user_id":  userID,
			"method":   method,
			"path":     path,
			"ip":       clientIP,
			"latency":  latency.String(),
			"status":   c.Writer.Status(),
		})

		if c.Writer.Status() >= 400 {
			entry.Warn("Operation failed")
		} else {
			entry.Info("Operation success")
		}
	}
}
