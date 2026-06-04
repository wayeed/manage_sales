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
	StoreID        string `form:"store_id" example:"1"`
	EmployeeID     string `form:"employee_id" example:"1"`
	CommissionType string `form:"commission_type" example:"1"`
	Status         string `form:"status" example:"1"`
	PeriodValue    string `form:"period_value" example:"2024-01"`
	Page           int    `form:"page" example:1`
	PageSize       int    `form:"page_size" example:10`
}

// CommissionSummary 提成汇总
type CommissionSummary struct {
	SalesCommission decimal.Decimal `json:"sales_commission" example:5000.00`
	PeerShare       decimal.Decimal `json:"peer_share" example:1000.00`
	TeamShare       decimal.Decimal `json:"team_share" example:500.00`
	FundPool        decimal.Decimal `json:"fund_pool" example:300.00`
	ReferralReward  decimal.Decimal `json:"referral_reward" example:200.00`
	Total           decimal.Decimal `json:"total" example:7000.00`
}

// EmployeeCommissionSummary 员工月度提成汇总
type EmployeeCommissionSummary struct {
	EmployeeID      int64           `json:"employee_id" example:1`
	EmployeeName    string          `json:"employee_name" example:"张三"`
	SalesCommission decimal.Decimal `json:"sales_commission" example:5000.00`
	TeamShare       decimal.Decimal `json:"team_share" example:500.00`
	FundPoolReward  decimal.Decimal `json:"fund_pool_reward" example:300.00`
	ReferralReward  decimal.Decimal `json:"referral_reward" example:200.00`
	Total           decimal.Decimal `json:"total" example:7000.00`
}

// ManualAdjustRequest 手工调整提成请求
type ManualAdjustRequest struct {
	CommissionID int64   `json:"commission_id" binding:"required" example:1`
	Amount       float64 `json:"amount" binding:"required" example:100.00`
	Remark       string  `json:"remark" example:"手工调整提成"`
}

// ManualSettleRequest 手动月度结算请求
type ManualSettleRequest struct {
	PeriodValue string `json:"period_value" binding:"required" example:"2024-01"`
	SettleType  string `json:"settle_type" example:"all"` // all, fund, fixed
}

// ManualSettleResult 手动结算结果
type ManualSettleResult struct {
	PeriodValue           string `json:"period_value"`
	FundPoolResult        string `json:"fund_pool_result"`
	FixedCommissionResult string `json:"fixed_commission_result"`
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

// getCommissionRate 根据订单类型和用户等级获取提成比例
func (s *CommissionService) getCommissionRate(orderType int8, salesmanLevel int8, rates map[string]decimal.Decimal) decimal.Decimal {
	// 非同行订单，根据用户等级获取比例
	if salesmanLevel <= 0 {
		salesmanLevel = 1 // 默认初级
	}

	isMulti := false
	switch orderType {
	case 1: // 单品
		isMulti = false
	case 2: // 多品
		isMulti = true
	case 3: // 特批
		isMulti = false
	default:
		isMulti = false
	}

	// 使用等级对应的提成比例
	key := fmt.Sprintf("commission_rate_level%d_", salesmanLevel)
	if isMulti {
		key += "multi"
	} else {
		key += "single"
	}

	if rate, ok := rates[key]; ok {
		return rate
	}

	// 降级使用初级等级比例
	if isMulti {
		return rates["commission_rate_level1_multi"]
	}
	return rates["commission_rate_level1_single"]
}

// getSalesmanRateByLevel 根据业务员级别和订单类型获取比例（用于同行带单订单）
// 同行带单订单中，业务员提成 = (级别比例 - 同行比例) × 利润
func (s *CommissionService) getSalesmanRateByLevel(salesmanLevel int8, orderType int8, rates map[string]decimal.Decimal) decimal.Decimal {
	if salesmanLevel <= 0 {
		salesmanLevel = 1 // 默认初级
	}

	// 根据同行订单类型判断单品/多品
	isMulti := false
	switch orderType {
	case 4: // 同行单品
		isMulti = false
	case 5: // 同行多品
		isMulti = true
	case 6: // 同行特批
		isMulti = false
	}

	// 使用等级对应的提成比例
	key := fmt.Sprintf("commission_rate_level%d_", salesmanLevel)
	if isMulti {
		key += "multi"
	} else {
		key += "single"
	}

	if rate, ok := rates[key]; ok {
		return rate
	}

	// 降级使用初级等级比例
	if isMulti {
		return rates["commission_rate_level1_multi"]
	}
	return rates["commission_rate_level1_single"]
}

// getPeerRate 根据订单类型获取同行比例
func (s *CommissionService) getPeerRate(orderType int8, rates map[string]decimal.Decimal) decimal.Decimal {
	switch orderType {
	case 4: // 同行单品
		return rates["commission_rate_peer_single"]
	case 5: // 同行多品
		return rates["commission_rate_peer_multi"]
	case 6: // 同行特批
		return rates["commission_rate_peer_special"]
	}
	// 默认返回单品比例
	return rates["commission_rate_peer_single"]
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

	// 3. 获取业务员信息和等级
	var salesman models.User
	var salesmanLevel int8 = 1 // 默认初级
	if err := s.db.First(&salesman, order.SalesmanID).Error; err == nil {
		salesmanLevel = salesman.Level
		if salesmanLevel <= 0 {
			salesmanLevel = 1
		}
	}

	// 获取当前月份作为周期值
	periodValue := time.Now().Format("2006-01")

	// 使用事务保证原子性
	return s.db.Transaction(func(tx *gorm.DB) error {
		var commissions []models.Commission

		// 4. 根据订单类型计算提成
		var salesmanRate, salesmanAmount decimal.Decimal
		var peerRate, peerAmount decimal.Decimal

		if order.IsPeerOrder == 1 && order.PeerID != nil {
			// 同行带单订单：业务员提成 = (级别比例 - 同行比例) × 利润
			salesmanRate = s.getSalesmanRateByLevel(salesmanLevel, order.OrderType, rates)
			peerRate = s.getPeerRate(order.OrderType, rates)

			// 业务员实际提成比例 = 级别比例 - 同行比例
			actualSalesmanRate := salesmanRate.Sub(peerRate)
			if actualSalesmanRate.LessThan(decimal.Zero) {
				actualSalesmanRate = decimal.Zero
			}
			salesmanAmount = order.ActualProfit.Mul(actualSalesmanRate)
			peerAmount = order.ActualProfit.Mul(peerRate)
		} else {
			// 普通订单：直接使用业务员级别对应比例
			salesmanRate = s.getCommissionRate(order.OrderType, salesmanLevel, rates)
			salesmanAmount = order.ActualProfit.Mul(salesmanRate)
		}

		// 创建业务员提成记录（commission_type=1）
		commissions = append(commissions, models.Commission{
			StoreID:        order.StoreID,
			OrderID:        order.ID,
			EmployeeID:     &order.SalesmanID,
			CommissionType: 1,
			PeriodValue:    periodValue,
			BaseAmount:     order.ActualProfit,
			Rate:           salesmanRate,
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
				Rate:           peerRate,
				Amount:         peerAmount,
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
	Items       []EstimateItem `json:"items" binding:"required" example:[]`
	Gifts       []EstimateGift `json:"gifts" example:[]`
	IsPeerOrder int            `json:"is_peer_order" example:0`
	UserID      int64          `json:"-"` // 从上下文获取，不从前端传
}

// EstimateItem 预估商品项
type EstimateItem struct {
	SKUID     int64   `json:"sku_id" example:1`
	ListPrice float64 `json:"list_price" example:8999.00`
	SalePrice float64 `json:"sale_price" example:8500.00"`
	CostPrice float64 `json:"cost_price" example:6000.00`
	Quantity  int     `json:"quantity" example:1`
}

// EstimateGift 预估礼品项
type EstimateGift struct {
	CostPrice float64 `json:"cost_price" example:50.00`
	Quantity  int     `json:"quantity" example:1`
}

// EstimateCommissionResponse 预估提成响应
type EstimateCommissionResponse struct {
	EstimatedProfit     decimal.Decimal `json:"estimated_profit" example:2500.00`
	CommissionRate      decimal.Decimal `json:"commission_rate" example:0.20`
	EstimatedCommission decimal.Decimal `json:"estimated_commission" example:500.00`
	CategoryCount       int             `json:"category_count" example:1`
	ShowCategoryHint    bool            `json:"show_category_hint" example:false`
	CategoryHint        string          `json:"category_hint" example:""`
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

	// 减去礼品成本
	for _, gift := range req.Gifts {
		giftCost := decimal.NewFromFloat(gift.CostPrice).Mul(decimal.NewFromInt(int64(gift.Quantity)))
		totalCostPrice = totalCostPrice.Add(giftCost)
	}

	// 根据SKU查询品类数量
	categorySet := make(map[int64]bool)
	for _, item := range req.Items {
		if item.SKUID > 0 {
			var sku models.ProductSKU
			if err := s.db.Preload("Product").First(&sku, item.SKUID).Error; err == nil {
				if sku.Product != nil && sku.Product.CategoryID != nil {
					categorySet[*sku.Product.CategoryID] = true
				}
			}
		}
	}
	categoryCount := len(categorySet)
	if categoryCount == 0 {
		categoryCount = 1 // 无品类信息时默认单品
	}

	isMulti := categoryCount >= 3

	// 多品折扣：3个及以上品类，售价总和打95折
	if isMulti {
		totalSalePrice = totalSalePrice.Mul(decimal.NewFromFloat(0.95))
	}

	estimatedProfit := totalSalePrice.Sub(totalCostPrice)
	if estimatedProfit.LessThan(decimal.Zero) {
		estimatedProfit = decimal.Zero
	}

	// 获取当前用户等级
	var userLevel int8 = 1 // 默认初级
	var user models.User
	if err := s.db.First(&user, req.UserID).Error; err == nil {
		if user.Level > 0 {
			userLevel = user.Level
		}
	}

	// 获取提成比例配置
	rates, err := s.configService.GetCommissionRates()
	if err != nil {
		rates = map[string]decimal.Decimal{
			"commission_rate_level1_single": decimal.NewFromFloat(0.08),
			"commission_rate_level1_multi":  decimal.NewFromFloat(0.10),
			"commission_rate_peer_single":   decimal.NewFromFloat(0.10),
		}
	}

	// 根据订单类型和用户等级获取提成比例
	var commissionRate decimal.Decimal
	var actualRate decimal.Decimal // 实际提成比例（用于显示）

	if req.IsPeerOrder == 1 {
		// 同行带单订单：业务员提成 = (级别比例 - 同行比例)
		var orderType int8 = 4 // 同行单品
		if isMulti {
			orderType = 5 // 同行多品
		}
		salesmanRate := s.getSalesmanRateByLevel(userLevel, orderType, rates)
		peerRate := s.getPeerRate(orderType, rates)

		// 实际提成比例 = 级别比例 - 同行比例
		actualRate = salesmanRate.Sub(peerRate)
		if actualRate.LessThan(decimal.Zero) {
			actualRate = decimal.Zero
		}
		// 显示比例用实际比例
		commissionRate = actualRate
	} else {
		// 非同行订单：使用用户等级对应比例
		var orderType int8 = 1 // 单品
		if isMulti {
			orderType = 2 // 多品
		}
		commissionRate = s.getCommissionRate(orderType, userLevel, rates)
		actualRate = commissionRate
	}

	// 计算预估提成
	estimatedCommission := estimatedProfit.Mul(actualRate)

	// 品类提示
	showCategoryHint := categoryCount == 2
	categoryHint := ""
	if showCategoryHint {
		categoryHint = "当前还差一个品类就可以按整单结算享受95折，且可拿更高提成！"
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

// ReverseCommission 回滚订单提成
func (s *CommissionService) ReverseCommission(orderID int64, tx *gorm.DB) error {
	// 查询该订单已结算的提成记录（status=1表示可发放状态）
	var commissions []models.Commission
	if err := tx.Where("order_id = ? AND status = 1", orderID).Find(&commissions).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "查询提成记录失败"}
	}

	for _, comm := range commissions {
		// 创建负数的提成冲减记录
		reverse := models.Commission{
			StoreID:        comm.StoreID,
			OrderID:        comm.OrderID,
			EmployeeID:     comm.EmployeeID,
			PeerID:         comm.PeerID,
			CommissionType: comm.CommissionType,
			PeriodValue:    comm.PeriodValue,
			BaseAmount:     comm.BaseAmount,
			Rate:           comm.Rate,
			Amount:         comm.Amount.Neg(), // 负数
			Status:         3,                 // 3=已冲减
			Remark:         "退货冲减",
		}
		if err := tx.Create(&reverse).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "创建提成冲减记录失败"}
		}

		// 更新原提成记录状态
		if err := tx.Model(&comm).Updates(map[string]interface{}{
			"status": 3, // 已冲减
			"remark": comm.Remark + " [被退货冲减]",
		}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新原提成状态失败"}
		}
	}

	return nil
}

// DeleteOrderCommissions 删除指定订单的所有提成记录
func (s *CommissionService) DeleteOrderCommissions(orderID int64) error {
	return s.db.Where("order_id = ?", orderID).Delete(&models.Commission{}).Error
}

// ManualMonthlySettlement 手动触发月度结算
func (s *CommissionService) ManualMonthlySettlement(req *ManualSettleRequest) (*ManualSettleResult, error) {
	result := &ManualSettleResult{
		PeriodValue: req.PeriodValue,
	}

	settleType := req.SettleType
	if settleType == "" {
		settleType = "all"
	}

	// 1. 基金池结算（遍历所有门店）
	if settleType == "all" || settleType == "fund" {
		var stores []models.Store
		if err := s.db.Where("status = 1").Find(&stores).Error; err != nil {
			result.FundPoolResult = "查询门店失败"
		} else {
			successCount := 0
			failCount := 0
			for _, store := range stores {
				// 使用内部方法执行基金池结算
				err := s.settleFundPoolForStore(store.ID, 1, req.PeriodValue)
				if err != nil {
					failCount++
				} else {
					successCount++
				}
			}
			result.FundPoolResult = fmt.Sprintf("共%d个门店，成功%d，失败%d", len(stores), successCount, failCount)
		}
	}

	// 2. 固定提成结算
	if settleType == "all" || settleType == "fixed" {
		// 查找 FixedCommissionService 并调用
		// 通过直接查询数据库执行固定提成逻辑
		err := s.calculateFixedCommission(req.PeriodValue)
		if err != nil {
			result.FixedCommissionResult = "失败: " + err.Error()
		} else {
			result.FixedCommissionResult = "成功"
		}
	}

	return result, nil
}

// settleFundPoolForStore 为指定门店执行基金池结算
func (s *CommissionService) settleFundPoolForStore(storeID int64, periodType int, periodValue string) error {
	// 获取提成比例配置
	rates, err := s.configService.GetCommissionRates()
	if err != nil {
		return fmt.Errorf("获取提成比例配置失败")
	}
	extractRate := rates["fund_pool_extract_rate"]

	// 检查是否已结算
	var existingCount int64
	s.db.Model(&models.FundPool{}).
		Where("store_id = ? AND period_type = ? AND period_value = ? AND status = 1", storeID, periodType, periodValue).
		Count(&existingCount)

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 汇总该周期内所有commission_type=5(基金池)的提成记录
		var fundCommissions []models.Commission
		err := tx.Where("store_id = ? AND commission_type = 5 AND period_value = ? AND status = 1", storeID, periodValue).
			Find(&fundCommissions).Error
		if err != nil {
			return err
		}

		if len(fundCommissions) == 0 {
			// 该门店该周期无基金池提成，跳过
			return nil
		}

		// 计算pool_amount
		poolAmount := decimal.Zero
		for _, c := range fundCommissions {
			poolAmount = poolAmount.Add(c.Amount)
		}

		// 汇总该周期内所有员工的个人利润
		type EmployeeProfit struct {
			EmployeeID     int64
			PersonalProfit decimal.Decimal
		}
		var employeeProfits []EmployeeProfit
		err = tx.Model(&models.Commission{}).
			Select("employee_id, COALESCE(SUM(base_amount), 0) as personal_profit").
			Where("store_id = ? AND commission_type = 1 AND period_value = ? AND status = 1", storeID, periodValue).
			Group("employee_id").
			Scan(&employeeProfits).Error
		if err != nil {
			return err
		}

		// 计算总利润
		totalProfit := decimal.Zero
		for _, ep := range employeeProfits {
			totalProfit = totalProfit.Add(ep.PersonalProfit)
		}

		// 计算每个员工的份额
		var shares []models.FundPoolShare
		totalShares := decimal.Zero
		perShareAmount := decimal.Zero

		if totalProfit.GreaterThan(decimal.Zero) {
			for _, ep := range employeeProfits {
				shareRatio := ep.PersonalProfit.Div(totalProfit)
				shares = append(shares, models.FundPoolShare{
					EmployeeID:     ep.EmployeeID,
					PersonalProfit: ep.PersonalProfit,
					Shares:         shareRatio,
					RewardAmount:   poolAmount.Mul(shareRatio),
					Status:         0,
				})
				totalShares = totalShares.Add(shareRatio)
			}
			if totalShares.GreaterThan(decimal.Zero) {
				perShareAmount = poolAmount.Div(totalShares)
			}
		}

		// 创建或更新fund_pools记录
		now := time.Now()
		fundPool := &models.FundPool{
			StoreID:        storeID,
			PeriodType:     int8(periodType),
			PeriodValue:    periodValue,
			TotalProfit:    totalProfit,
			ExtractRate:    extractRate,
			PoolAmount:     poolAmount,
			TotalShares:    totalShares,
			PerShareAmount: perShareAmount,
			Status:         1,
			SettledAt:      &now,
		}

		if existingCount > 0 {
			// 更新已有记录
			if err := tx.Model(fundPool).
				Where("store_id = ? AND period_type = ? AND period_value = ? AND status = 1", storeID, periodType, periodValue).
				Updates(map[string]interface{}{
					"total_profit":     totalProfit,
					"extract_rate":     extractRate,
					"pool_amount":      poolAmount,
					"total_shares":     totalShares,
					"per_share_amount": perShareAmount,
					"settled_at":       now,
				}).Error; err != nil {
				return err
			}
			// 获取fund_pool id用于更新shares
			tx.Where("store_id = ? AND period_type = ? AND period_value = ? AND status = 1", storeID, periodType, periodValue).
				First(fundPool)
		} else {
			if err := tx.Create(fundPool).Error; err != nil {
				return err
			}
		}

		// 创建fund_pool_shares记录
		if len(shares) > 0 {
			// 删除旧的份额记录
			tx.Where("fund_pool_id = ?", fundPool.ID).Delete(&models.FundPoolShare{})
			for i := range shares {
				shares[i].FundPoolID = fundPool.ID
			}
			if err := tx.Create(&shares).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// calculateFixedCommission 计算指定月份的固定提成
func (s *CommissionService) calculateFixedCommission(periodValue string) error {
	rates, err := s.configService.GetCommissionRates()
	if err != nil {
		return fmt.Errorf("获取提成比例配置失败")
	}
	fixedRate := rates["fixed_commission_rate"]
	if fixedRate.IsZero() {
		fixedRate, _ = decimal.NewFromString("0.05") // 默认5%
	}

	// 查询该月份内已回款的订单（非同行带单）
	type OrderSummary struct {
		OrderID    int64
		SalesmanID int64
		TotalPaid  decimal.Decimal
	}
	var orders []OrderSummary
	err = s.db.Model(&models.Order{}).
		Select("orders.id as order_id, orders.salesman_id, COALESCE(SUM(payments.amount), 0) as total_paid").
		Joins("LEFT JOIN payments ON payments.order_id = orders.id AND payments.status = 2").
		Where("orders.payment_status = 2 AND orders.order_status = 1 AND (orders.is_draft = 0 OR orders.is_draft IS NULL)").
		Where("orders.peer_id = 0 OR orders.peer_id IS NULL").
		Where("DATE_FORMAT(orders.created_at, '%Y-%m') = ?", periodValue).
		Group("orders.id").
		Scan(&orders).Error
	if err != nil {
		return err
	}

	// 为每个订单创建固定提成记录
	for _, o := range orders {
		if o.SalesmanID <= 0 || o.TotalPaid.LessThanOrEqual(decimal.Zero) {
			continue
		}

		amount := o.TotalPaid.Mul(fixedRate)

		// 检查是否已存在
		var existingCount int64
		s.db.Model(&models.Commission{}).
			Where("order_id = ? AND employee_id = ? AND commission_type = 7 AND period_value = ?", o.OrderID, o.SalesmanID, periodValue).
			Count(&existingCount)

		if existingCount > 0 {
			// 更新
			s.db.Model(&models.Commission{}).
				Where("order_id = ? AND employee_id = ? AND commission_type = 7 AND period_value = ?", o.OrderID, o.SalesmanID, periodValue).
				Updates(map[string]interface{}{
					"base_amount": o.TotalPaid,
					"amount":      amount,
				})
		} else {
			// 创建
			salesmanID := o.SalesmanID
			fixedPct, _ := fixedRate.Mul(decimal.NewFromInt(100)).Float64()
			s.db.Create(&models.Commission{
				OrderID:        o.OrderID,
				EmployeeID:     &salesmanID,
				CommissionType: 7,
				BaseAmount:     o.TotalPaid,
				Amount:         amount,
				PeriodValue:    periodValue,
				Status:         1,
				Remark:         fmt.Sprintf("固定提成(%.0f%%)", fixedPct),
			})
		}
	}

	return nil
}
