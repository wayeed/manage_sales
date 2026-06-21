package router

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"time"

	"furniture-commission/configs"
	"furniture-commission/internal/handler"
	"furniture-commission/internal/middleware"
	"furniture-commission/internal/pkg/database"
	"furniture-commission/internal/repository"
	"furniture-commission/internal/service"
	"log"
	"os"

	_ "furniture-commission/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// generateCSRFToken 生成CSRF token
func generateCSRFToken() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String()+randString(16))))
}

// randString 生成随机字符串
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// SetupRouter 注册路由
func SetupRouter() *gin.Engine {
	r := gin.New()

	// 获取上传配置
	uploadDir := getUploadDir()
	baseURL := getUploadBaseURL()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.RateLimit())
	r.Use(middleware.Cors())
	r.Use(middleware.Logger())

	// 静态文件服务（上传的文件）
	r.Static("/uploads", uploadDir)

	// 初始化依赖
	db := database.GetDB()

	// 设置认证中间件的数据库实例
	middleware.SetDB(db)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)

	// 新增：商品管理与库存管理 Repository
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	skuRepo := repository.NewSKURepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)

	// 订单管理 Repository
	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	peerRepo := repository.NewPeerRepository(db)
	orderReturnRepo := repository.NewOrderReturnRepository(db)

	// 送货管理 Repository
	deliveryRepo := repository.NewDeliveryRepository(db)

	// 提成与工资管理 Repository
	commissionRepo := repository.NewCommissionRepository(db)
	fundPoolRepo := repository.NewFundPoolRepository(db)
	referralRepo := repository.NewReferralRelationRepository(db)
	salaryRepo := repository.NewSalaryRecordRepository(db)
	configRepo := repository.NewSystemConfigRepository(db)

	// 初始化服务
	authService := service.NewAuthService(userRepo, permRepo)
	userService := service.NewUserService(db, userRepo, roleRepo)
	roleService := service.NewRoleService(db, roleRepo, permRepo)

	// 新增：商品管理与库存管理 Service
	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(db, productRepo, configRepo, inventoryRepo, warehouseRepo)
	skuService := service.NewSKUService(db, skuRepo)
	warehouseService := service.NewWarehouseService(warehouseRepo)
	inventoryService := service.NewInventoryService(db, inventoryRepo, skuRepo, warehouseRepo)
	purchaseService := service.NewPurchaseService(db, inventoryService)
	receiptService := service.NewReceiptService(db, inventoryService)
	giftService := service.NewGiftService(db, inventoryService)
	supplierService := service.NewSupplierService(db)
	transferService := service.NewTransferService(db, inventoryService)
	stockAlertService := service.NewStockAlertService(db, inventoryRepo)

	// 库存盘点 Service
	stocktakeRepo := repository.NewStocktakeRepository(db)
	stocktakeService := service.NewStocktakeService(db, stocktakeRepo, inventoryRepo)

	// 提成与工资管理 Service
	configService := service.NewConfigService(configRepo)
	commissionService := service.NewCommissionService(db, commissionRepo, orderRepo, referralRepo, configService)

	// 订单管理 Service（依赖 commissionService, userService）
	orderService := service.NewOrderService(db, orderRepo, paymentRepo, customerRepo, peerRepo, inventoryService, orderReturnRepo, commissionService, userService)
	customerService := service.NewCustomerService(db, customerRepo)
	peerService := service.NewPeerService(db, peerRepo, userRepo)

	// 送货管理 Service
	deliveryService := service.NewDeliveryService(deliveryRepo, orderRepo, userRepo, inventoryService, db)

	// 出库申请 Service
	outboundRequestService := service.NewOutboundRequestService(db, orderRepo)

	// 回款 Service（依赖 commissionService）
	paymentService := service.NewPaymentService(db, paymentRepo, orderRepo, commissionService)
	fundPoolService := service.NewFundPoolService(db, fundPoolRepo, commissionRepo, configService)
	fixedCommissionService := service.NewFixedCommissionService(db, commissionRepo, orderRepo, configService)
	salaryService := service.NewSalaryService(db, salaryRepo, commissionRepo, fundPoolRepo)

	// 数据分析与报表 Service
	dashboardService := service.NewDashboardService(db, userService)
	reportService := service.NewReportService(db)

	// APP版本管理与平台维护 Service
	appVersionService := service.NewAppVersionService(db)
	maintenanceService := service.NewMaintenanceService(db)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	permHandler := handler.NewPermissionHandler(permRepo, roleService)

	// 新增：商品管理与库存管理 Handler
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService, skuService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	purchaseHandler := handler.NewPurchaseHandler(purchaseService)
	receiptHandler := handler.NewReceiptHandler(receiptService)
	giftHandler := handler.NewGiftHandler(giftService)
	supplierHandler := handler.NewSupplierHandler(supplierService)
	transferHandler := handler.NewTransferHandler(transferService)
	warehouseHandler := handler.NewWarehouseHandler(warehouseService)
	stockAlertHandler := handler.NewStockAlertHandler(stockAlertService)
	stocktakeHandler := handler.NewStocktakeHandler(stocktakeService)

	// 订单管理 Handler
	orderHandler := handler.NewOrderHandler(orderService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	customerHandler := handler.NewCustomerHandler(customerService)
	peerHandler := handler.NewPeerHandler(peerService)

	// 送货管理 Handler
	deliveryHandler := handler.NewDeliveryHandler(deliveryService)

	// 出库申请 Handler
	outboundRequestHandler := handler.NewOutboundRequestHandler(outboundRequestService)

	// 库存穿透查询 Service & Handler
	inventoryTraceService := service.NewInventoryTraceService(db)
	inventoryTraceHandler := handler.NewInventoryTraceHandler(inventoryTraceService)

	// 提成与工资管理 Handler
	commissionHandler := handler.NewCommissionHandler(commissionService, fixedCommissionService, fundPoolService)
	salaryHandler := handler.NewSalaryHandler(salaryService)
	fundPoolHandler := handler.NewFundPoolHandler(fundPoolService)
	configHandler := handler.NewConfigHandler(configService)
	performanceHandler := handler.NewPerformanceHandler(db, commissionService)
	referralService := service.NewReferralService(db, referralRepo, userRepo)
	referralHandler := handler.NewReferralHandler(referralService)

	// 门店 Repository（提前初始化供审批服务使用）
	storeRepo := repository.NewStoreRepository(db)

	// 申请跟进审批 Handler
	followUpApprovalRepo := repository.NewFollowUpApprovalRepository(db)
	followUpApprovalService := service.NewFollowUpApprovalService(db, followUpApprovalRepo, customerRepo, userRepo, storeRepo)
	followUpApprovalHandler := handler.NewFollowUpApprovalHandler(followUpApprovalService)

	// 操作日志 Handler
	operationLogRepo := repository.NewOperationLogRepository(db)
	operationLogService := service.NewOperationLogService(operationLogRepo)
	operationLogHandler := handler.NewOperationLogHandler(operationLogService)

	// 门店管理 Handler
	storeService := service.NewStoreService(storeRepo)
	storeHandler := handler.NewStoreHandler(storeService)

	// 文件上传 Handler
	os.MkdirAll(uploadDir, 0755)
	uploadHandler := handler.NewUploadHandler(uploadDir, baseURL)

	// 数据分析与报表 Handler
	dashboardHandler := handler.NewDashboardHandler(dashboardService, db)
	reportHandler := handler.NewReportHandler(reportService, db)

	// APP版本管理与平台维护 Handler
	appVersionHandler := handler.NewAppVersionHandler(appVersionService)
	maintenanceHandler := handler.NewMaintenanceHandler(maintenanceService)

	// 健康检查（公开）
	r.GET("/api/health", func(c *gin.Context) {
		handler.Success(c, gin.H{
			"status":  "ok",
			"service": "furniture-commission",
		})
	})

	// 获取CSRF token（公开）
	r.GET("/api/csrf-token", func(c *gin.Context) {
		// 生成并设置CSRF token到cookie
		csrfToken := generateCSRFToken()
		// HttpOnly设为false，允许前端JavaScript读取cookie
		c.SetCookie("csrf_token", csrfToken, 3600*24*7, "/", "", false, false)
		handler.Success(c, gin.H{
			"message": "CSRF token set",
		})
	})

	// 公开路由
	public := r.Group("/api")
	{
		public.POST("/login", authHandler.Login)
	}

	// 需要认证的路由
	auth := r.Group("/api")
	auth.Use(middleware.Auth())
	auth.Use(middleware.CSRF())
	auth.Use(middleware.OperationLog(operationLogService, db))
	{
		// 登出
		auth.POST("/logout", authHandler.Logout)

		// 当前用户信息
		auth.GET("/users/me", authHandler.GetCurrentUser)

		// 用户管理
		users := auth.Group("/users")
		{
			users.GET("", userHandler.List)                              // 用户列表
			users.GET("/:id", userHandler.GetDetail)                     // 用户详情
			users.PUT("/:id", userHandler.Update)                        // 更新用户
			users.POST("/change-password", userHandler.ChangePassword)   // 修改密码
			users.POST("/:id/reset-password", userHandler.ResetPassword) // 重置密码
			users.PUT("/:id/status", userHandler.UpdateStatus)           // 启用/禁用
			users.POST("/:id/roles", userHandler.AssignRole)             // 分配角色
		}

		// 员工列表（供选择器使用，别名到用户列表）
		auth.GET("/employees", userHandler.List)

		// 角色管理
		roles := auth.Group("/roles")
		{
			roles.GET("", roleHandler.List)                      // 角色列表
			roles.GET("/:id", roleHandler.GetDetail)             // 角色详情
			roles.GET("/:id/permissions", permHandler.GetByRole) // 角色权限
		}

		// 权限管理
		permissions := auth.Group("/permissions")
		{
			permissions.GET("", permHandler.List)         // 权限列表
			permissions.GET("/tree", permHandler.GetTree) // 权限树
		}

		// ===== 品类管理 =====
		categories := auth.Group("/categories")
		{
			categories.GET("", categoryHandler.List)     // 品类列表
			categories.GET("/:id", categoryHandler.List) // 品类详情（复用List）
		}

		// ===== 商品管理 =====
		products := auth.Group("/products")
		{
			products.GET("", productHandler.List)             // 商品列表
			products.GET("/:id", productHandler.GetDetail)    // 商品详情
			products.GET("/:id/skus", productHandler.ListSKU) // 商品SKU列表
		}

		// ===== SKU管理 =====
		skus := auth.Group("/skus")
		skus.GET("", productHandler.ListAllSKU)                  // SKU列表（支持搜索）
		skus.GET("/with-stock", productHandler.ListSKUWithStock) // 带库存的SKU列表（订单选商品）
		{
			skus.PUT("/:id", productHandler.UpdateSKU)    // 更新SKU
			skus.DELETE("/:id", productHandler.DeleteSKU) // 删除SKU
		}

		// ===== 仓库管理 =====
		warehouses := auth.Group("/warehouses")
		{
			warehouses.GET("", warehouseHandler.List)          // 仓库列表
			warehouses.POST("", warehouseHandler.Create)       // 创建仓库
			warehouses.PUT("/:id", warehouseHandler.Update)    // 更新仓库
			warehouses.DELETE("/:id", warehouseHandler.Delete) // 删除仓库
			warehouses.GET("/:id", warehouseHandler.GetByID)   // 仓库详情
		}

		// ===== 库存管理 =====
		inventory := auth.Group("/inventory")
		{
			inventory.GET("/stock", inventoryHandler.GetStock)               // 查询库存
			inventory.GET("/stocks", inventoryHandler.GetStockList)          // 库存列表
			inventory.GET("/transactions", inventoryHandler.GetTransactions) // 库存流水
			inventory.GET("/:id", inventoryHandler.GetStockDetail)           // 库存详情
		}

		// ===== 采购管理 =====
		purchases := auth.Group("/purchases")
		{
			purchases.GET("", purchaseHandler.List)                          // 采购订单列表
			purchases.GET("/mergeable", purchaseHandler.ListMergeableOrders) // 可合并采购单列表
			purchases.GET("/:id", purchaseHandler.GetDetail)                 // 采购订单详情
		}

		// ===== 礼品管理 =====
		gifts := auth.Group("/gifts")
		{
			gifts.GET("", giftHandler.List)          // 礼品列表
			gifts.GET("/:id", giftHandler.GetDetail) // 礼品详情
		}

		// ===== 供应商管理 =====
		suppliers := auth.Group("/suppliers")
		{
			suppliers.GET("", supplierHandler.List)                     // 供应商列表
			suppliers.GET("/:id", supplierHandler.GetDetail)            // 供应商详情
			suppliers.GET("/:id/products", supplierHandler.GetProducts) // 供应商商品列表
		}

		// ===== 调拨管理 =====
		transfers := auth.Group("/transfers")
		{
			transfers.GET("", transferHandler.List)          // 调拨单列表
			transfers.GET("/:id", transferHandler.GetDetail) // 调拨单详情
		}

		// ===== 库存预警 =====
		stockAlerts := auth.Group("/stock-alerts")
		{
			stockAlerts.GET("", stockAlertHandler.List) // 预警列表
		}

		// ===== 库存盘点 =====
		stocktakes := auth.Group("/stocktakes")
		{
			stocktakes.GET("", stocktakeHandler.List)                 // 盘点单列表
			stocktakes.POST("", stocktakeHandler.Create)              // 创建盘点单
			stocktakes.GET("/:id", stocktakeHandler.GetDetail)        // 盘点单详情
			stocktakes.PUT("/:id", stocktakeHandler.Update)           // 更新盘点单
			stocktakes.POST("/:id/submit", stocktakeHandler.Submit)   // 提交盘点单
			stocktakes.POST("/:id/approve", stocktakeHandler.Approve) // 审核盘点单
			stocktakes.DELETE("/:id", stocktakeHandler.Delete)        // 删除盘点单
		}

		// ===== 订单管理 =====
		orders := auth.Group("/orders")
		{
			orders.GET("", orderHandler.List)                               // 订单列表
			orders.GET("/feed", orderHandler.GetOrderFeed)                  // 订单动态
			orders.GET("/customer-draft", orderHandler.GetCustomerDraft)    // 获取客户最新草稿订单
			orders.GET("/:id", orderHandler.GetDetail)                      // 订单详情
			orders.GET("/:id/commission", orderHandler.GetCommissionDetail) // 订单利润提成详情
			orders.POST("", orderHandler.CreateOrder)                       // 创建订单
			orders.PUT("/:id", orderHandler.UpdateOrder)                    // 修改订单
			orders.DELETE("/:id", orderHandler.Delete)                      // 删除订单
		}

		// ===== 回款管理 =====
		payments := auth.Group("/payments")
		{
			payments.GET("", paymentHandler.List)                        // 回款列表
			payments.POST("", paymentHandler.CreatePayment)              // 录入回款
			payments.POST("/:id/approve", paymentHandler.ApprovePayment) // 审核回款（财务可访问）
		}

		// ===== 客户管理 =====
		customers := auth.Group("/customers")
		{
			customers.GET("", customerHandler.List)                                          // 客户列表
			customers.GET("/with-draft-status", customerHandler.GetCustomersWithDraftStatus) // 客户列表（含草稿状态）
			customers.GET("/:id", customerHandler.GetDetail)                                 // 客户详情
			customers.GET("/:id/follow-ups", customerHandler.GetFollowUps)                   // 跟进记录
			customers.POST("/:id/follow-ups", customerHandler.AddFollowUp)                   // 添加跟进记录
			customers.POST("", customerHandler.Create)                                       // 创建客户
			customers.PUT("/:id", customerHandler.Update)                                    // 更新客户
			customers.DELETE("/:id", customerHandler.Delete)                                 // 删除客户
		}

		// ===== 同行管理 =====
		peers := auth.Group("/peers")
		{
			peers.GET("", peerHandler.List)          // 同行列表
			peers.POST("", peerHandler.Create)       // 创建同行
			peers.GET("/:id", peerHandler.GetDetail) // 同行详情
		}

		// ===== 送货管理 =====
		handler.RegisterDeliveryRoutes(auth, deliveryHandler, middleware.RequirePermission)

		// ===== 出库申请管理 =====
		outboundGroup := auth.Group("/outbound-requests")
		{
			outboundGroup.POST("", outboundRequestHandler.CreateRequest)
			outboundGroup.GET("/order/:orderID", outboundRequestHandler.GetByOrderID)
			outboundGroup.GET("/pending", outboundRequestHandler.ListPending)
			outboundGroup.POST("/:id/supervisor-approve", outboundRequestHandler.SupervisorApprove)
			outboundGroup.POST("/:id/finance-approve", outboundRequestHandler.FinanceApprove)
			outboundGroup.POST("/:id/reject", outboundRequestHandler.Reject)
		}

		// ===== 库存穿透查询 =====
		traceGroup := auth.Group("/inventory/trace")
		{
			traceGroup.GET("/forward", inventoryTraceHandler.ForwardTrace)
			traceGroup.GET("/backward", inventoryTraceHandler.BackwardTrace)
			traceGroup.GET("/sku", inventoryTraceHandler.SKUBatchTrace)
		}

		// ===== 采购管理（创建和审核）- 财务角色可访问 =====
		// 创建采购单和审核移到auth组，让财务角色可以访问
		auth.POST("/purchases", middleware.RequireRoleOrPermission([]string{"FINANCE"}, []string{"purchase:manage"}), purchaseHandler.CreateOrder)
		auth.PUT("/purchases/:id/approve", middleware.RequireRoleOrPermission([]string{"FINANCE"}, []string{"purchase:manage"}), purchaseHandler.ApproveOrder)

		// ===== 订单审核 - 主管/店长/财务角色可访问 =====
		auth.POST("/orders/:id/approve", middleware.RequireRoleOrPermission([]string{"FINANCE", "SUPERVISOR", "STORE_MANAGER", "BOSS"}, []string{"order:approve"}), orderHandler.ApproveOrder)
		auth.POST("/orders/:id/cancel", middleware.RequireRoleOrPermission([]string{"FINANCE", "SUPERVISOR", "STORE_MANAGER", "BOSS"}, []string{"order:approve"}), orderHandler.CancelOrder)
		auth.POST("/orders/:id/return", middleware.RequireRoleOrPermission([]string{"FINANCE", "SUPERVISOR", "STORE_MANAGER", "BOSS"}, []string{"order:approve"}), orderHandler.ReturnOrder)
		auth.POST("/orders/:id/generate-purchase", middleware.RequireRoleOrPermission([]string{"FINANCE", "SUPERVISOR", "STORE_MANAGER", "BOSS"}, []string{"order:approve"}), orderHandler.GeneratePurchaseFromOrder)
		auth.POST("/orders/:id/print-approval", middleware.RequireRoleOrPermission([]string{"FINANCE", "SUPERVISOR", "STORE_MANAGER", "BOSS"}, []string{"order:approve"}), followUpApprovalHandler.CreatePrintApproval)
		auth.GET("/orders/:id/print-approval", middleware.RequireRoleOrPermission([]string{"FINANCE", "SUPERVISOR", "STORE_MANAGER", "BOSS"}, []string{"order:approve"}), followUpApprovalHandler.GetPrintApprovalStatus)

		// ===== 提成管理（读操作） =====
		commissions := auth.Group("/commissions")
		{
			commissions.GET("", commissionHandler.List)                         // 提成列表
			commissions.GET("/summary", commissionHandler.GetSummary)           // 提成汇总
			commissions.GET("/order/:order_id", commissionHandler.GetByOrderID) // 订单提成明细
			commissions.POST("/estimate", commissionHandler.EstimateCommission) // 预估提成
		}

		// ===== 业绩查询 =====
		performance := auth.Group("/performance")
		{
			performance.GET("/overview", performanceHandler.GetOverview) // 业绩概览
		}

		// ===== 引荐关系管理 =====
		referrals := auth.Group("/referrals")
		{
			referrals.GET("", referralHandler.List)                     // 引荐关系列表
			referrals.POST("", referralHandler.Create)                  // 创建引荐关系
			referrals.POST("/:id/terminate", referralHandler.Terminate) // 终止引荐关系
		}

		// ===== 申请跟进审批 =====
		followUpApprovals := auth.Group("/follow-up-approvals")
		{
			followUpApprovals.POST("", followUpApprovalHandler.Create)                          // 创建申请
			followUpApprovals.GET("/my", followUpApprovalHandler.ListMyApplications)            // 我的申请列表
			followUpApprovals.GET("/pending", followUpApprovalHandler.ListPendingApprovals)     // 待我审批列表
			followUpApprovals.GET("/processed", followUpApprovalHandler.ListProcessedApprovals) // 我已处理的审批列表
			followUpApprovals.GET("/:id", followUpApprovalHandler.GetStatus)                    // 查询状态
			followUpApprovals.POST("/:id/approve", followUpApprovalHandler.Approve)             // 审批通过
			followUpApprovals.POST("/:id/reject", followUpApprovalHandler.Reject)               // 审批拒绝
			followUpApprovals.POST("/:id/cancel", followUpApprovalHandler.Cancel)               // 撤回申请
		}

		// ===== 工资管理（读操作） =====
		salaries := auth.Group("/salaries")
		{
			salaries.GET("", salaryHandler.List)                                                  // 工资列表
			salaries.GET("/:id", salaryHandler.GetDetail)                                         // 工资详情
			salaries.GET("/:id/export", salaryHandler.ExportSalarySlip)                           // 导出工资条
			salaries.GET("/employee/:employee_id/:salary_month", salaryHandler.GetEmployeeSalary) // 员工月度工资
		}

		// ===== 基金池（读操作） =====
		fundPools := auth.Group("/fund-pools")
		{
			fundPools.GET("", fundPoolHandler.List)                 // 基金池列表
			fundPools.GET("/:id/shares", fundPoolHandler.GetShares) // 基金池份额详情
		}

		// ===== 系统配置（读操作） =====
		configs := auth.Group("/configs")
		{
			configs.GET("", configHandler.GetAll)                              // 所有配置
			configs.GET("/commission-rates", configHandler.GetCommissionRates) // 提成比例配置
			configs.GET("/:key", configHandler.Get)                            // 单个配置
		}

		// ===== 操作日志 =====
		auth.GET("/operation-logs", operationLogHandler.List) // 操作日志列表

		// ===== 文件上传 =====
		auth.POST("/upload/image", uploadHandler.UploadImage) // 上传图片

		// ===== 门店管理 =====
		auth.GET("/stores", storeHandler.List) // 门店列表

		// ===== 仪表盘 =====
		dashboard := auth.Group("/dashboard")
		{
			dashboard.GET("/overview", dashboardHandler.GetOverview) // 仪表盘概览
		}

		// ===== 数据报表 =====
		reports := auth.Group("/reports")
		{
			// 销售报表
			reports.GET("/sales/summary", reportHandler.GetSalesSummary)       // 销售总览
			reports.GET("/sales/trend", reportHandler.GetSalesTrend)           // 销售趋势
			reports.GET("/sales/ranking", reportHandler.GetPerformanceRanking) // 业绩排行

			// 利润报表
			reports.GET("/profit/analysis", reportHandler.GetProfitAnalysis) // 利润分析

			// 回款报表
			reports.GET("/payment/analysis", reportHandler.GetPaymentAnalysis) // 回款分析

			// 库存报表
			reports.GET("/inventory/analysis", reportHandler.GetInventoryAnalysis) // 库存分析

			// 提成报表
			reports.GET("/commission/analysis", reportHandler.GetCommissionAnalysis) // 提成分析
		}
	}

	// 管理员路由（基于权限码控制）
	admin := r.Group("/api")
	admin.Use(middleware.Auth())
	admin.Use(middleware.CSRF())
	admin.Use(middleware.OperationLog(operationLogService, db))
	// RequireAdmin 保留作为兜底：拥有 admin 角色或任意管理权限即可访问
	admin.Use(middleware.RequireRoleOrPermission(
		[]string{"admin"},
		[]string{"user:create", "role:manage", "category:manage", "product:manage",
			"purchase:manage", "gift:manage", "supplier:manage", "transfer:manage",
			"stock_alert:manage", "order:approve", "payment:manage",
			"peer:manage", "commission:manage", "salary:manage", "fund_pool:manage",
			"config:manage", "store:manage"},
	))
	{
		// 用户创建和删除
		admin.POST("/users", middleware.RequirePermission("user:create"), userHandler.Create)
		admin.DELETE("/users/:id", middleware.RequirePermission("user:delete"), userHandler.Delete)

		// 角色管理（创建、更新、删除、分配权限）
		admin.POST("/roles", middleware.RequirePermission("role:manage"), roleHandler.Create)
		admin.PUT("/roles/:id", middleware.RequirePermission("role:manage"), roleHandler.Update)
		admin.DELETE("/roles/:id", middleware.RequirePermission("role:manage"), roleHandler.Delete)
		admin.POST("/roles/:id/permissions", middleware.RequirePermission("role:manage"), roleHandler.AssignPermissions)

		// ===== 品类管理（写操作） =====
		admin.POST("/categories", middleware.RequirePermission("category:manage"), categoryHandler.Create)
		admin.PUT("/categories/:id", middleware.RequirePermission("category:manage"), categoryHandler.Update)
		admin.DELETE("/categories/:id", middleware.RequirePermission("category:manage"), categoryHandler.Delete)

		// ===== 商品管理（写操作） =====
		admin.POST("/products", middleware.RequirePermission("product:manage"), productHandler.Create)
		admin.POST("/products/import", middleware.RequirePermission("product:manage"), productHandler.Import)
		admin.GET("/products/import-template", productHandler.DownloadTemplate)
		admin.PUT("/products/:id", middleware.RequirePermission("product:manage"), productHandler.Update)
		admin.PUT("/products/:id/status", middleware.RequirePermission("product:manage"), productHandler.UpdateStatus)
		admin.DELETE("/products/:id", middleware.RequirePermission("product:manage"), productHandler.Delete)
		admin.POST("/products/:id/skus", middleware.RequirePermission("product:manage"), productHandler.CreateSKU)

		// ===== 采购管理（写操作） =====
		// 创建采购单和审核已移到auth组，让财务角色可以访问
		admin.PUT("/purchases/:id", middleware.RequirePermission("purchase:manage"), purchaseHandler.UpdateOrder)
		admin.POST("/purchases/:id/items", middleware.RequirePermission("purchase:manage"), purchaseHandler.AppendItems)
		admin.PUT("/purchases/:id/receipt", middleware.RequirePermission("purchase:manage"), purchaseHandler.ConfirmReceipt)
		admin.PUT("/purchases/:id/cancel", middleware.RequirePermission("purchase:manage"), purchaseHandler.CancelOrder)

		// ===== 回货单管理 =====
		admin.POST("/receipts", middleware.RequirePermission("purchase:manage"), receiptHandler.CreateReceipt)           // 创建回货单
		admin.GET("/receipts", receiptHandler.ListReceipts)                                                           // 回货单列表
		admin.GET("/receipts/:id", receiptHandler.GetReceiptDetail)                                                   // 回货单详情
		admin.PUT("/receipts/:id/approve", middleware.RequirePermission("purchase:manage"), receiptHandler.ApproveReceipt) // 审核回货单
		admin.PUT("/receipts/:id/receive", middleware.RequirePermission("purchase:manage"), receiptHandler.ReceiveReceipt) // 入库操作
		admin.PUT("/receipts/:id/cancel", middleware.RequirePermission("purchase:manage"), receiptHandler.CancelReceipt)   // 取消回货单

		// ===== 礼品管理（写操作） =====
		admin.POST("/gifts", middleware.RequirePermission("gift:manage"), giftHandler.Create)
		admin.PUT("/gifts/:id", middleware.RequirePermission("gift:manage"), giftHandler.Update)
		admin.DELETE("/gifts/:id", middleware.RequirePermission("gift:manage"), giftHandler.Delete)
		admin.POST("/gifts/:id/stock", middleware.RequirePermission("gift:manage"), giftHandler.AddStock)

		// ===== 供应商管理（写操作） =====
		admin.POST("/suppliers", middleware.RequirePermission("supplier:manage"), supplierHandler.Create)
		admin.PUT("/suppliers/:id", middleware.RequirePermission("supplier:manage"), supplierHandler.Update)
		admin.DELETE("/suppliers/:id", middleware.RequirePermission("supplier:manage"), supplierHandler.Delete)
		admin.POST("/suppliers/:id/products", middleware.RequirePermission("supplier:manage"), supplierHandler.AddProduct)
		admin.DELETE("/suppliers/:id/products/:sku_id", middleware.RequirePermission("supplier:manage"), supplierHandler.RemoveProduct)

		// ===== 调拨管理（写操作） =====
		admin.POST("/transfers", middleware.RequirePermission("transfer:manage"), transferHandler.CreateOrder)
		admin.PUT("/transfers/:id/approve", middleware.RequirePermission("transfer:manage"), transferHandler.ApproveOrder)
		admin.PUT("/transfers/:id/out", middleware.RequirePermission("transfer:manage"), transferHandler.ConfirmOut)
		admin.PUT("/transfers/:id/in", middleware.RequirePermission("transfer:manage"), transferHandler.ConfirmIn)
		admin.PUT("/transfers/:id/cancel", middleware.RequirePermission("transfer:manage"), transferHandler.CancelOrder)

		// ===== 库存预警（写操作） =====
		admin.POST("/stock-alerts/check", middleware.RequirePermission("stock_alert:manage"), stockAlertHandler.CheckAlerts)
		admin.PUT("/stock-alerts/:id", middleware.RequirePermission("stock_alert:manage"), stockAlertHandler.Handle)

		// ===== 订单管理（写操作） =====
		// 审核订单相关路由已移到auth组，让财务/主管/店长角色可以访问

		// ===== 同行管理（写操作） =====
		admin.PUT("/peers/:id", middleware.RequirePermission("peer:manage"), peerHandler.Update)    // 更新同行
		admin.DELETE("/peers/:id", middleware.RequirePermission("peer:manage"), peerHandler.Delete) // 删除同行

		// ===== 提成管理（写操作） =====
		admin.POST("/commissions/calculate/:order_id", middleware.RequirePermission("commission:manage"), commissionHandler.CalculateOrderCommission)     // 计算订单提成
		admin.POST("/commissions/adjust", middleware.RequirePermission("commission:manage"), commissionHandler.ManualAdjust)                              // 手工调整提成
		admin.POST("/commissions/fixed-calculate", middleware.RequirePermission("commission:manage"), commissionHandler.CalculateFixedCommission)         // 手动触发固定提成计算
		admin.POST("/commissions/settle-monthly", middleware.RequirePermission("commission:manage"), commissionHandler.ManualMonthlySettlement)           // 手动触发月度结算
		admin.POST("/commissions/recalculate/:order_id", middleware.RequirePermission("commission:manage"), commissionHandler.RecalculateOrderCommission) // 重新计算订单提成

		// ===== 工资管理（写操作） =====
		admin.POST("/salaries/generate", middleware.RequirePermission("salary:manage"), salaryHandler.GenerateSalary)   // 生成月度工资
		admin.POST("/salaries/:id/confirm", middleware.RequirePermission("salary:manage"), salaryHandler.ConfirmSalary) // 审核确认工资
		admin.POST("/salaries/:id/pay", middleware.RequirePermission("salary:manage"), salaryHandler.PaySalary)         // 发放工资

		// ===== 基金池（写操作） =====
		admin.POST("/fund-pools/settle", middleware.RequirePermission("fund_pool:manage"), fundPoolHandler.SettleFundPool) // 基金池结算

		// ===== 系统配置（写操作） =====
		admin.POST("/configs", middleware.RequirePermission("config:manage"), configHandler.Set)        // 设置系统配置
		admin.PUT("/configs/:key", middleware.RequirePermission("config:manage"), configHandler.Update) // 更新系统配置

		// ===== 门店管理（写操作） =====
		admin.POST("/stores", middleware.RequirePermission("store:manage"), storeHandler.Create)       // 创建门店
		admin.PUT("/stores/:id", middleware.RequirePermission("store:manage"), storeHandler.Update)    // 更新门店
		admin.DELETE("/stores/:id", middleware.RequirePermission("store:manage"), storeHandler.Delete) // 删除门店

		// ===== APP版本管理 =====
		admin.GET("/app-versions", middleware.RequirePermission("app_version:manage"), appVersionHandler.List)          // 版本列表
		admin.GET("/app-versions/:id", middleware.RequirePermission("app_version:manage"), appVersionHandler.GetByID)   // 版本详情
		admin.POST("/app-versions", middleware.RequirePermission("app_version:manage"), appVersionHandler.Create)       // 创建版本
		admin.PUT("/app-versions/:id", middleware.RequirePermission("app_version:manage"), appVersionHandler.Update)    // 更新版本
		admin.DELETE("/app-versions/:id", middleware.RequirePermission("app_version:manage"), appVersionHandler.Delete) // 删除版本

		// ===== 平台维护 =====
		admin.GET("/backups", middleware.RequirePermission("maintenance:manage"), maintenanceHandler.ListBackups)                                // 备份列表
		admin.POST("/backups", middleware.RequirePermission("maintenance:manage"), maintenanceHandler.CreateBackup)                              // 创建备份
		admin.DELETE("/backups/:id", middleware.RequirePermission("maintenance:manage"), maintenanceHandler.DeleteBackup)                        // 删除备份
		admin.POST("/backups/:id/restore", middleware.RequirePermission("maintenance:manage"), maintenanceHandler.RestoreBackup)                 // 还原备份
		admin.GET("/maintenance/data-tables", middleware.RequirePermission("maintenance:manage"), maintenanceHandler.GetDataTables)              // 可清除数据表列表
		admin.POST("/maintenance/check-recent-backup", middleware.RequirePermission("maintenance:manage"), maintenanceHandler.CheckRecentBackup) // 检查近期备份
		admin.POST("/maintenance/clear-data", middleware.RequirePermission("maintenance:manage"), maintenanceHandler.ClearData)                  // 清除业务数据
	}

	// 公开API：获取最新APP版本（无需登录）
	r.GET("/api/app-versions/latest", appVersionHandler.GetLatest)

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

// getUploadDir 获取上传文件存储目录
func getUploadDir() string {
	// 默认上传目录
	defaultDir := "./uploads"

	if configs.GlobalConfig == nil {
		log.Println("[Upload] 警告：全局配置未加载，使用默认上传目录")
		return defaultDir
	}

	if configs.GlobalConfig.Upload.Dir == "" {
		log.Println("[Upload] 警告：配置文件中未设置upload.dir，使用默认上传目录")
		return defaultDir
	}

	return configs.GlobalConfig.Upload.Dir
}

// getUploadBaseURL 获取上传文件访问的基础URL
func getUploadBaseURL() string {
	// 默认基础URL（开发环境）
	defaultBaseURL := "http://localhost:8080/uploads"

	if configs.GlobalConfig == nil {
		log.Println("[Upload] 警告：全局配置未加载，使用默认基础URL")
		return defaultBaseURL
	}

	if configs.GlobalConfig.Upload.BaseURL == "" {
		log.Println("[Upload] 警告：配置文件中未设置upload.base_url，使用默认基础URL")
		return defaultBaseURL
	}

	return configs.GlobalConfig.Upload.BaseURL
}
