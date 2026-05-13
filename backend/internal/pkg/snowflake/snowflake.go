package snowflake

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// IDGenerator 雪花算法ID生成器
type IDGenerator struct {
	mu        sync.Mutex
	lastTime  int64
	sequence  int64
	nodeID    int64
}

var generator *IDGenerator
var once sync.Once

// Init 初始化ID生成器
func Init(nodeID int64) {
	once.Do(func() {
		generator = &IDGenerator{
			nodeID:   nodeID,
			lastTime: time.Now().UnixMilli(),
		}
	})
}

// getGenerator 获取生成器实例（默认nodeID=1）
func getGenerator() *IDGenerator {
	if generator == nil {
		Init(1)
	}
	return generator
}

// Generate 生成唯一ID
// 格式: 时间戳(41bit) + 节点ID(10bit) + 序列号(12bit)
func Generate() int64 {
	g := getGenerator()
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()

	if now == g.lastTime {
		g.sequence = (g.sequence + 1) & 0xFFF // 12位序列号，最大4095
		if g.sequence == 0 {
			// 序列号溢出，等待下一毫秒
			for now <= g.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTime = now

	// 组合ID: 时间戳 | 节点ID | 序列号
	id := (now << 22) | (g.nodeID << 12) | g.sequence
	return id
}

// GenerateOrderNo 生成订单号
// 格式: FC + 年月日 + 时分秒 + 6位随机数
func GenerateOrderNo() string {
	now := time.Now()
	dateStr := now.Format("20060102")
	timeStr := now.Format("150405")
	randNum := rand.Intn(900000) + 100000 // 100000-999999
	return fmt.Sprintf("FC%s%s%d", dateStr, timeStr, randNum)
}
