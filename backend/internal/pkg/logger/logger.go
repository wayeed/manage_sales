package logger

import (
	"os"
	"regexp"

	"github.com/sirupsen/logrus"
)

// SensitiveDataFilter 过滤敏感数据
type SensitiveDataFilter struct {
	sensitiveFields map[string]bool
	secretPatterns  []*regexp.Regexp
}

// NewSensitiveDataFilter 创建敏感数据过滤器
func NewSensitiveDataFilter() *SensitiveDataFilter {
	return &SensitiveDataFilter{
		sensitiveFields: map[string]bool{
			"password":      true,
			"token":         true,
			"secret":        true,
			"api_key":       true,
			"access_token":  true,
			"refresh_token": true,
			"credit_card":   true,
			"ssn":           true,
		},
		secretPatterns: []*regexp.Regexp{
			regexp.MustCompile(`password["\']?\s*[:=]\s*["\']?[^"\']*["\']?`),
			regexp.MustCompile(`token["\']?\s*[:=]\s*["\']?[^"\']*["\']?`),
			regexp.MustCompile(`secret["\']?\s*[:=]\s*["\']?[^"\']*["\']?`),
		},
	}
}

// Filter 过滤敏感信息
func (f *SensitiveDataFilter) Filter(entry *logrus.Entry) *logrus.Entry {
	// 过滤 Data 字段
	for key, value := range entry.Data {
		if f.sensitiveFields[key] {
			entry.Data[key] = "***REDACTED***"
		} else if str, ok := value.(string); ok {
			// 应用正则表达式过滤
			for _, pattern := range f.secretPatterns {
				if pattern.MatchString(str) {
					entry.Data[key] = "***REDACTED***"
					break
				}
			}
		}
	}

	// 过滤 Message
	for _, pattern := range f.secretPatterns {
		if pattern.MatchString(entry.Message) {
			entry.Message = "***SENSITIVE DATA REDACTED***"
			break
		}
	}

	return entry
}

// InitLogger 初始化日志记录器
func InitLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	// 设置日志级别
	switch logLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// 使用 JSON 格式
	logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: false,
	})

	// 添加敏感数据过滤钩子
	filter := NewSensitiveDataFilter()
	logger.AddHook(&LogFilterHook{filter: filter})

	return logger
}

// LogFilterHook 日志钩子
type LogFilterHook struct {
	filter *SensitiveDataFilter
}

func (h *LogFilterHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *LogFilterHook) Fire(entry *logrus.Entry) error {
	h.filter.Filter(entry)
	return nil
}

// GetLogger 获取默认日志记录器
func GetLogger() *logrus.Logger {
	return InitLogger("info")
}
