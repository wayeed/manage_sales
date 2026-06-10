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
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// 初始化日志（先写文件，再记录后续日志）
	middleware.InitLogger(&cfg.Log)
	logrus.Info("日志系统初始化完成")

	// 设置gin运行模式
	setGinMode(cfg.Server.Mode)

	// 初始化数据库（传入运行模式以控制 SQL 日志级别）
	if err := database.InitDB(&cfg.Database, cfg.Server.Mode); err != nil {
		logrus.Fatalf("数据库初始化失败: %v", err)
	}
	logrus.Info("数据库连接成功")

	// 初始化雪花算法ID生成器
	snowflake.Init(1)
	logrus.Info("ID生成器初始化完成")

	// 初始化请求频率限制：每分钟 200 次 / IP
	middleware.InitRateLimiter(200, time.Minute)
	logrus.Info("请求频率限制已启用（每分钟 200 次 / IP）")

	// 注册路由
	r := router.SetupRouter()

	// 创建 HTTP Server，配置超时
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 在 goroutine 中启动服务，使主线程可以监听信号
	go func() {
		logrus.Infof("服务器启动中，监听端口: %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 优雅关闭：等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("收到关闭信号，开始优雅关闭...")

	// 设置关闭超时（最多等待 30 秒）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭 HTTP 服务（等待正在处理的请求完成）
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("服务器强制关闭: %v", err)
	}

	// 关闭数据库连接
	if sqlDB, err := database.GetDB().DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			logrus.Errorf("关闭数据库连接失败: %v", err)
		} else {
			logrus.Info("数据库连接已关闭")
		}
	}

	logrus.Info("服务已安全关闭")
}

// setGinMode 设置 Gin 框架的运行模式
func setGinMode(mode string) {
	switch mode {
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
}
