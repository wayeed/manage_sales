package task

import (
	"time"

	"furniture-commission/internal/service"

	"github.com/sirupsen/logrus"
)

// CommissionTask 提成核算定时任务
type CommissionTask struct {
	commissionService *service.CommissionService
	fundPoolService   *service.FundPoolService
}

// NewCommissionTask 创建提成核算定时任务
func NewCommissionTask(commissionService *service.CommissionService, fundPoolService *service.FundPoolService) *CommissionTask {
	return &CommissionTask{
		commissionService: commissionService,
		fundPoolService:   fundPoolService,
	}
}

// MonthlySettlement 月度结算任务
func (t *CommissionTask) MonthlySettlement() {
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 1)
	periodValue := lastMonth.Format("2006-01")

	logrus.Infof("开始处理 %s 月度结算", periodValue)

	// 1. 基金池月度结算
	// 注意：实际使用时需要遍历所有门店
	// 这里仅做示例，实际应从数据库获取所有门店ID
	err := t.fundPoolService.SettleFundPool(1, 1, periodValue)
	if err != nil {
		logrus.Errorf("基金池月度结算失败: %v", err)
	} else {
		logrus.Infof("基金池 %s 月度结算完成", periodValue)
	}

	logrus.Infof("月度结算任务完成: %s", periodValue)
}
