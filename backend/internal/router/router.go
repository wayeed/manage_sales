package router

import (
	"furniture-commission/internal/handler"
	"furniture-commission/internal/middleware"
	"furniture-commission/internal/pkg/database"
	"furniture-commission/internal/repository"
	"furniture-commission/internal/service"
	"os"

	"github.com/gin-gonic/gin"
	_ "furniture-commission/docs"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
)

// SetupRouter 注册路由
func SetupRouter() *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Cors())
	r.Use(middleware.Logger())

	// 静态文件服务（上传的文件）
	r.Static("/uploads", "./uploads")

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
	productService := service.NewProductService(db, productRepo, configRepo)
	skuService := service.NewSKUService(db, skuRepo)
	warehouseService := service.NewWarehouseService(warehouseRepo)
	inventoryService := service.NewInventoryService(db, inventoryRepo, skuRepo, warehouseRepo)
	purchaseService := service.NewPurchaseService(db, inventoryService)
	giftService := service.NewGiftService(db, inventoryService)
	supplierService := service.NewSupplierService(db)
	transferService := service.NewTransferService(db, inventoryService)
	stockAlertService := service.NewStockAlertService(db, inventoryRepo)

	// 订单管理 Service
	orderService := service.NewOrderService(db, orderRepo, paymentRepo, customerRepo, peerRepo, inventoryService)
	customerService := service.NewCustomerService(db, customerRepo)
	peerService := service.NewPeerService(db, peerRepo, userRepo)

	// 提成与工资管理 Service（需要在 paymentService 之前初始化）
	configService := service.NewConfigService(configRepo)
	commissionService := service.NewCommissionService(db, commissionRepo, orderRepo, referralRepo, configService)

	// 回款 Service（依赖 commissionService）
	paymentService := service.NewPaymentService(db, paymentRepo, orderRepo, commissionService)
	fundPoolService := service.NewFundPoolService(db, fundPoolRepo, commissionRepo, configService)
	salaryService := service.NewSalaryService(db, salaryRepo, commissionRepo, fundPoolRepo)

	// 数据分析与报表 Service
	dashboardService := service.NewDashboardService(db)
	reportService := service.NewReportService(db)

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
	giftHandler := handler.NewGiftHandler(giftService)
	supplierHandler := handler.NewSupplierHandler(supplierService)
	transferHandler := handler.NewTransferHandler(transferService)
	warehouseHandler := handler.NewWarehouseHandler(warehouseService)
	stockAlertHandler := handler.NewStockAlertHandler(stockAlertService)

	// 订单管理 Handler
	orderHandler := handler.NewOrderHandler(orderService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	customerHandler := handler.NewCustomerHandler(customerService)
	peerHandler := handler.NewPeerHandler(peerService)

	// 提成与工资管理 Handler
	commissionHandler := handler.NewCommissionHandler(commissionService)
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
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, 0755)
	baseURL := "http://localhost:8080/uploads"
	uploadHandler := handler.NewUploadHandler(uploadDir, baseURL)

	// 数据分析与报表 Handler
	dashboardHandler := handler.NewDashboardHandler(dashboardService, db)
	reportHandler := handler.NewReportHandler(reportService, db)

	// 健康检查（公开）
	r.GET("/api/health", func(c *gin.Context) {
		handler.Success(c, gin.H{
			"status":  "ok",
			"service": "furniture-commission",
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
	auth.Use(middleware.OperationLog(operationLogService, db))
	{
		// 登出
		auth.POST("/logout", authHandler.Logout)

		// 当前用户信息
		auth.GET("/users/me", authHandler.GetCurrentUser)

		// 用户管理
		users := auth.Group("/users")
		{
			users.GET("", userHandler.List)                         // 用户列表
			users.GET("/:id", userHandler.GetDetail)                 // 用户详情
			users.PUT("/:id", userHandler.Update)                    // 更新用户
			users.POST("/:id/reset-password", userHandler.ResetPassword) // 重置密码
			users.PUT("/:id/status", userHandler.UpdateStatus)           // 启用/禁用
			users.POST("/:id/roles", userHandler.AssignRole)         // 分配角色
		}

		// 员工列表（供选择器使用，别名到用户列表）
		auth.GET("/employees", userHandler.List)

		// 角色管理
		roles := auth.Group("/roles")
		{
			roles.GET("", roleHandler.List)                            // 角色列表
			roles.GET("/:id", roleHandler.GetDetail)                    // 角色详情
			roles.GET("/:id/permissions", permHandler.GetByRole)        // 角色权限
		}

		// 权限管理
		permissions := auth.Group("/permissions")
		{
			permissions.GET("", permHandler.List)       // 权限列表
			permissions.GET("/tree", permHandler.GetTree) // 权限树
		}

		// ===== 品类管理 =====
		categories := auth.Group("/categories")
		{
			categories.GET("", categoryHandler.List)       // 品类列表
			categories.GET("/:id", categoryHandler.List)   // 品类详情（复用List）
		}

		// ===== 商品管理 =====
		products := auth.Group("/products")
		{
			products.GET("", productHandler.List)            // 商品列表
			products.GET("/:id", productHandler.GetDetail)    // 商品详情
			products.GET("/:id/skus", productHandler.ListSKU) // 商品SKU列表
		}

		// ===== SKU管理 =====
		skus := auth.Group("/skus")
		skus.GET("", productHandler.ListAllSKU) // SKU列表（支持搜索）
		skus.GET("/with-stock", productHandler.ListSKUWithStock) // 带库存的SKU列表（订单选商品）
		{
			skus.PUT("/:id", productHandler.UpdateSKU)   // 更新SKU
			skus.DELETE("/:id", productHandler.DeleteSKU) // 删除SKU
		}

		// ===== 仓库管理 =====
		warehouses := auth.Group("/warehouses")
		{
			warehouses.GET("", warehouseHandler.List) // 仓库列表
			warehouses.POST("", warehouseHandler.Create) // 创建仓库
			warehouses.PUT(":id", warehouseHandler.Update) // 更新仓库
			warehouses.DELETE(":id", warehouseHandler.Delete) // 删除仓库
			warehouses.GET(":id", warehouseHandler.GetByID) // 仓库详情
		}

		// ===== 库存管理 =====
		inventory := auth.Group("/inventory")
		{
			inventory.GET("/stock", inventoryHandler.GetStock)     // 查询库存
			inventory.GET("/stocks", inventoryHandler.GetStockList) // 库存列表
                        inventory.GET("/transactions", inventoryHandler.GetTransactions) // 库存流水
		}

		// ===== 采购管理 =====
		purchases := auth.Group("/purchases")
		{
			purchases.GET("", purchaseHandler.List)             // 采购订单列表
			purchases.GET("/:id", purchaseHandler.GetDetail)     // 采购订单详情
		}

		// ===== 礼品管理 =====
		gifts := auth.Group("/gifts")
		{
			gifts.GET("", giftHandler.List)           // 礼品列表
			gifts.GET("/:id", giftHandler.GetDetail)   // 礼品详情
		}

		// ===== 供应商管理 =====
		suppliers := auth.Group("/suppliers")
		{
			suppliers.GET("", supplierHandler.List)               // 供应商列表
			suppliers.GET("/:id", supplierHandler.GetDetail)       // 供应商详情
			suppliers.GET("/:id/products", supplierHandler.GetProducts) // 供应商商品列表
		}

		// ===== 调拨管理 =====
		transfers := auth.Group("/transfers")
		{
			transfers.GET("", transferHandler.List)           // 调拨单列表
			transfers.GET("/:id", transferHandler.GetDetail)   // 调拨单详情
		}

		// ===== 库存预警 =====
		stockAlerts := auth.Group("/stock-alerts")
		{
			stockAlerts.GET("", stockAlertHandler.List) // 预警列表
		}

		// ===== 订单管理 =====
		orders := auth.Group("/orders")
		{
			orders.GET("", orderHandler.List)               // 订单列表
			orders.GET("/feed", orderHandler.GetOrderFeed)   // 订单动态
			orders.GET("/:id", orderHandler.GetDetail)       // 订单详情
			orders.POST("", orderHandler.CreateOrder)        // 创建订单
		}

		// ===== 回款管理 =====
		payments := auth.Group("/payments")
		{
			payments.GET("", paymentHandler.List)            // 回款列表
		}

		// ===== 客户管理 =====
		customers := auth.Group("/customers")
		{
			customers.GET("", customerHandler.List)                         // 客户列表
			customers.GET("/:id", customerHandler.GetDetail)                 // 客户详情
			customers.GET("/:id/follow-ups", customerHandler.GetFollowUps)   // 跟进记录
			customers.POST("/:id/follow-ups", customerHandler.AddFollowUp)   // 添加跟进记录
			customers.POST("", customerHandler.Create)                       // 创建客户
			customers.PUT("/:id", customerHandler.Update)                     // 更新客户
			customers.DELETE("/:id", customerHandler.Delete)                   // 删除客户
		}

		// ===== 同行管理 =====
		peers := auth.Group("/peers")
		{
			peers.GET("", peerHandler.List)       // 同行列表
			peers.GET("/:id", peerHandler.GetDetail) // 同行详情
		}

		// ===== 提成管理（读操作） =====
		commissions := auth.Group("/commissions")
		{
			commissions.GET("", commissionHandler.List)                          // 提成列表
			commissions.GET("/summary", commissionHandler.GetSummary)            // 提成汇总
			commissions.GET("/order/:order_id", commissionHandler.GetByOrderID)  // 订单提成明细
			commissions.POST("/estimate", commissionHandler.EstimateCommission)  // 预估提成
		}

		// ===== 业绩查询 =====
		performance := auth.Group("/performance")
		{
			performance.GET("/overview", performanceHandler.GetOverview) // 业绩概览
		}

		// ===== 引荐关系管理 =====
		referrals := auth.Group("/referrals")
		{
			referrals.GET("", referralHandler.List)              // 引荐关系列表
			referrals.POST("", referralHandler.Create)           // 创建引荐关系
			referrals.POST("/:id/terminate", referralHandler.Terminate) // 终止引荐关系
		}

		// ===== 申请跟进审批 =====
		followUpApprovals := auth.Group("/follow-up-approvals")
		{
			followUpApprovals.POST("", followUpApprovalHandler.Create)                    // 创建申请
			followUpApprovals.GET("/my", followUpApprovalHandler.ListMyApplications)      // 我的申请列表
			followUpApprovals.GET("/pending", followUpApprovalHandler.ListPendingApprovals) // 待我审批列表
			followUpApprovals.GET("/:id", followUpApprovalHandler.GetStatus)               // 查询状态
			followUpApprovals.POST("/:id/approve", followUpApprovalHandler.Approve)        // 审批通过
			followUpApprovals.POST("/:id/reject", followUpApprovalHandler.Reject)          // 审批拒绝
			followUpApprovals.POST("/:id/cancel", followUpApprovalHandler.Cancel)          // 撤回申请
		}

		// ===== 工资管理（读操作） =====
		salaries := auth.Group("/salaries")
		{
			salaries.GET("", salaryHandler.List)                                        // 工资列表
			salaries.GET("/:id", salaryHandler.GetDetail)                                // 工资详情
			salaries.GET("/employee/:employee_id/:salary_month", salaryHandler.GetEmployeeSalary) // 员工月度工资
		}

		// ===== 基金池（读操作） =====
		fundPools := auth.Group("/fund-pools")
		{
			fundPools.GET("", fundPoolHandler.List)                // 基金池列表
			fundPools.GET("/:id/shares", fundPoolHandler.GetShares) // 基金池份额详情
		}

		// ===== 系统配置（读操作） =====
		configs := auth.Group("/configs")
		{
			configs.GET("", configHandler.GetAll)                    // 所有配置
			configs.GET("/commission-rates", configHandler.GetCommissionRates) // 提成比例配置
			configs.GET("/:key", configHandler.Get)                  // 单个配置
		}

		// ===== 操作日志 =====
		auth.GET("/operation-logs", operationLogHandler.List) // 操作日志列表

		// ===== 文件上传 =====
		auth.POST("/upload/image", uploadHandler.UploadImage) // 上传图片

		// ===== 门店管理 =====
		auth.GET("/stores", storeHandler.List)             // 门店列表

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
			reports.GET("/profit/analysis", reportHandler.GetProfitAnalysis)   // 利润分析

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
		admin.PUT("/products/:id", middleware.RequirePermission("product:manage"), productHandler.Update)
		admin.PUT("/products/:id/status", middleware.RequirePermission("product:manage"), productHandler.UpdateStatus)
		admin.DELETE("/products/:id", middleware.RequirePermission("product:manage"), productHandler.Delete)
		admin.POST("/products/:id/skus", middleware.RequirePermission("product:manage"), productHandler.CreateSKU)

		// ===== 采购管理（写操作） =====
		admin.POST("/purchases", middleware.RequirePermission("purchase:manage"), purchaseHandler.CreateOrder)
		admin.PUT("/purchases/:id/approve", middleware.RequirePermission("purchase:manage"), purchaseHandler.ApproveOrder)
		admin.PUT("/purchases/:id/receipt", middleware.RequirePermission("purchase:manage"), purchaseHandler.ConfirmReceipt)
		admin.PUT("/purchases/:id/cancel", middleware.RequirePermission("purchase:manage"), purchaseHandler.CancelOrder)

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
		admin.POST("/orders/:id/approve", middleware.RequirePermission("order:approve"), orderHandler.ApproveOrder)    // 审核订单
		admin.POST("/orders/:id/cancel", middleware.RequirePermission("order:approve"), orderHandler.CancelOrder)      // 取消订单
		admin.POST("/orders/:id/return", middleware.RequirePermission("order:approve"), orderHandler.ReturnOrder)      // 退货处理

		// ===== 回款管理（写操作） =====
		admin.POST("/payments", middleware.RequirePermission("payment:manage"), paymentHandler.CreatePayment)            // 录入回款
		admin.POST("/payments/:id/approve", middleware.RequirePermission("payment:manage"), paymentHandler.ApprovePayment) // 审核回款


		// ===== 同行管理（写操作） =====
		admin.POST("/peers", middleware.RequirePermission("peer:manage"), peerHandler.Create)       // 创建同行
		admin.PUT("/peers/:id", middleware.RequirePermission("peer:manage"), peerHandler.Update)    // 更新同行
		admin.DELETE("/peers/:id", middleware.RequirePermission("peer:manage"), peerHandler.Delete) // 删除同行

		// ===== 提成管理（写操作） =====
		admin.POST("/commissions/calculate/:order_id", middleware.RequirePermission("commission:manage"), commissionHandler.CalculateOrderCommission) // 计算订单提成
		admin.POST("/commissions/adjust", middleware.RequirePermission("commission:manage"), commissionHandler.ManualAdjust)                         // 手工调整提成

		// ===== 工资管理（写操作） =====
		admin.POST("/salaries/generate", middleware.RequirePermission("salary:manage"), salaryHandler.GenerateSalary)   // 生成月度工资
		admin.POST("/salaries/:id/confirm", middleware.RequirePermission("salary:manage"), salaryHandler.ConfirmSalary) // 审核确认工资
		admin.POST("/salaries/:id/pay", middleware.RequirePermission("salary:manage"), salaryHandler.PaySalary)         // 发放工资

		// ===== 基金池（写操作） =====
		admin.POST("/fund-pools/settle", middleware.RequirePermission("fund_pool:manage"), fundPoolHandler.SettleFundPool) // 基金池结算

		// ===== 系统配置（写操作） =====
		admin.POST("/configs", middleware.RequirePermission("config:manage"), configHandler.Set) // 设置系统配置
		admin.PUT("/configs/:key", middleware.RequirePermission("config:manage"), configHandler.Update) // 更新系统配置

		// ===== 门店管理（写操作） =====
		admin.POST("/stores", middleware.RequirePermission("store:manage"), storeHandler.Create)         // 创建门店
		admin.PUT("/stores/:id", middleware.RequirePermission("store:manage"), storeHandler.Update)      // 更新门店
		admin.DELETE("/stores/:id", middleware.RequirePermission("store:manage"), storeHandler.Delete)   // 删除门店
	}

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
