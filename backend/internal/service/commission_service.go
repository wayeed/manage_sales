package service

import (
	"errors"
	"fmt"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ListCommissionRequest 提成列表查询请求
type ListCommissionRequest struct {
	StoreID string `form:"store_id" example:"1"`
	EmployeeID string `form:"employee_id" example:"1"`
	CommissionType string `form:"commission_type" example:"1"`
	Status string `form:"status" example:"1"`
	PeriodValue string `form:"period_value" example:"2024-01"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// CommissionSummary 提成汇总
type CommissionSummary struct {
	SalesCommission decimal.Decimal `json:"sales_commission" example:5000.00`
	PeerShare decimal.Decimal `json:"peer_share" example:1000.00`
	TeamShare decimal.Decimal `json:"team_share" example:500.00`
	FundPool decimal.Decimal `json:"fund_pool" example:300.00`
	ReferralReward decimal.Decimal `json:"referral_reward" example:200.00`
	Total decimal.Decimal `json:"total" example:7000.00`
}

// EmployeeCommissionSummary 员工月度提成汇总
type EmployeeCommissionSummary struct {
	EmployeeID int64 `json:"employee_id" example:1`
	EmployeeName string `json:"employee_name" example:"张三"`
	SalesCommission decimal.Decimal `json:"sales_commission" example:5000.00`
	TeamShare decimal.Decimal `json:"team_share" example:500.00`
	FundPoolReward decimal.Decimal `json:"fund_pool_reward" example:300.00`
	ReferralReward decimal.Decimal `json:"referral_reward" example:200.00`
	Total decimal.Decimal `json:"total" example:7000.00`
}

// ManualAdjustRequest 手工调整提成请求
type ManualAdjustRequest struct {
	CommissionID int64 `json:"commission_id" binding:"required" example:1`
	Amount float64 `json:"amount" binding:"required" example:100.00`
	Remark string `json:"remark" example:"手工调整提成"`
}

// CommissionService 提成核算核心服务
type CommissionService struct {
	db             *gorm.DB
	commissionRepo *repository.CommissionRepository
	orderRepo      *repository.OrderRepository
	referralRepo   *repository.ReferralRelationRepository
	configService  *ConfigService
}

// NewCommissionService 创建提成服务实例
func NewCommissionService(
	db *gorm.DB,
	commissionRepo *repository.CommissionRepository,
	orderRepo *repository.OrderRepository,
	referralRepo *repository.ReferralRelationRepository,
	configService *ConfigService,
) *CommissionService {
	return &CommissionService{
		db:             db,
		commissionRepo: commissionRepo,
		orderRepo:      orderRepo,
		referralRepo:   referralRepo,
		configService:  configService,
	}
}

// getCommissionRate 根据订单类型获取提成比例
func (s *CommissionService) getCommissionRate(orderType int8, rates map[string]decimal.Decimal) decimal.Decimal {
	switch orderType {
	case 1: // 单品
		return rates["commission_rate_single"]
	case 2: // 多品
		return rates["commission_rate_multi"]
	case 3: // 特批
		return rates["commission_rate_special"]
	case 4: // 同行单品
		return rates["commission_rate_peer_single"]
	case 5: // 同行多品
		return rates["commission_rate_peer_multi"]
	case 6: // 同行特批
		return rates["commission_rate_peer_special"]
	default:
		return rates["commission_rate_single"]
	}
}

// CalculateOrderCommission 计算订单提成（核心方法）
// 当订单回款完成时调用，为订单计算所有提成
func (s *CommissionService) CalculateOrderCommission(orderID int64) error {
	// 1. 获取订单详情
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "订单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "查询订单失败"}
	}

	// 验证订单状态（已生效且非退货）
	if order.OrderStatus != 1 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "订单状态不允许计算提成"}
	}

	// 检查是否已计算过提成
	exists, err := s.commissionRepo.ExistsByOrderID(orderID)
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "检查提成记录失败"}
	}
	if exists {
		return &AppError{Code: apperrors.BadRequest, Message: "该订单已计算过提成"}
	}

	// 2. 获取提成比例配置
	rates, err := s.configService.GetCommissionRates()
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "获取提成比例配置失败"}
	}

	// 3. 根据order_type确定提成比例
	rate := s.getCommissionRate(order.OrderType, rates)

	// 获取当前月份作为周期值
	periodValue := time.Now().Format("2006-01")

	// 使用事务保证原子性
	return s.db.Transaction(func(tx *gorm.DB) error {
		var commissions []models.Commission

		// 4. 创建业务员提成记录（commission_type=1）
		salesmanAmount := order.ActualProfit.Mul(rate)
		commissions = append(commissions, models.Commission{
			StoreID:        order.StoreID,
			OrderID:        order.ID,
			EmployeeID:     &order.SalesmanID,
			CommissionType: 1,
			PeriodValue:    periodValue,
			BaseAmount:     order.ActualProfit,
			Rate:           rate,
			Amount:         salesmanAmount,
			Status:         1, // 可发放
			Remark:         fmt.Sprintf("订单%s业务员提成", order.OrderNo),
		})

		// 5. 如果是同行订单，创建同行分成记录（commission_type=2）
		if order.IsPeerOrder == 1 && order.PeerID != nil {
			commissions = append(commissions, models.Commission{
				StoreID:        order.StoreID,
				OrderID:        order.ID,
				PeerID:         order.PeerID,
				CommissionType: 2,
				PeriodValue:    periodValue,
				BaseAmount:     order.ActualProfit,
				Rate:           rate,
				Amount:         order.ActualProfit.Mul(rate),
				Status:         1, // 可发放
				Remark:         fmt.Sprintf("订单%s同行分成", order.OrderNo),
			})
		}

		// 6. 查找业务员的直属主管（parent_id），创建主管团队分润（commission_type=3）
		if order.SalesmanID > 0 {
			var salesman models.User
			if err := tx.First(&salesman, order.SalesmanID).Error; err == nil {
				if salesman.ParentID != nil && *salesman.ParentID > 0 {
					managerShareRate := rates["team_share_rate_manager"]
					commissions = append(commissions, models.Commission{
						StoreID:        order.StoreID,
						OrderID:        order.ID,
						EmployeeID:     salesman.ParentID,
						CommissionType: 3,
						PeriodValue:    periodValue,
						BaseAmount:     order.ActualProfit,
						Rate:           managerShareRate,
						Amount:         order.ActualProfit.Mul(managerShareRate),
						Status:         1, // 可发放
						Remark:         fmt.Sprintf("订单%s主管团队分润", order.OrderNo),
					})
				}

				// 7. 查找门店店长（store.manager_id），创建店长团队分润（commission_type=4）
				var store models.Store
				if err := tx.First(&store, order.StoreID).Error; err == nil {
					if store.ManagerID != nil && *store.ManagerID > 0 {
						// 避免与主管重复（如果主管就是店长）
						if store.ManagerID != nil && (salesman.ParentID == nil || *store.ManagerID != *salesman.ParentID) {
							storeShareRate := rates["team_share_rate_store"]
							commissions = append(commissions, models.Commission{
								StoreID:        order.StoreID,
								OrderID:        order.ID,
								EmployeeID:     store.ManagerID,
								CommissionType: 4,
								PeriodValue:    periodValue,
								BaseAmount:     order.ActualProfit,
								Rate:           storeShareRate,
								Amount:         order.ActualProfit.Mul(storeShareRate),
								Status:         1, // 可发放
								Remark:         fmt.Sprintf("订单%s店长团队分润", order.OrderNo),
							})
						}
					}
				}
			}
		}

		// 8. 创建基金池提取记录（commission_type=5）
		fundPoolRate := rates["fund_pool_extract_rate"]
		commissions = append(commissions, models.Commission{
			StoreID:        order.StoreID,
			OrderID:        order.ID,
			CommissionType: 5,
			PeriodValue:    periodValue,
			BaseAmount:     order.ActualProfit,
			Rate:           fundPoolRate,
			Amount:         order.ActualProfit.Mul(fundPoolRate),
			Status:         1, // 可发放
			Remark:         fmt.Sprintf("订单%s基金池提取", order.OrderNo),
		})

		// 9. 检查老带新关系，创建老带新奖励记录（commission_type=6）
		referral, err := s.referralRepo.FindActiveByReferredID(order.SalesmanID)
		if err == nil && referral != nil && referral.Status == 1 {
			referralRewardRate := rates["referral_reward_rate"]
			referralAmount := salesmanAmount.Mul(referralRewardRate)
			commissions = append(commissions, models.Commission{
				StoreID:        order.StoreID,
				OrderID:        order.ID,
				EmployeeID:     &referral.ReferrerID,
				CommissionType: 6,
				PeriodValue:    periodValue,
				BaseAmount:     salesmanAmount,
				Rate:           referralRewardRate,
				Amount:         referralAmount,
				Status:         1, // 可发放
				Remark:         fmt.Sprintf("订单%s老带新奖励（被引荐人:%d）", order.OrderNo, order.SalesmanID),
			})
		}

		// 批量创建提成记录
		if len(commissions) > 0 {
			if err := tx.Create(&commissions).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "创建提成记录失败"}
			}
		}

		return nil
	})
}

// ManualAdjust 手工调整提成金额
func (s *CommissionService) ManualAdjust(commissionID int64, amount decimal.Decimal, remark string) error {
	commission, err := s.commissionRepo.FindByID(commissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "提成记录不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "查询提成记录失败"}
	}

	if commission.Status == 2 {
		return &AppError{Code: apperrors.BadRequest, Message: "已发放的提成不能调整"}
	}

	adjustRemark := "手工调整"
	if remark != "" {
		adjustRemark = remark
	}

	return s.commissionRepo.UpdateFields(commissionID, map[string]interface{}{
		"amount": amount,
		"remark": adjustRemark,
	})
}

// List 提成列表
func (s *CommissionService) List(req *ListCommissionRequest) (*PageResult, error) {
	commissions, total, err := s.commissionRepo.ListWithFilter(
		req.StoreID, req.EmployeeID, req.CommissionType, req.Status, req.PeriodValue,
		req.Page, req.PageSize,
	)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询提成列表失败"}
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	return &PageResult{
		List:     commissions,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetSummary 提成汇总
func (s *CommissionService) GetSummary(employeeID int64, startDate, endDate string) (*CommissionSummary, error) {
	summary, err := s.commissionRepo.GetSummary(employeeID, startDate, endDate)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询提成汇总失败"}
	}

	result := &CommissionSummary{
		SalesCommission: summary["1"],
		PeerShare:       summary["2"],
		TeamShare:       decimal.Zero.Add(summary["3"]).Add(summary["4"]),
		FundPool:        summary["5"],
		ReferralReward:  summary["6"],
	}

	total := decimal.Zero
	for _, v := range summary {
		total = total.Add(v)
	}
	result.Total = total

	return result, nil
}

// GetByOrderID 订单提成明细
func (s *CommissionService) GetByOrderID(orderID int64) ([]models.Commission, error) {
	commissions, err := s.commissionRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询订单提成明细失败"}
	}
	return commissions, nil
}

// GetMonthlySummary 获取月度提成汇总（所有员工）
func (s *CommissionService) GetMonthlySummary(month string) ([]EmployeeCommissionSummary, error) {
	// 查询该月份所有提成记录（预加载员工信息）
	var commissions []models.Commission
	err := s.db.Where("period_value = ?", month).Preload("Employee").Find(&commissions).Error
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询提成记录失败"}
	}

	// 按员工分组汇总
	summaryMap := make(map[int64]*EmployeeCommissionSummary)
	for _, c := range commissions {
		if c.EmployeeID == nil {
			continue
		}
		empID := *c.EmployeeID
		if _, ok := summaryMap[empID]; !ok {
			empName := ""
			if c.Employee != nil {
				empName = c.Employee.RealName
				if empName == "" {
					empName = c.Employee.Username
				}
			}
			summaryMap[empID] = &EmployeeCommissionSummary{
				EmployeeID:   empID,
				EmployeeName: empName,
			}
		}
		s := summaryMap[empID]
		switch c.CommissionType {
		case 1: // 销售提成
			s.SalesCommission = s.SalesCommission.Add(c.Amount)
		case 2, 3, 4: // 团队分成
			s.TeamShare = s.TeamShare.Add(c.Amount)
		case 5: // 基金池奖励
			s.FundPoolReward = s.FundPoolReward.Add(c.Amount)
		case 6: // 推荐奖励
			s.ReferralReward = s.ReferralReward.Add(c.Amount)
		}
		s.Total = s.Total.Add(c.Amount)
	}

	// 转换为切片
	result := make([]EmployeeCommissionSummary, 0, len(summaryMap))
	for _, s := range summaryMap {
		result = append(result, *s)
	}

	return result, nil
}

// EstimateCommissionRequest 预估提成请求
type EstimateCommissionRequest struct {
	Items []EstimateItem `json:"items" binding:"required" example:[]`
	IsPeerOrder int `json:"is_peer_order" example:0`
}

// EstimateItem 预估商品项
type EstimateItem struct {
	ListPrice float64 `json:"list_price" example:8999.00`
	SalePrice float64 `json:"sale_price" example:8500.00`
	CostPrice float64 `json:"cost_price" example:6000.00`
	Quantity int `json:"quantity" example:1`
}

// EstimateCommissionResponse 预估提成响应
type EstimateCommissionResponse struct {
	EstimatedProfit decimal.Decimal `json:"estimated_profit" example:2500.00`
	CommissionRate decimal.Decimal `json:"commission_rate" example:0.20`
	EstimatedCommission decimal.Decimal `json:"estimated_commission" example:500.00`
	CategoryCount int `json:"category_count" example:1`
	ShowCategoryHint bool `json:"show_category_hint" example:false`
	CategoryHint string `json:"category_hint" example:""`
}

// EstimateCommission 预估提成
func (s *CommissionService) EstimateCommission(req *EstimateCommissionRequest) (*EstimateCommissionResponse, error) {
	// 计算预估利润：售价 - 成本价
	var totalSalePrice, totalCostPrice decimal.Decimal
	for _, item := range req.Items {
		totalSalePrice = totalSalePrice.Add(decimal.NewFromFloat(item.SalePrice).Mul(decimal.NewFromInt(int64(item.Quantity))))
		cost := decimal.NewFromFloat(item.CostPrice)
		if cost.LessThanOrEqual(decimal.Zero) {
			// 如果没传成本价（0或负数），降级用标价作为成本
			cost = decimal.NewFromFloat(item.ListPrice)
		}
		totalCostPrice = totalCostPrice.Add(cost.Mul(decimal.NewFromInt(int64(item.Quantity))))
	}
	estimatedProfit := totalSalePrice.Sub(totalCostPrice)
	if estimatedProfit.LessThan(decimal.Zero) {
		estimatedProfit = decimal.Zero
	}

	// 获取提成比例配置
	rates, err := s.configService.GetCommissionRates()
	if err != nil {
		rates = map[string]decimal.Decimal{
			"commission_rate_single": decimal.NewFromFloat(0.05),
		}
	}

	// 根据订单类型获取提成比例
	commissionRate := decimal.NewFromFloat(0.05) // 默认5%
	if req.IsPeerOrder == 1 {
		if rate, ok := rates["commission_rate_peer_single"]; ok {
			commissionRate = rate
		}
	} else {
		if rate, ok := rates["commission_rate_single"]; ok {
			commissionRate = rate
		}
	}

	// 计算预估提成
	estimatedCommission := estimatedProfit.Mul(commissionRate)

	// 品类提示（假设2个品类时提示）
	categoryCount := len(req.Items)
	showCategoryHint := categoryCount == 2
	categoryHint := ""
	if showCategoryHint {
		categoryHint = "当前还差一个品类就可以拿按整单结算拿更高提成！"
	}

	return &EstimateCommissionResponse{
		EstimatedProfit:     estimatedProfit,
		CommissionRate:      commissionRate,
		EstimatedCommission: estimatedCommission,
		CategoryCount:       categoryCount,
		ShowCategoryHint:    showCategoryHint,
		CategoryHint:        categoryHint,
	}, nil
}
