package middleware

import (
	"sync"
	"time"

	"furniture-commission/internal/handler"

	"github.com/gin-gonic/gin"
)

// RateLimiter 基于 IP 的滑动窗口限流器
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // 窗口内允许的最大请求数
	window   time.Duration // 时间窗口大小
	cleanup  time.Duration // 清理过期记录间隔
}

type visitor struct {
	lastSeen time.Time
	timestamps []time.Time
}

// NewRateLimiter 创建限流器
// rate: 每个时间窗口允许的请求数
// window: 时间窗口大小
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
		cleanup:  window * 2, // 清理间隔为窗口的 2 倍
	}

	// 后台定期清理过期记录
	go rl.cleanupRoutine()

	return rl
}

// Allow 检查指定 key 是否允许本次请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]

	if !exists {
		rl.visitors[key] = &visitor{
			lastSeen:   now,
			timestamps: []time.Time{now},
		}
		return true
	}

	v.lastSeen = now

	// 清理窗口外的旧记录
	cutoff := now.Add(-rl.window)
	valid := v.timestamps[:0]
	for _, t := range v.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	v.timestamps = valid

	// 判断是否超过限制
	if len(v.timestamps) >= rl.rate {
		return false
	}

	v.timestamps = append(v.timestamps, now)
	return true
}

// cleanupRoutine 定期清理长时间未访问的记录
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.cleanup {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// 默认全局限流器实例
var globalLimiter *RateLimiter

// InitRateLimiter 初始化全局限流器
func InitRateLimiter(rate int, window time.Duration) {
	if rate <= 0 {
		rate = 100 // 默认每分钟 100 次
	}
	if window <= 0 {
		window = time.Minute
	}
	globalLimiter = NewRateLimiter(rate, window)
}

// RateLimit 请求频率限制中间件
// 基于客户端 IP 进行限流，返回 HTTP 429 Too Many Requests
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalLimiter == nil {
			c.Next()
			return
		}

		key := c.ClientIP()
		if !globalLimiter.Allow(key) {
			handler.ErrorWithHTTPStatus(c, 429, 429, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}
