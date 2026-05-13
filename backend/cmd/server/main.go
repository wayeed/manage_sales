// @title           家具提成管理系统 API
// @version         1.0
// @description     家具提成管理系统后端 API 接口文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   API 支持
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer Token 认证，格式: Bearer {token}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"furniture-commission/configs"
	"furniture-commission/internal/middleware"
	"furniture-commission/internal/pkg/database"
	"furniture-commission/internal/pkg/snowflake"
	"furniture-commission/internal/router"

	"github.com/sirupsen/logrus"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := configs.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化日志
	middleware.InitLogger(&cfg.Log)
	logrus.Info("日志系统初始化完成")

	// 确保日志目录存在
	logDir := filepath.Dir(cfg.Log.File)
	if logDir != "" && logDir != "." {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			logrus.Warnf("创建日志目录失败: %v", err)
		}
	}

	// 设置gin运行模式
	switch cfg.Server.Mode {
	case "debug":
		ginMode := os.Getenv("GIN_MODE")
		if ginMode == "" {
			os.Setenv("GIN_MODE", "debug")
		}
	case "release":
		os.Setenv("GIN_MODE", "release")
	case "test":
		os.Setenv("GIN_MODE", "test")
	}

	// 初始化数据库
	if err := database.InitDB(&cfg.Database); err != nil {
		logrus.Fatalf("数据库初始化失败: %v", err)
	}
	logrus.Info("数据库连接成功")

	// 初始化雪花算法ID生成器
	snowflake.Init(1)
	logrus.Info("ID生成器初始化完成")

	// 注册路由
	r := router.SetupRouter()

	// 启动HTTP服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logrus.Infof("服务器启动中，监听端口: %s", addr)

	if err := r.Run(addr); err != nil {
		logrus.Fatalf("服务器启动失败: %v", err)
	}
}
