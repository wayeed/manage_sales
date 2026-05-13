-- ============================================================
-- 家具销售提成管理系统 - 数据库初始化脚本
-- 数据库: furniture_commission
-- MySQL 8.0+
-- ============================================================

CREATE DATABASE IF NOT EXISTS `furniture_commission` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

USE `furniture_commission`;

-- ============================================================
-- 表1: stores - 门店表
-- ============================================================
DROP TABLE IF EXISTS `stores`;
CREATE TABLE `stores` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_code` VARCHAR(32) NOT NULL COMMENT '门店编码',
    `store_name` VARCHAR(100) NOT NULL COMMENT '门店名称',
    `address` VARCHAR(255) DEFAULT NULL COMMENT '门店地址',
    `contact_phone` VARCHAR(20) DEFAULT NULL COMMENT '联系电话',
    `manager_id` BIGINT DEFAULT NULL COMMENT '店长ID',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_store_code` (`store_code`),
    KEY `idx_manager_id` (`manager_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门店表';

-- ============================================================
-- 表2: users - 用户表
-- ============================================================
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
    `id` BIGINT NOT NULL COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `employee_no` VARCHAR(32) DEFAULT NULL COMMENT '工号',
    `username` VARCHAR(50) DEFAULT NULL COMMENT '用户名(手机号登录)',
    `password` VARCHAR(128) DEFAULT NULL COMMENT '密码',
    `real_name` VARCHAR(50) DEFAULT NULL COMMENT '真实姓名',
    `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    `email` VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
    `department_id` BIGINT DEFAULT NULL COMMENT '部门ID',
    `role` TINYINT DEFAULT NULL COMMENT '角色(1-业务员,2-主管,3-店长,4-财务,5-管理员/老板,6-仓管)',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用,2-离职)',
    `entry_date` DATE DEFAULT NULL COMMENT '入职日期',
    `probation_end_date` DATE DEFAULT NULL COMMENT '试用期结束日期',
    `is_formal` TINYINT DEFAULT 0 COMMENT '是否转正(0-否,1-是)',
    `parent_id` BIGINT DEFAULT NULL COMMENT '上级ID(主管/店长)',
    `referrer_id` BIGINT DEFAULT NULL COMMENT '推荐人ID(老带新)',
    `base_salary` DECIMAL(10,2) DEFAULT 0.00 COMMENT '基本工资',
    `id_card` VARCHAR(18) DEFAULT NULL COMMENT '身份证号',
    `bank_account` VARCHAR(50) DEFAULT NULL COMMENT '银行卡号',
    `bank_name` VARCHAR(100) DEFAULT NULL COMMENT '开户行',
    `avatar` VARCHAR(255) DEFAULT NULL COMMENT '头像URL',
    `last_login_at` DATETIME DEFAULT NULL COMMENT '最后登录时间',
    `last_login_ip` VARCHAR(50) DEFAULT NULL COMMENT '最后登录IP',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `created_by` BIGINT DEFAULT NULL COMMENT '创建人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_employee_no` (`employee_no`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_phone` (`phone`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_department_id` (`department_id`),
    KEY `idx_role` (`role`),
    KEY `idx_status` (`status`),
    KEY `idx_parent_id` (`parent_id`),
    KEY `idx_referrer_id` (`referrer_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- ============================================================
-- 表3: categories - 品类表
-- ============================================================
DROP TABLE IF EXISTS `categories`;
CREATE TABLE `categories` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `category_code` VARCHAR(32) DEFAULT NULL COMMENT '品类编码',
    `category_name` VARCHAR(50) DEFAULT NULL COMMENT '品类名称',
    `sort_order` INT DEFAULT 0 COMMENT '排序',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_category_code` (`category_code`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='品类表';

-- ============================================================
-- 表4: products - 商品表
-- ============================================================
DROP TABLE IF EXISTS `products`;
CREATE TABLE `products` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `category_id` BIGINT DEFAULT NULL COMMENT '品类ID',
    `product_code` VARCHAR(32) DEFAULT NULL COMMENT '商品编码',
    `product_name` VARCHAR(100) DEFAULT NULL COMMENT '商品名称',
    `brand` VARCHAR(50) DEFAULT NULL COMMENT '品牌',
    `product_image` VARCHAR(255) DEFAULT NULL COMMENT '商品图片',
    `description` TEXT DEFAULT NULL COMMENT '商品描述',
    `list_price` DECIMAL(12,2) DEFAULT 0.00 COMMENT '标价(零售价)',
    `min_price` DECIMAL(12,2) DEFAULT 0.00 COMMENT '最低售价',
    `reference_cost` DECIMAL(12,2) DEFAULT 0.00 COMMENT '参考成本',
    `total_cost_rate` DECIMAL(5,4) DEFAULT 1.2000 COMMENT '总成本系数(含运费安装等)',
    `warning_stock` INT DEFAULT 10 COMMENT '库存预警值',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-下架,1-上架)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `created_by` BIGINT DEFAULT NULL COMMENT '创建人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_product_code` (`product_code`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_category_id` (`category_id`),
    KEY `idx_brand` (`brand`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品表';

-- ============================================================
-- 表5: product_skus - SKU表
-- ============================================================
DROP TABLE IF EXISTS `product_skus`;
CREATE TABLE `product_skus` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `product_id` BIGINT NOT NULL COMMENT '商品ID',
    `sku_code` VARCHAR(50) DEFAULT NULL COMMENT 'SKU编码',
    `sku_name` VARCHAR(100) DEFAULT NULL COMMENT 'SKU名称',
    `attributes` JSON DEFAULT NULL COMMENT '属性(颜色/尺寸等)',
    `barcode` VARCHAR(50) DEFAULT NULL COMMENT '条形码',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sku_code` (`sku_code`),
    KEY `idx_product_id` (`product_id`),
    KEY `idx_barcode` (`barcode`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SKU表';

-- ============================================================
-- 表6: inventory_batches - 库存批次表
-- ============================================================
DROP TABLE IF EXISTS `inventory_batches`;
CREATE TABLE `inventory_batches` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `sku_id` BIGINT NOT NULL COMMENT 'SKU ID',
    `batch_no` VARCHAR(32) DEFAULT NULL COMMENT '批次号',
    `purchase_order_id` BIGINT DEFAULT NULL COMMENT '采购订单ID',
    `purchase_price` DECIMAL(12,2) DEFAULT 0.00 COMMENT '采购单价',
    `total_cost` DECIMAL(12,2) DEFAULT 0.00 COMMENT '总成本(含分摊)',
    `initial_quantity` INT DEFAULT 0 COMMENT '初始数量',
    `remaining_quantity` INT DEFAULT 0 COMMENT '剩余数量',
    `warehouse_id` BIGINT DEFAULT NULL COMMENT '仓库ID',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-已耗尽,1-可用,2-冻结)',
    `entry_date` DATE DEFAULT NULL COMMENT '入库日期',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_batch_no` (`batch_no`),
    KEY `idx_sku_id` (`sku_id`),
    KEY `idx_purchase_order_id` (`purchase_order_id`),
    KEY `idx_warehouse_id` (`warehouse_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存批次表';

-- ============================================================
-- 表7: purchase_orders - 采购订单表
-- ============================================================
DROP TABLE IF EXISTS `purchase_orders`;
CREATE TABLE `purchase_orders` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `purchase_no` VARCHAR(32) DEFAULT NULL COMMENT '采购单号',
    `supplier_id` BIGINT DEFAULT NULL COMMENT '供应商ID',
    `supplier_name` VARCHAR(100) DEFAULT NULL COMMENT '供应商名称',
    `total_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '采购总金额',
    `total_quantity` INT DEFAULT 0 COMMENT '采购总数量',
    `status` TINYINT DEFAULT 0 COMMENT '状态(0-待审核,1-已审核,2-已入库,3-已取消)',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `created_by` BIGINT DEFAULT NULL COMMENT '创建人ID',
    `audited_by` BIGINT DEFAULT NULL COMMENT '审核人ID',
    `audited_at` DATETIME DEFAULT NULL COMMENT '审核时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_purchase_no` (`purchase_no`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_supplier_id` (`supplier_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_by` (`created_by`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='采购订单表';

-- ============================================================
-- 表8: purchase_items - 采购商品明细表
-- ============================================================
DROP TABLE IF EXISTS `purchase_items`;
CREATE TABLE `purchase_items` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `purchase_order_id` BIGINT NOT NULL COMMENT '采购订单ID',
    `sku_id` BIGINT DEFAULT NULL COMMENT 'SKU ID',
    `product_name` VARCHAR(100) DEFAULT NULL COMMENT '商品名称',
    `sku_name` VARCHAR(100) DEFAULT NULL COMMENT 'SKU名称',
    `purchase_price` DECIMAL(12,2) DEFAULT 0.00 COMMENT '采购单价',
    `quantity` INT DEFAULT 0 COMMENT '数量',
    `subtotal` DECIMAL(12,2) DEFAULT 0.00 COMMENT '小计',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_purchase_order_id` (`purchase_order_id`),
    KEY `idx_sku_id` (`sku_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='采购商品明细表';

-- ============================================================
-- 表9: peers - 同行表
-- ============================================================
DROP TABLE IF EXISTS `peers`;
CREATE TABLE `peers` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `peer_name` VARCHAR(50) DEFAULT NULL COMMENT '同行姓名',
    `phone` VARCHAR(20) DEFAULT NULL COMMENT '联系电话',
    `id_card` VARCHAR(18) DEFAULT NULL COMMENT '身份证号',
    `company` VARCHAR(100) DEFAULT NULL COMMENT '公司名称',
    `bank_account` VARCHAR(50) DEFAULT NULL COMMENT '银行卡号',
    `bank_name` VARCHAR(100) DEFAULT NULL COMMENT '开户行',
    `total_orders` INT DEFAULT 0 COMMENT '总订单数',
    `total_amount` DECIMAL(14,2) DEFAULT 0.00 COMMENT '总金额',
    `total_profit` DECIMAL(14,2) DEFAULT 0.00 COMMENT '总利润',
    `total_commission` DECIMAL(14,2) DEFAULT 0.00 COMMENT '总提成',
    `paid_commission` DECIMAL(14,2) DEFAULT 0.00 COMMENT '已付提成',
    `unpaid_commission` DECIMAL(14,2) DEFAULT 0.00 COMMENT '未付提成',
    `last_order_at` DATETIME DEFAULT NULL COMMENT '最后下单时间',
    `commission_rate` DECIMAL(5,4) DEFAULT NULL COMMENT '分成比例(如0.15表示15%,NULL表示使用系统配置)',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_phone` (`phone`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='同行表';

-- ============================================================
-- 表10: orders - 订单表（核心）
-- ============================================================
DROP TABLE IF EXISTS `orders`;
CREATE TABLE `orders` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `order_no` VARCHAR(32) DEFAULT NULL COMMENT '订单编号',
    `salesman_id` BIGINT DEFAULT NULL COMMENT '业务员ID',
    `customer_id` BIGINT DEFAULT NULL COMMENT '客户ID',
    `customer_name` VARCHAR(50) DEFAULT NULL COMMENT '客户姓名',
    `customer_phone` VARCHAR(20) DEFAULT NULL COMMENT '客户电话',
    `customer_address` VARCHAR(500) DEFAULT NULL COMMENT '客户地址',
    `source` TINYINT DEFAULT NULL COMMENT '来源(1-自然进店,2-老客户推荐,3-线上推广,4-同行介绍,5-其他)',
    `delivery_status` TINYINT DEFAULT 0 COMMENT '配送状态(0-未配送,1-配送中,2-已配送)',
    `order_type` TINYINT DEFAULT NULL COMMENT '订单类型(1-单品,2-多品,3-特殊审批,4-同行单品,5-同行多品,6-同行特批)',
    `order_status` TINYINT DEFAULT 0 COMMENT '订单状态(0-待审批,1-已生效,2-已驳回,3-已取消,4-已退货)',
    `payment_status` TINYINT DEFAULT 0 COMMENT '回款状态(0-未回款,1-部分回款,2-已回款)',
    `total_list_price` DECIMAL(12,2) DEFAULT 0.00 COMMENT '标价总额',
    `total_sale_price` DECIMAL(12,2) DEFAULT 0.00 COMMENT '售价总额',
    `discount_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '优惠金额',
    `final_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '最终金额',
    `total_cost` DECIMAL(12,2) DEFAULT 0.00 COMMENT '总成本',
    `gift_cost` DECIMAL(12,2) DEFAULT 0.00 COMMENT '礼品成本',
    `actual_profit` DECIMAL(12,2) DEFAULT 0.00 COMMENT '实际利润',
    `category_count` INT DEFAULT 0 COMMENT '品类数',
    `sku_count` INT DEFAULT 0 COMMENT 'SKU数',
    `total_quantity` INT DEFAULT 0 COMMENT '总数量',
    `paid_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '已回款金额',
    `remaining_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '未回款金额',
    `is_peer_order` TINYINT DEFAULT 0 COMMENT '是否同行订单(0-否,1-是)',
    `peer_id` BIGINT DEFAULT NULL COMMENT '同行ID',
    `is_special_approved` TINYINT DEFAULT 0 COMMENT '是否特批(0-否,1-是)',
    `approval_remark` VARCHAR(500) DEFAULT NULL COMMENT '审批备注',
    `approved_by` BIGINT DEFAULT NULL COMMENT '审批人ID',
    `approved_at` DATETIME DEFAULT NULL COMMENT '审批时间',
    `is_returned` TINYINT DEFAULT 0 COMMENT '是否退货(0-否,1-是)',
    `return_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '退货金额',
    `return_profit` DECIMAL(12,2) DEFAULT 0.00 COMMENT '退货利润',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `order_date` DATE DEFAULT NULL COMMENT '订单日期',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `created_by` BIGINT DEFAULT NULL COMMENT '创建人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_no` (`order_no`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_salesman_id` (`salesman_id`),
    KEY `idx_customer_id` (`customer_id`),
    KEY `idx_customer_phone` (`customer_phone`),
    KEY `idx_order_type` (`order_type`),
    KEY `idx_order_status` (`order_status`),
    KEY `idx_payment_status` (`payment_status`),
    KEY `idx_is_peer_order` (`is_peer_order`),
    KEY `idx_peer_id` (`peer_id`),
    KEY `idx_order_date` (`order_date`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';

-- ============================================================
-- 表11: order_items - 订单商品明细表
-- ============================================================
DROP TABLE IF EXISTS `order_items`;
CREATE TABLE `order_items` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `order_id` BIGINT NOT NULL COMMENT '订单ID',
    `sku_id` BIGINT DEFAULT NULL COMMENT 'SKU ID',
    `product_name` VARCHAR(100) DEFAULT NULL COMMENT '商品名称',
    `sku_name` VARCHAR(100) DEFAULT NULL COMMENT 'SKU名称',
    `category_id` BIGINT DEFAULT NULL COMMENT '品类ID',
    `quantity` INT DEFAULT 1 COMMENT '数量',
    `list_price` DECIMAL(12,2) DEFAULT 0.00 COMMENT '标价',
    `sale_price` DECIMAL(12,2) DEFAULT 0.00 COMMENT '售价',
    `discount_rate` DECIMAL(5,4) DEFAULT 1.0000 COMMENT '折扣率',
    `batch_id` BIGINT DEFAULT NULL COMMENT '库存批次ID',
    `unit_cost` DECIMAL(12,2) DEFAULT 0.00 COMMENT '单位成本',
    `total_cost` DECIMAL(12,2) DEFAULT 0.00 COMMENT '总成本',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_sku_id` (`sku_id`),
    KEY `idx_category_id` (`category_id`),
    KEY `idx_batch_id` (`batch_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单商品明细表';

-- ============================================================
-- 表12: order_gifts - 订单礼品表
-- ============================================================
DROP TABLE IF EXISTS `order_gifts`;
CREATE TABLE `order_gifts` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `order_id` BIGINT NOT NULL COMMENT '订单ID',
    `gift_id` BIGINT DEFAULT NULL COMMENT '礼品ID',
    `gift_name` VARCHAR(100) DEFAULT NULL COMMENT '礼品名称',
    `cost_price` DECIMAL(10,2) DEFAULT 0.00 COMMENT '成本价',
    `quantity` INT DEFAULT 1 COMMENT '数量',
    `total_cost` DECIMAL(10,2) DEFAULT 0.00 COMMENT '总成本',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_gift_id` (`gift_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单礼品表';

-- ============================================================
-- 表13: payments - 回款记录表
-- ============================================================
DROP TABLE IF EXISTS `payments`;
CREATE TABLE `payments` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `order_id` BIGINT NOT NULL COMMENT '订单ID',
    `payment_no` VARCHAR(32) DEFAULT NULL COMMENT '回款单号',
    `amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '回款金额',
    `payment_date` DATE DEFAULT NULL COMMENT '回款日期',
    `payment_method` TINYINT DEFAULT 1 COMMENT '回款方式(1-银行转账,2-现金,3-微信,4-支付宝)',
    `status` TINYINT DEFAULT 0 COMMENT '状态(0-待审核,1-已审核,2-已驳回)',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `created_by` BIGINT DEFAULT NULL COMMENT '创建人ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `audited_by` BIGINT DEFAULT NULL COMMENT '审核人ID',
    `audited_at` DATETIME DEFAULT NULL COMMENT '审核时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_payment_no` (`payment_no`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_payment_date` (`payment_date`),
    KEY `idx_status` (`status`),
    KEY `idx_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='回款记录表';

-- ============================================================
-- 表14: commissions - 提成明细表
-- ============================================================
DROP TABLE IF EXISTS `commissions`;
CREATE TABLE `commissions` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `order_id` BIGINT DEFAULT NULL COMMENT '订单ID',
    `employee_id` BIGINT DEFAULT NULL COMMENT '员工ID',
    `peer_id` BIGINT DEFAULT NULL COMMENT '同行ID',
    `commission_type` TINYINT DEFAULT NULL COMMENT '提成类型(1-业务员提成,2-同行分成,3-主管团队分润,4-店长团队分润,5-基金池奖励,6-老带新奖励)',
    `period_value` VARCHAR(7) DEFAULT NULL COMMENT '归属周期(yyyy-MM)',
    `base_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '提成基数',
    `rate` DECIMAL(5,4) DEFAULT 0.0000 COMMENT '提成比例',
    `amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '提成金额',
    `status` TINYINT DEFAULT 0 COMMENT '状态(0-待回款,1-可发放,2-已发放)',
    `settled_at` DATETIME DEFAULT NULL COMMENT '结算时间',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_employee_id` (`employee_id`),
    KEY `idx_peer_id` (`peer_id`),
    KEY `idx_commission_type` (`commission_type`),
    KEY `idx_period_value` (`period_value`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提成明细表';

-- ============================================================
-- 表15: fund_pools - 基金池表
-- ============================================================
DROP TABLE IF EXISTS `fund_pools`;
CREATE TABLE `fund_pools` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `period_type` TINYINT DEFAULT NULL COMMENT '周期类型(1-月度,2-季度,3-年度)',
    `period_value` VARCHAR(10) DEFAULT NULL COMMENT '周期值(如2024-01,2024-Q1,2024)',
    `total_profit` DECIMAL(14,2) DEFAULT 0.00 COMMENT '总利润',
    `extract_rate` DECIMAL(5,4) DEFAULT 0.0500 COMMENT '提取比例',
    `pool_amount` DECIMAL(14,2) DEFAULT 0.00 COMMENT '基金池金额',
    `total_shares` DECIMAL(12,2) DEFAULT 0.00 COMMENT '总份额',
    `per_share_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '每份金额',
    `status` TINYINT DEFAULT 0 COMMENT '状态(0-待结算,1-已结算,2-已发放)',
    `settled_at` DATETIME DEFAULT NULL COMMENT '结算时间',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_period_type` (`period_type`),
    KEY `idx_period_value` (`period_value`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='基金池表';

-- ============================================================
-- 表16: fund_pool_shares - 基金池份额表
-- ============================================================
DROP TABLE IF EXISTS `fund_pool_shares`;
CREATE TABLE `fund_pool_shares` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `fund_pool_id` BIGINT NOT NULL COMMENT '基金池ID',
    `employee_id` BIGINT NOT NULL COMMENT '员工ID',
    `personal_profit` DECIMAL(14,2) DEFAULT 0.00 COMMENT '个人利润',
    `shares` DECIMAL(10,2) DEFAULT 0.00 COMMENT '份额',
    `reward_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '奖励金额',
    `status` TINYINT DEFAULT 0 COMMENT '状态(0-待结算,1-已结算)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_fund_employee` (`fund_pool_id`, `employee_id`),
    KEY `idx_employee_id` (`employee_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='基金池份额表';

-- ============================================================
-- 表17: referral_relations - 老带新关系表
-- ============================================================
DROP TABLE IF EXISTS `referral_relations`;
CREATE TABLE `referral_relations` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `referrer_id` BIGINT NOT NULL COMMENT '推荐人(老员工)ID',
    `referred_id` BIGINT NOT NULL COMMENT '被推荐人(新员工)ID',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-已结束,1-有效)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `ended_at` DATETIME DEFAULT NULL COMMENT '结束时间',
    `ended_reason` VARCHAR(255) DEFAULT NULL COMMENT '结束原因',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_referred_id` (`referred_id`),
    KEY `idx_referrer_id` (`referrer_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='老带新关系表';

-- ============================================================
-- 表18: salary_records - 工资记录表
-- ============================================================
DROP TABLE IF EXISTS `salary_records`;
CREATE TABLE `salary_records` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `employee_id` BIGINT DEFAULT NULL COMMENT '员工ID',
    `salary_month` VARCHAR(7) DEFAULT NULL COMMENT '工资月份(yyyy-MM)',
    `base_salary` DECIMAL(12,2) DEFAULT 0.00 COMMENT '基本工资',
    `sales_commission` DECIMAL(12,2) DEFAULT 0.00 COMMENT '销售提成',
    `team_commission` DECIMAL(12,2) DEFAULT 0.00 COMMENT '团队分润',
    `fund_reward` DECIMAL(12,2) DEFAULT 0.00 COMMENT '基金池奖励',
    `referral_reward` DECIMAL(12,2) DEFAULT 0.00 COMMENT '老带新奖励',
    `deduction` DECIMAL(12,2) DEFAULT 0.00 COMMENT '扣款',
    `bonus` DECIMAL(12,2) DEFAULT 0.00 COMMENT '奖金',
    `gross_salary` DECIMAL(12,2) DEFAULT 0.00 COMMENT '应发工资',
    `net_salary` DECIMAL(12,2) DEFAULT 0.00 COMMENT '实发工资',
    `status` TINYINT DEFAULT 0 COMMENT '状态(0-草稿,1-已确认,2-已发放)',
    `paid_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '实发金额',
    `paid_at` DATETIME DEFAULT NULL COMMENT '发放时间',
    `paid_by` BIGINT DEFAULT NULL COMMENT '发放人ID',
    `pay_method` TINYINT DEFAULT NULL COMMENT '发放方式(1-银行转账,2-现金)',
    `pay_remark` VARCHAR(500) DEFAULT NULL COMMENT '发放备注',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `confirmed_by` BIGINT DEFAULT NULL COMMENT '确认人ID',
    `confirmed_at` DATETIME DEFAULT NULL COMMENT '确认时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_employee_month` (`employee_id`, `salary_month`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_salary_month` (`salary_month`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工资记录表';

-- ============================================================
-- 表19: salary_items - 工资明细表
-- ============================================================
DROP TABLE IF EXISTS `salary_items`;
CREATE TABLE `salary_items` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `salary_record_id` BIGINT NOT NULL COMMENT '工资记录ID',
    `item_type` TINYINT DEFAULT NULL COMMENT '项目类型(1-基本工资,2-销售提成,3-团队分润,4-基金池,5-老带新,6-扣款,7-奖金)',
    `item_name` VARCHAR(50) DEFAULT NULL COMMENT '项目名称',
    `amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '金额',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_salary_record_id` (`salary_record_id`),
    KEY `idx_item_type` (`item_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工资明细表';

-- ============================================================
-- 表20: gifts - 礼品表
-- ============================================================
DROP TABLE IF EXISTS `gifts`;
CREATE TABLE `gifts` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `gift_code` VARCHAR(32) DEFAULT NULL COMMENT '礼品编码',
    `gift_name` VARCHAR(100) DEFAULT NULL COMMENT '礼品名称',
    `gift_image` VARCHAR(255) DEFAULT NULL COMMENT '礼品图片',
    `cost_price` DECIMAL(10,2) DEFAULT 0.00 COMMENT '成本价',
    `stock_quantity` INT DEFAULT 0 COMMENT '库存数量',
    `warning_stock` INT DEFAULT 10 COMMENT '预警库存',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `created_by` BIGINT DEFAULT NULL COMMENT '创建人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_gift_code` (`gift_code`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='礼品表';

-- ============================================================
-- 表21: gift_inventory_batches - 礼品批次表
-- ============================================================
DROP TABLE IF EXISTS `gift_inventory_batches`;
CREATE TABLE `gift_inventory_batches` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `gift_id` BIGINT NOT NULL COMMENT '礼品ID',
    `batch_no` VARCHAR(32) DEFAULT NULL COMMENT '批次号',
    `purchase_price` DECIMAL(10,2) DEFAULT 0.00 COMMENT '采购单价',
    `initial_quantity` INT DEFAULT 0 COMMENT '初始数量',
    `remaining_quantity` INT DEFAULT 0 COMMENT '剩余数量',
    `warehouse_id` BIGINT DEFAULT NULL COMMENT '仓库ID',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-已耗尽,1-可用,2-冻结)',
    `entry_date` DATE DEFAULT NULL COMMENT '入库日期',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_batch_no` (`batch_no`),
    KEY `idx_gift_id` (`gift_id`),
    KEY `idx_warehouse_id` (`warehouse_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='礼品批次表';

-- ============================================================
-- 表22: inventory_transactions - 库存流水表
-- ============================================================
DROP TABLE IF EXISTS `inventory_transactions`;
CREATE TABLE `inventory_transactions` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `warehouse_id` BIGINT DEFAULT NULL COMMENT '仓库ID',
    `transaction_type` TINYINT DEFAULT NULL COMMENT '交易类型(1-采购入库,2-销售出库,3-调拨出库,4-调拨入库,5-盘盈,6-盘亏,7-礼品出库,8-礼品入库)',
    `biz_type` TINYINT DEFAULT NULL COMMENT '业务类型(1-商品,2-礼品)',
    `biz_id` BIGINT DEFAULT NULL COMMENT '业务ID(SKU或礼品ID)',
    `batch_id` BIGINT DEFAULT NULL COMMENT '批次ID',
    `related_order_id` BIGINT DEFAULT NULL COMMENT '关联订单ID',
    `related_purchase_id` BIGINT DEFAULT NULL COMMENT '关联采购单ID',
    `quantity` INT DEFAULT NULL COMMENT '数量',
    `before_stock` INT DEFAULT NULL COMMENT '变动前库存',
    `after_stock` INT DEFAULT NULL COMMENT '变动后库存',
    `unit_cost` DECIMAL(12,2) DEFAULT 0.00 COMMENT '单位成本',
    `total_cost` DECIMAL(12,2) DEFAULT 0.00 COMMENT '总成本',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `created_by` BIGINT DEFAULT NULL COMMENT '创建人ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_warehouse_id` (`warehouse_id`),
    KEY `idx_transaction_type` (`transaction_type`),
    KEY `idx_biz_type` (`biz_type`),
    KEY `idx_biz_id` (`biz_id`),
    KEY `idx_related_order_id` (`related_order_id`),
    KEY `idx_related_purchase_id` (`related_purchase_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存流水表';

-- ============================================================
-- 表23: customers - 客户表
-- ============================================================
DROP TABLE IF EXISTS `customers`;
CREATE TABLE `customers` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `customer_code` VARCHAR(32) DEFAULT NULL COMMENT '客户编码',
    `customer_name` VARCHAR(50) DEFAULT NULL COMMENT '客户姓名',
    `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    `email` VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
    `address` VARCHAR(255) DEFAULT NULL COMMENT '地址',
    `gender` TINYINT DEFAULT NULL COMMENT '性别(0-未知,1-男,2-女)',
    `birthday` DATE DEFAULT NULL COMMENT '生日',
    `level` TINYINT DEFAULT 1 COMMENT '客户等级(1-普通,2-VIP,3-高级VIP)',
    `total_orders` INT DEFAULT 0 COMMENT '总订单数',
    `total_amount` DECIMAL(14,2) DEFAULT 0.00 COMMENT '总消费金额',
    `total_profit` DECIMAL(14,2) DEFAULT 0.00 COMMENT '总利润贡献',
    `last_order_at` DATETIME DEFAULT NULL COMMENT '最后下单时间',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `created_by` BIGINT DEFAULT NULL COMMENT '创建人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_store_customer_code` (`store_id`, `customer_code`),
    UNIQUE KEY `uk_store_phone` (`store_id`, `phone`),
    KEY `idx_customer_name` (`customer_name`),
    KEY `idx_level` (`level`),
    KEY `idx_status` (`status`),
    KEY `idx_last_order_at` (`last_order_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客户表';

-- ============================================================
-- 表24: customer_follow_ups - 客户跟进记录表
-- ============================================================
DROP TABLE IF EXISTS `customer_follow_ups`;
CREATE TABLE `customer_follow_ups` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `customer_id` BIGINT NOT NULL COMMENT '客户ID',
    `follower_id` BIGINT DEFAULT NULL COMMENT '跟进人ID',
    `follow_type` TINYINT DEFAULT NULL COMMENT '跟进方式(1-电话,2-微信,3-上门,4-其他)',
    `content` TEXT DEFAULT NULL COMMENT '跟进内容',
    `next_follow_date` DATE DEFAULT NULL COMMENT '下次跟进日期',
    `next_follow_content` VARCHAR(500) DEFAULT NULL COMMENT '下次跟进计划',
    `is_deal` TINYINT DEFAULT 0 COMMENT '是否成交(0-否,1-是)',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_customer_id` (`customer_id`),
    KEY `idx_follower_id` (`follower_id`),
    KEY `idx_follow_type` (`follow_type`),
    KEY `idx_next_follow_date` (`next_follow_date`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客户跟进记录表';

-- ============================================================
-- 表25: suppliers - 供应商表
-- ============================================================
DROP TABLE IF EXISTS `suppliers`;
CREATE TABLE `suppliers` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `supplier_code` VARCHAR(32) DEFAULT NULL COMMENT '供应商编码',
    `supplier_name` VARCHAR(100) DEFAULT NULL COMMENT '供应商名称',
    `contact_person` VARCHAR(50) DEFAULT NULL COMMENT '联系人',
    `contact_phone` VARCHAR(20) DEFAULT NULL COMMENT '联系电话',
    `address` VARCHAR(255) DEFAULT NULL COMMENT '地址',
    `business_scope` VARCHAR(255) DEFAULT NULL COMMENT '经营范围',
    `bank_name` VARCHAR(100) DEFAULT NULL COMMENT '开户行',
    `bank_account` VARCHAR(50) DEFAULT NULL COMMENT '银行账号',
    `tax_no` VARCHAR(50) DEFAULT NULL COMMENT '税号',
    `total_purchase_amount` DECIMAL(14,2) DEFAULT 0.00 COMMENT '采购总金额',
    `total_purchase_orders` INT DEFAULT 0 COMMENT '采购订单总数',
    `last_purchase_at` DATETIME DEFAULT NULL COMMENT '最后采购时间',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_store_supplier_code` (`store_id`, `supplier_code`),
    KEY `idx_supplier_name` (`supplier_name`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商表';

-- ============================================================
-- 表26: supplier_products - 供应商商品关联表
-- ============================================================
DROP TABLE IF EXISTS `supplier_products`;
CREATE TABLE `supplier_products` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `supplier_id` BIGINT NOT NULL COMMENT '供应商ID',
    `sku_id` BIGINT NOT NULL COMMENT 'SKU ID',
    `supply_price` DECIMAL(12,2) DEFAULT 0.00 COMMENT '供货价',
    `min_order_quantity` INT DEFAULT 1 COMMENT '最小起订量',
    `lead_time` INT DEFAULT NULL COMMENT '交货周期(天)',
    `is_default` TINYINT DEFAULT 0 COMMENT '是否默认供应商(0-否,1-是)',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_supplier_sku` (`supplier_id`, `sku_id`),
    KEY `idx_sku_id` (`sku_id`),
    KEY `idx_is_default` (`is_default`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商商品关联表';

-- ============================================================
-- 表27: transfer_orders - 调拨单表
-- ============================================================
DROP TABLE IF EXISTS `transfer_orders`;
CREATE TABLE `transfer_orders` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `transfer_no` VARCHAR(32) DEFAULT NULL COMMENT '调拨单号',
    `from_warehouse_id` BIGINT DEFAULT NULL COMMENT '调出仓库ID',
    `to_warehouse_id` BIGINT DEFAULT NULL COMMENT '调入仓库ID',
    `total_quantity` INT DEFAULT 0 COMMENT '调拨总数量',
    `total_amount` DECIMAL(12,2) DEFAULT 0.00 COMMENT '调拨总金额',
    `status` TINYINT DEFAULT 0 COMMENT '状态(0-待审核,1-已审核,2-已出库,3-已入库,4-已取消)',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `created_by` BIGINT DEFAULT NULL COMMENT '创建人ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `audited_by` BIGINT DEFAULT NULL COMMENT '审核人ID',
    `audited_at` DATETIME DEFAULT NULL COMMENT '审核时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_transfer_no` (`transfer_no`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_from_warehouse_id` (`from_warehouse_id`),
    KEY `idx_to_warehouse_id` (`to_warehouse_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_by` (`created_by`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='调拨单表';

-- ============================================================
-- 表28: transfer_items - 调拨明细表
-- ============================================================
DROP TABLE IF EXISTS `transfer_items`;
CREATE TABLE `transfer_items` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `transfer_order_id` BIGINT NOT NULL COMMENT '调拨单ID',
    `sku_id` BIGINT DEFAULT NULL COMMENT 'SKU ID',
    `quantity` INT DEFAULT 0 COMMENT '数量',
    `unit_cost` DECIMAL(12,2) DEFAULT 0.00 COMMENT '单位成本',
    `subtotal` DECIMAL(12,2) DEFAULT 0.00 COMMENT '小计',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_transfer_order_id` (`transfer_order_id`),
    KEY `idx_sku_id` (`sku_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='调拨明细表';

-- ============================================================
-- 表29: warehouses - 仓库表
-- ============================================================
DROP TABLE IF EXISTS `warehouses`;
CREATE TABLE `warehouses` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `warehouse_code` VARCHAR(32) DEFAULT NULL COMMENT '仓库编码',
    `warehouse_name` VARCHAR(100) DEFAULT NULL COMMENT '仓库名称',
    `warehouse_type` TINYINT DEFAULT 1 COMMENT '仓库类型(1-主仓库,2-分仓库,3-展厅)',
    `address` VARCHAR(255) DEFAULT NULL COMMENT '地址',
    `contact_person` VARCHAR(50) DEFAULT NULL COMMENT '联系人',
    `contact_phone` VARCHAR(20) DEFAULT NULL COMMENT '联系电话',
    `manager_id` BIGINT DEFAULT NULL COMMENT '管理员ID',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_warehouse_code` (`warehouse_code`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_warehouse_type` (`warehouse_type`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='仓库表';

-- ============================================================
-- 表30: warehouse_stocks - 仓库商品库存表
-- ============================================================
DROP TABLE IF EXISTS `warehouse_stocks`;
CREATE TABLE `warehouse_stocks` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `warehouse_id` BIGINT NOT NULL COMMENT '仓库ID',
    `sku_id` BIGINT NOT NULL COMMENT 'SKU ID',
    `stock_quantity` INT DEFAULT 0 COMMENT '库存数量',
    `available_quantity` INT DEFAULT 0 COMMENT '可用数量',
    `locked_quantity` INT DEFAULT 0 COMMENT '锁定数量',
    `warning_stock` INT DEFAULT 10 COMMENT '预警库存',
    `version` INT DEFAULT 0 COMMENT '乐观锁版本号',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_warehouse_sku` (`warehouse_id`, `sku_id`),
    KEY `idx_sku_id` (`sku_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='仓库商品库存表';

-- ============================================================
-- 表31: warehouse_gift_stocks - 仓库礼品库存表
-- ============================================================
DROP TABLE IF EXISTS `warehouse_gift_stocks`;
CREATE TABLE `warehouse_gift_stocks` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `warehouse_id` BIGINT NOT NULL COMMENT '仓库ID',
    `gift_id` BIGINT NOT NULL COMMENT '礼品ID',
    `stock_quantity` INT DEFAULT 0 COMMENT '库存数量',
    `available_quantity` INT DEFAULT 0 COMMENT '可用数量',
    `locked_quantity` INT DEFAULT 0 COMMENT '锁定数量',
    `warning_stock` INT DEFAULT 10 COMMENT '预警库存',
    `version` INT DEFAULT 0 COMMENT '乐观锁版本号',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_warehouse_gift` (`warehouse_id`, `gift_id`),
    KEY `idx_gift_id` (`gift_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='仓库礼品库存表';

-- ============================================================
-- 表32: roles - 角色表
-- ============================================================
DROP TABLE IF EXISTS `roles`;
CREATE TABLE `roles` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `role_code` VARCHAR(50) DEFAULT NULL COMMENT '角色编码',
    `role_name` VARCHAR(50) DEFAULT NULL COMMENT '角色名称',
    `role_type` TINYINT DEFAULT 1 COMMENT '角色类型(1-系统预设,2-自定义)',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '描述',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_code` (`role_code`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

-- ============================================================
-- 表33: permissions - 权限表
-- ============================================================
DROP TABLE IF EXISTS `permissions`;
CREATE TABLE `permissions` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `permission_code` VARCHAR(50) DEFAULT NULL COMMENT '权限编码',
    `permission_name` VARCHAR(50) DEFAULT NULL COMMENT '权限名称',
    `permission_type` TINYINT DEFAULT NULL COMMENT '权限类型(1-菜单,2-按钮,3-接口)',
    `parent_id` BIGINT DEFAULT 0 COMMENT '父级ID',
    `path` VARCHAR(100) DEFAULT NULL COMMENT '路由路径',
    `icon` VARCHAR(50) DEFAULT NULL COMMENT '图标',
    `sort_order` INT DEFAULT 0 COMMENT '排序',
    `status` TINYINT DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_permission_code` (`permission_code`),
    KEY `idx_parent_id` (`parent_id`),
    KEY `idx_permission_type` (`permission_type`),
    KEY `idx_status` (`status`),
    KEY `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';

-- ============================================================
-- 表34: role_permissions - 角色权限关联表
-- ============================================================
DROP TABLE IF EXISTS `role_permissions`;
CREATE TABLE `role_permissions` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `role_id` BIGINT NOT NULL COMMENT '角色ID',
    `permission_id` BIGINT NOT NULL COMMENT '权限ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_permission` (`role_id`, `permission_id`),
    KEY `idx_permission_id` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';

-- ============================================================
-- 表35: user_roles - 用户角色关联表
-- ============================================================
DROP TABLE IF EXISTS `user_roles`;
CREATE TABLE `user_roles` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT NOT NULL COMMENT '用户ID',
    `role_id` BIGINT NOT NULL COMMENT '角色ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

-- ============================================================
-- 表36: system_configs - 系统配置表
-- ============================================================
DROP TABLE IF EXISTS `system_configs`;
CREATE TABLE `system_configs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `config_key` VARCHAR(100) DEFAULT NULL COMMENT '配置键',
    `config_value` TEXT DEFAULT NULL COMMENT '配置值',
    `config_type` VARCHAR(50) DEFAULT NULL COMMENT '配置类型',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';

-- ============================================================
-- 表37: operation_logs - 操作日志表
-- ============================================================
DROP TABLE IF EXISTS `operation_logs`;
CREATE TABLE `operation_logs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT DEFAULT NULL COMMENT '操作用户ID',
    `username` VARCHAR(50) DEFAULT NULL COMMENT '操作用户名',
    `action` VARCHAR(100) DEFAULT NULL COMMENT '操作动作',
    `biz_type` VARCHAR(50) DEFAULT NULL COMMENT '业务类型',
    `biz_id` BIGINT DEFAULT NULL COMMENT '业务ID',
    `detail` TEXT DEFAULT NULL COMMENT '操作详情',
    `ip_address` VARCHAR(50) DEFAULT NULL COMMENT 'IP地址',
    `user_agent` VARCHAR(255) DEFAULT NULL COMMENT '用户代理',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_action` (`action`),
    KEY `idx_biz_type` (`biz_type`),
    KEY `idx_biz_id` (`biz_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志表';

-- ============================================================
-- 表38: stock_alerts - 库存预警记录表
-- ============================================================
DROP TABLE IF EXISTS `stock_alerts`;
CREATE TABLE `stock_alerts` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `store_id` BIGINT DEFAULT 1 COMMENT '所属门店ID',
    `warehouse_id` BIGINT DEFAULT NULL COMMENT '仓库ID',
    `alert_type` TINYINT DEFAULT NULL COMMENT '预警类型(1-商品库存不足,2-礼品库存不足,3-库存积压)',
    `sku_id` BIGINT DEFAULT NULL COMMENT 'SKU ID',
    `gift_id` BIGINT DEFAULT NULL COMMENT '礼品ID',
    `current_stock` INT DEFAULT 0 COMMENT '当前库存',
    `warning_stock` INT DEFAULT 0 COMMENT '预警库存',
    `alert_status` TINYINT DEFAULT 0 COMMENT '处理状态(0-未处理,1-已处理,2-已忽略)',
    `handled_by` BIGINT DEFAULT NULL COMMENT '处理人ID',
    `handled_at` DATETIME DEFAULT NULL COMMENT '处理时间',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_store_id` (`store_id`),
    KEY `idx_warehouse_id` (`warehouse_id`),
    KEY `idx_alert_type` (`alert_type`),
    KEY `idx_sku_id` (`sku_id`),
    KEY `idx_gift_id` (`gift_id`),
    KEY `idx_alert_status` (`alert_status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存预警记录表';
