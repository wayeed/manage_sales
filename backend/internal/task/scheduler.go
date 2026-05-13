package task

import (
	"furniture-commission/internal/pkg/database"
	"furniture-commission/internal/repository"
	"furniture-commission/internal/service"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron *cron.Cron
}

// NewScheduler 创建定时任务调度器
func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
	}
}

// Start 启动定时任务
func (s *Scheduler) Start() {
	db := database.GetDB()

	// 初始化依赖
	configRepo := repository.NewSystemConfigRepository(db)
	commissionRepo := repository.NewCommissionRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	referralRepo := repository.NewReferralRelationRepository(db)
	fundPoolRepo := repository.NewFundPoolRepository(db)

	configService := service.NewConfigService(configRepo)
	commissionService := service.NewCommissionService(db, commissionRepo, orderRepo, referralRepo, configService)
	fundPoolService := service.NewFundPoolService(db, fundPoolRepo, commissionRepo, configService)

	// 注册定时任务
	commissionTask := NewCommissionTask(commissionService, fundPoolService)

	// 每月1日凌晨2点执行提成核算
	_, err := s.cron.AddFunc("0 0 2 1 * *", func() {
		logrus.Info("定时任务：开始执行月度提成核算")
		commissionTask.MonthlySettlement()
	})
	if err != nil {
		logrus.Errorf("注册月度提成核算任务失败: %v", err)
	}

	// 启动调度器
	s.cron.Start()
	logrus.Info("定时任务调度器已启动")
}

// Stop 停止定时任务
func (s *Scheduler) Stop() {
	s.cron.Stop()
	logrus.Info("定时任务调度器已停止")
}
