-- ============================================================
-- 家具销售提成管理系统 - 种子数据脚本
-- 数据库: furniture_commission
-- 注意: 请在 init.sql 执行完毕后再执行本脚本
-- ============================================================

USE `furniture_commission`;

-- ============================================================
-- 1. 默认门店
-- ============================================================
INSERT INTO `stores` (`id`, `store_code`, `store_name`, `address`, `contact_phone`, `status`) VALUES
(1, 'STORE001', '默认门店', '默认地址', '400-000-0000', 1);

-- ============================================================
-- 2. 预设角色
-- ============================================================
INSERT INTO `roles` (`id`, `role_code`, `role_name`, `role_type`, `description`, `status`) VALUES
(1, 'BOSS',       '老板',     1, '系统最高权限，拥有所有功能的管理权限', 1),
(2, 'ADMIN',      '管理员',   1, '系统管理员，拥有大部分功能的管理权限', 1),
(3, 'STORE_MANAGER', '店长',  1, '门店管理，负责门店运营和团队管理', 1),
(4, 'SUPERVISOR', '主管',     1, '团队主管，负责团队管理和业务指导', 1),
(5, 'SALESMAN',   '业务员',   1, '销售人员，负责客户开发和订单跟进', 1),
(6, 'FINANCE',    '财务',     1, '财务人员，负责回款审核和工资发放', 1),
(7, 'WAREHOUSE',  '仓管',     1, '仓库管理员，负责库存管理和出入库操作', 1);

-- ============================================================
-- 3. 管理员账号
--    username: admin
--    password: 123456 (MD5: e10adc3949ba59abbe56e057f20f883e)
-- ============================================================
INSERT INTO `users` (`id`, `store_id`, `employee_no`, `username`, `password`, `real_name`, `phone`, `role`, `status`, `entry_date`, `is_formal`, `base_salary`, `created_by`) VALUES
(1, 1, 'EMP001', 'admin', 'e10adc3949ba59abbe56e057f20f883e', '系统管理员', '13800000000', 5, 1, '2024-01-01', 1, 0.00, 1);

-- 管理员关联老板角色
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES
(1, 1);

-- ============================================================
-- 4. 权限菜单数据
-- ============================================================

-- 一级菜单
INSERT INTO `permissions` (`id`, `permission_code`, `permission_name`, `permission_type`, `parent_id`, `path`, `icon`, `sort_order`, `status`) VALUES
(100, 'dashboard',       '首页概览',   1, 0, '/dashboard',       'dashboard',     1, 1),
(200, 'user_mgmt',       '用户管理',   1, 0, '/user',            'user',          2, 1),
(300, 'product_mgmt',    '商品管理',   1, 0, '/product',         'shopping',      3, 1),
(400, 'order_mgmt',      '订单管理',   1, 0, '/order',           'order',         4, 1),
(500, 'inventory_mgmt',  '库存管理',   1, 0, '/inventory',       'warehouse',     5, 1),
(600, 'commission_mgmt', '提成管理',   1, 0, '/commission',      'money',         6, 1),
(700, 'salary_mgmt',     '工资管理',   1, 0, '/salary',          'pay-circle',    7, 1),
(800, 'data_analysis',   '数据分析',   1, 0, '/analysis',        'bar-chart',     8, 1),
(900, 'system_config',   '系统配置',   1, 0, '/system',          'setting',       9, 1);

-- 用户管理子菜单和按钮
INSERT INTO `permissions` (`id`, `permission_code`, `permission_name`, `permission_type`, `parent_id`, `path`, `icon`, `sort_order`, `status`) VALUES
(201, 'user_list',        '用户列表',   1, 200, '/user/list',        NULL, 1, 1),
(202, 'user_add',         '新增用户',   2, 200, NULL,                NULL, 2, 1),
(203, 'user_edit',        '编辑用户',   2, 200, NULL,                NULL, 3, 1),
(204, 'user_delete',      '删除用户',   2, 200, NULL,                NULL, 4, 1),
(205, 'user_view',        '查看用户',   2, 200, NULL,                NULL, 5, 1),
(206, 'user_role_assign', '分配角色',   2, 200, NULL,                NULL, 6, 1),
(207, 'user_reset_pwd',   '重置密码',   2, 200, NULL,                NULL, 7, 1),
(208, 'user_enable',      '启用/禁用',  2, 200, NULL,                NULL, 8, 1);

-- 商品管理子菜单和按钮
INSERT INTO `permissions` (`id`, `permission_code`, `permission_name`, `permission_type`, `parent_id`, `path`, `icon`, `sort_order`, `status`) VALUES
(301, 'category_list',    '品类管理',   1, 300, '/product/category',  NULL, 1, 1),
(302, 'category_add',     '新增品类',   2, 300, NULL,                NULL, 2, 1),
(303, 'category_edit',    '编辑品类',   2, 300, NULL,                NULL, 3, 1),
(304, 'category_delete',  '删除品类',   2, 300, NULL,                NULL, 4, 1),
(311, 'product_list',     '商品列表',   1, 300, '/product/list',      NULL, 5, 1),
(312, 'product_add',      '新增商品',   2, 300, NULL,                NULL, 6, 1),
(313, 'product_edit',     '编辑商品',   2, 300, NULL,                NULL, 7, 1),
(314, 'product_delete',   '删除商品',   2, 300, NULL,                NULL, 8, 1),
(315, 'product_view',     '查看商品',   2, 300, NULL,                NULL, 9, 1),
(316, 'product_onoff',    '上架/下架',  2, 300, NULL,                NULL, 10, 1);

-- 订单管理子菜单和按钮
INSERT INTO `permissions` (`id`, `permission_code`, `permission_name`, `permission_type`, `parent_id`, `path`, `icon`, `sort_order`, `status`) VALUES
(401, 'order_list',       '订单列表',   1, 400, '/order/list',        NULL, 1, 1),
(402, 'order_add',        '新增订单',   2, 400, NULL,                NULL, 2, 1),
(403, 'order_edit',       '编辑订单',   2, 400, NULL,                NULL, 3, 1),
(404, 'order_view',       '查看订单',   2, 400, NULL,                NULL, 4, 1),
(405, 'order_approve',    '审批订单',   2, 400, NULL,                NULL, 5, 1),
(406, 'order_cancel',     '取消订单',   2, 400, NULL,                NULL, 6, 1),
(407, 'order_return',     '订单退货',   2, 400, NULL,                NULL, 7, 1),
(411, 'payment_list',     '回款管理',   1, 400, '/order/payment',     NULL, 10, 1),
(412, 'payment_add',      '新增回款',   2, 400, NULL,                NULL, 11, 1),
(413, 'payment_audit',    '审核回款',   2, 400, NULL,                NULL, 12, 1),
(421, 'peer_list',        '同行管理',   1, 400, '/order/peer',        NULL, 20, 1),
(422, 'peer_add',         '新增同行',   2, 400, NULL,                NULL, 21, 1),
(423, 'peer_edit',        '编辑同行',   2, 400, NULL,                NULL, 22, 1),
(424, 'peer_delete',      '删除同行',   2, 400, NULL,                NULL, 23, 1);

-- 库存管理子菜单和按钮
INSERT INTO `permissions` (`id`, `permission_code`, `permission_name`, `permission_type`, `parent_id`, `path`, `icon`, `sort_order`, `status`) VALUES
(501, 'warehouse_list',   '仓库管理',   1, 500, '/inventory/warehouse', NULL, 1, 1),
(502, 'warehouse_add',    '新增仓库',   2, 500, NULL,                NULL, 2, 1),
(503, 'warehouse_edit',   '编辑仓库',   2, 500, NULL,                NULL, 3, 1),
(511, 'stock_list',       '库存查询',   1, 500, '/inventory/stock',   NULL, 5, 1),
(512, 'stock_view',       '查看库存',   2, 500, NULL,                NULL, 6, 1),
(521, 'purchase_list',    '采购管理',   1, 500, '/inventory/purchase', NULL, 10, 1),
(522, 'purchase_add',     '新增采购',   2, 500, NULL,                NULL, 11, 1),
(523, 'purchase_audit',   '审核采购',   2, 500, NULL,                NULL, 12, 1),
(524, 'purchase_inbound', '采购入库',   2, 500, NULL,                NULL, 13, 1),
(531, 'transfer_list',    '调拨管理',   1, 500, '/inventory/transfer', NULL, 20, 1),
(532, 'transfer_add',     '新增调拨',   2, 500, NULL,                NULL, 21, 1),
(533, 'transfer_audit',   '审核调拨',   2, 500, NULL,                NULL, 22, 1),
(541, 'stocktaking_list', '盘点管理',   1, 500, '/inventory/stocktaking', NULL, 30, 1),
(542, 'stocktaking_add',  '新增盘点',   2, 500, NULL,                NULL, 31, 1),
(551, 'gift_list',        '礼品管理',   1, 500, '/inventory/gift',    NULL, 40, 1),
(552, 'gift_add',         '新增礼品',   2, 500, NULL,                NULL, 41, 1),
(553, 'gift_edit',        '编辑礼品',   2, 500, NULL,                NULL, 42, 1),
(554, 'gift_delete',      '删除礼品',   2, 500, NULL,                NULL, 43, 1),
(561, 'stock_alert_list', '库存预警',   1, 500, '/inventory/alert',   NULL, 50, 1),
(562, 'stock_alert_handle','处理预警',  2, 500, NULL,                NULL, 51, 1);

-- 提成管理子菜单和按钮
INSERT INTO `permissions` (`id`, `permission_code`, `permission_name`, `permission_type`, `parent_id`, `path`, `icon`, `sort_order`, `status`) VALUES
(601, 'commission_list',  '提成明细',   1, 600, '/commission/list',   NULL, 1, 1),
(602, 'commission_view',  '查看提成',   2, 600, NULL,                NULL, 2, 1),
(603, 'commission_settle','提成结算',   2, 600, NULL,                NULL, 3, 1),
(611, 'fund_pool_list',   '基金池管理', 1, 600, '/commission/fund',   NULL, 10, 1),
(612, 'fund_pool_settle', '基金池结算', 2, 600, NULL,                NULL, 11, 1),
(613, 'fund_pool_distribute', '基金池发放', 2, 600, NULL,            NULL, 12, 1),
(621, 'referral_list',    '老带新管理', 1, 600, '/commission/referral', NULL, 20, 1),
(622, 'referral_add',     '建立关系',   2, 600, NULL,                NULL, 21, 1),
(623, 'referral_end',     '结束关系',   2, 600, NULL,                NULL, 22, 1);

-- 工资管理子菜单和按钮
INSERT INTO `permissions` (`id`, `permission_code`, `permission_name`, `permission_type`, `parent_id`, `path`, `icon`, `sort_order`, `status`) VALUES
(701, 'salary_list',      '工资列表',   1, 700, '/salary/list',       NULL, 1, 1),
(702, 'salary_generate',  '生成工资',   2, 700, NULL,                NULL, 2, 1),
(703, 'salary_view',      '查看工资',   2, 700, NULL,                NULL, 3, 1),
(704, 'salary_confirm',   '确认工资',   2, 700, NULL,                NULL, 4, 1),
(705, 'salary_pay',       '发放工资',   2, 700, NULL,                NULL, 5, 1),
(706, 'salary_export',    '导出工资',   2, 700, NULL,                NULL, 6, 1);

-- 数据分析子菜单和按钮
INSERT INTO `permissions` (`id`, `permission_code`, `permission_name`, `permission_type`, `parent_id`, `path`, `icon`, `sort_order`, `status`) VALUES
(801, 'sales_report',     '销售报表',   1, 800, '/analysis/sales',    NULL, 1, 1),
(802, 'profit_report',    '利润报表',   1, 800, '/analysis/profit',   NULL, 2, 1),
(803, 'commission_report','提成报表',   1, 800, '/analysis/commission', NULL, 3, 1),
(804, 'employee_report',  '员工业绩',   1, 800, '/analysis/employee', NULL, 4, 1),
(805, 'customer_report',  '客户分析',   1, 800, '/analysis/customer', NULL, 5, 1),
(806, 'inventory_report', '库存报表',   1, 800, '/analysis/inventory', NULL, 6, 1),
(807, 'report_export',    '导出报表',   2, 800, NULL,                NULL, 10, 1);

-- 系统配置子菜单和按钮
INSERT INTO `permissions` (`id`, `permission_code`, `permission_name`, `permission_type`, `parent_id`, `path`, `icon`, `sort_order`, `status`) VALUES
(901, 'store_list',       '门店管理',   1, 900, '/system/store',      NULL, 1, 1),
(902, 'store_add',        '新增门店',   2, 900, NULL,                NULL, 2, 1),
(903, 'store_edit',       '编辑门店',   2, 900, NULL,                NULL, 3, 1),
(911, 'role_list',        '角色管理',   1, 900, '/system/role',       NULL, 5, 1),
(912, 'role_add',         '新增角色',   2, 900, NULL,                NULL, 6, 1),
(913, 'role_edit',        '编辑角色',   2, 900, NULL,                NULL, 7, 1),
(914, 'role_delete',      '删除角色',   2, 900, NULL,                NULL, 8, 1),
(915, 'role_perm_assign', '分配权限',   2, 900, NULL,                NULL, 9, 1),
(921, 'config_list',      '参数配置',   1, 900, '/system/config',     NULL, 10, 1),
(922, 'config_edit',      '编辑配置',   2, 900, NULL,                NULL, 11, 1),
(931, 'supplier_list',    '供应商管理', 1, 900, '/system/supplier',   NULL, 15, 1),
(932, 'supplier_add',     '新增供应商', 2, 900, NULL,                NULL, 16, 1),
(933, 'supplier_edit',    '编辑供应商', 2, 900, NULL,                NULL, 17, 1),
(934, 'supplier_delete',  '删除供应商', 2, 900, NULL,                NULL, 18, 1),
(941, 'log_list',         '操作日志',   1, 900, '/system/log',        NULL, 20, 1),
(942, 'log_view',         '查看日志',   2, 900, NULL,                NULL, 21, 1),
(943, 'log_export',       '导出日志',   2, 900, NULL,                NULL, 22, 1);

-- ============================================================
-- 5. 系统配置 - 提成比例
-- ============================================================
INSERT INTO `system_configs` (`id`, `config_key`, `config_value`, `config_type`, `remark`, `sort`) VALUES
(1,  'commission_rate_level1_single',  '0.0800', 'commission', '初级业务员-单品提成比例(8%)',  10),
(2,  'commission_rate_level1_multi',   '0.1000', 'commission', '初级业务员-多品提成比例(10%)', 11),
(3,  'commission_rate_level1_remark',  '建议底薪3000-4000', 'commission', '初级业务员-备注',         12),
(4,  'commission_rate_level2_single',  '0.1800', 'commission', '中级业务员-单品提成比例(18%)', 20),
(5,  'commission_rate_level2_multi',   '0.2200', 'commission', '中级业务员-多品提成比例(22%)', 21),
(6,  'commission_rate_level2_remark',  '建议底薪1500-2500', 'commission', '中级业务员-备注',         22),
(7,  'commission_rate_level3_single',  '0.3500', 'commission', '高级业务员-单品提成比例(35%)', 30),
(8,  'commission_rate_level3_multi',   '0.3800', 'commission', '高级业务员-多品提成比例(38%)', 31),
(9,  'commission_rate_level3_remark',  '建议底薪0',          'commission', '高级业务员-备注',         32),
(10, 'commission_rate_peer_single',    '0.1000', 'commission', '同行单品提成比例(10%)',        40),
(11, 'commission_rate_peer_multi',     '0.1200', 'commission', '同行多品提成比例(12%)',        41),
(12, 'commission_rate_peer_special',   '0.0800', 'commission', '同行特批提成比例(8%)',         42),
(13, 'fund_pool_extract_rate',         '0.0500', 'commission', '基金池提取比例(5%)',           50),
(14, 'team_share_rate_manager','0.0200', 'commission', '主管团队分润比例(2%)',         51),
(15, 'team_share_rate_store',   '0.0300', 'commission', '店长团队分润比例(3%)',         52),
(16, 'referral_reward_rate',          '0.0100', 'commission', '老带新奖励比例(1%)',           53),
(17, 'fixed_commission_rate',          '0.0500', 'commission', '固定提成比例(月度回款提成)',    60),
(18, 'default_store_id',              '1',      'system',     '默认门店ID',                    0),
(19, 'salary_base_date',              '25',     'salary',     '工资计算截止日(每月25日)',      0),
(20, 'probation_commission_rate',     '0.8000', 'commission', '试用期提成系数(80%)',           0),
(21, 'cost_rate',                     '1.2000', 'product',    '成本系数(进货价×成本系数=成本价)', 0),
(22, 'min_discount_rate',             '0.9000', 'product',    '最低折扣系数(挂牌价×折扣系数=最低价)', 0);

-- ============================================================
-- 6. 默认仓库
-- ============================================================
INSERT INTO `warehouses` (`id`, `store_id`, `warehouse_code`, `warehouse_name`, `warehouse_type`, `address`, `contact_person`, `contact_phone`, `status`) VALUES
(1, 1, 'WH001', '主仓库', 1, '默认仓库地址', '仓管员', '13800000001', 1);

-- ============================================================
-- 7. 测试数据 - 品类
-- ============================================================
INSERT INTO `categories` (`id`, `store_id`, `category_code`, `category_name`, `sort_order`, `status`) VALUES
(1, 1, 'CAT001', '沙发', 1, 1),
(2, 1, 'CAT002', '床',   2, 1);

-- ============================================================
-- 8. 测试数据 - 商品
-- ============================================================
INSERT INTO `products` (`id`, `store_id`, `category_id`, `product_code`, `product_name`, `brand`, `list_price`, `min_price`, `reference_cost`, `total_cost_rate`, `warning_stock`, `status`, `created_by`) VALUES
(1, 1, 1, 'PRD001', '现代简约布艺沙发',   '舒适家', 5999.00, 4500.00, 2000.00, 1.2000, 5, 1, 1),
(2, 1, 1, 'PRD002', '北欧真皮沙发',       '舒适家', 12999.00, 9500.00, 5000.00, 1.2000, 3, 1, 1),
(3, 1, 2, 'PRD003', '实木框架双人床',     '安睡宝', 3999.00, 2800.00, 1200.00, 1.1500, 5, 1, 1),
(4, 1, 2, 'PRD004', '乳胶床垫1.8m',       '安睡宝', 2999.00, 2000.00, 900.00,  1.1500, 10, 1, 1);

-- ============================================================
-- 9. 测试数据 - SKU
-- ============================================================
INSERT INTO `product_skus` (`id`, `product_id`, `sku_code`, `sku_name`, `attributes`, `barcode`, `status`) VALUES
(1, 1, 'SKU001', '现代简约布艺沙发-灰色三人位',  '{"color":"灰色","size":"三人位(2.2m)"}',  '6901234567001', 1),
(2, 1, 'SKU002', '现代简约布艺沙发-米色三人位',  '{"color":"米色","size":"三人位(2.2m)"}',  '6901234567002', 1),
(3, 2, 'SKU003', '北欧真皮沙发-棕色L型',        '{"color":"棕色","size":"L型(3.5m)"}',     '6901234567003', 1),
(4, 2, 'SKU004', '北欧真皮沙发-黑色三人位',      '{"color":"黑色","size":"三人位(2.4m)"}',  '6901234567004', 1),
(5, 3, 'SKU005', '实木框架双人床-1.5m',          '{"size":"1.5m"}',                         '6901234567005', 1),
(6, 3, 'SKU006', '实木框架双人床-1.8m',          '{"size":"1.8m"}',                         '6901234567006', 1),
(7, 4, 'SKU007', '乳胶床垫1.8m-加厚款',          '{"thickness":"加厚(25cm)"}',               '6901234567007', 1),
(8, 4, 'SKU008', '乳胶床垫1.8m-标准款',          '{"thickness":"标准(20cm)"}',               '6901234567008', 1);

-- ============================================================
-- 10. 测试数据 - 仓库商品库存初始化
-- ============================================================
INSERT INTO `warehouse_stocks` (`warehouse_id`, `sku_id`, `stock_quantity`, `available_quantity`, `locked_quantity`, `warning_stock`) VALUES
(1, 1, 10, 10, 0, 5),
(1, 2,  8,  8, 0, 5),
(1, 3,  3,  3, 0, 3),
(1, 4,  5,  5, 0, 3),
(1, 5,  6,  6, 0, 5),
(1, 6,  4,  4, 0, 5),
(1, 7, 15, 15, 0, 10),
(1, 8, 12, 12, 0, 10);

-- ============================================================
-- 11. 测试数据 - 供应商
-- ============================================================
INSERT INTO `suppliers` (`id`, `store_id`, `supplier_code`, `supplier_name`, `contact_person`, `contact_phone`, `address`, `business_scope`, `status`) VALUES
(1, 1, 'SUP001', '舒适家家具制造有限公司', '张经理', '13900000001', '广东省佛山市顺德区XX路XX号', '沙发、软体家具', 1),
(2, 1, 'SUP002', '安睡宝家居用品有限公司', '李经理', '13900000002', '广东省东莞市厚街镇XX路XX号', '床具、床垫', 1);

-- ============================================================
-- 12. 测试数据 - 供应商商品关联
-- ============================================================
INSERT INTO `supplier_products` (`supplier_id`, `sku_id`, `supply_price`, `min_order_quantity`, `lead_time`, `is_default`, `status`) VALUES
(1, 1, 2000.00, 3, 15, 1, 1),
(1, 2, 2000.00, 3, 15, 1, 1),
(1, 3, 5000.00, 2, 20, 1, 1),
(1, 4, 5000.00, 2, 20, 1, 1),
(2, 5, 1200.00, 5, 10, 1, 1),
(2, 6, 1200.00, 5, 10, 1, 1),
(2, 7,  900.00, 5,  7, 1, 1),
(2, 8,  900.00, 5,  7, 1, 1);

-- ============================================================
-- 13. 测试数据 - 库存批次（模拟采购入库）
-- ============================================================
INSERT INTO `inventory_batches` (`id`, `sku_id`, `batch_no`, `purchase_order_id`, `purchase_price`, `total_cost`, `initial_quantity`, `remaining_quantity`, `warehouse_id`, `status`, `entry_date`) VALUES
(1, 1, 'BATCH202401001', NULL, 2000.00, 24000.00, 10, 10, 1, 1, '2024-01-15'),
(2, 2, 'BATCH202401002', NULL, 2000.00, 19200.00,  8,  8, 1, 1, '2024-01-15'),
(3, 3, 'BATCH202401003', NULL, 5000.00, 18000.00,  3,  3, 1, 1, '2024-01-20'),
(4, 4, 'BATCH202401004', NULL, 5000.00, 30000.00,  5,  5, 1, 1, '2024-01-20'),
(5, 5, 'BATCH202401005', NULL, 1200.00,  8640.00,  6,  6, 1, 1, '2024-02-01'),
(6, 6, 'BATCH202401006', NULL, 1200.00,  5520.00,  4,  4, 1, 1, '2024-02-01'),
(7, 7, 'BATCH202401007', NULL,  900.00, 13500.00, 15, 15, 1, 1, '2024-02-10'),
(8, 8, 'BATCH202401008', NULL,  900.00, 12420.00, 12, 12, 1, 1, '2024-02-10');
