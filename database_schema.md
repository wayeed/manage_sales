# 家具销售提成管理系统 - 数据库表结构设计文档

---

## 文档版本控制

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|----------|
| v1.0 | 2026-05-07 | 产品组 | 初始版本，基于实际项目结构 |

---

## 1. 数据库设计概述

### 1.1 设计原则

| 原则 | 说明 |
|------|------|
| 规范化设计 | 遵循第三范式（3NF），减少数据冗余 |
| 性能优先 | 合理创建索引，优化查询性能 |
| 扩展性 | 预留多门店、多仓库扩展能力 |
| 数据完整性 | 通过外键约束、唯一约束保证数据一致性 |
| 乐观锁 | 库存操作使用版本号防止并发冲突 |

### 1.2 数据库ER图

```
                                      ┌──────────────────┐
                                      │    stores        │
                                      └────────┬─────────┘
                                               │
          ┌────────────────────────────────────┼────────────────────────────────────┐
          │                                    │                                    │
          ▼                                    ▼                                    ▼
┌─────────────────┐                  ┌─────────────────┐                  ┌─────────────────┐
│     users       │                  │   warehouses    │                  │   suppliers     │
│ (用户表)        │                  │   (仓库表)       │                  │   (供应商表)     │
└────────┬────────┘                  └────────┬────────┘                  └────────┬────────┘
         │                                    │                                    │
         ├────────────────────────────────────┼────────────────────────────────────┤
         │                                    │                                    │
         ▼                                    ▼                                    ▼
┌─────────────────┐          ┌───────────────────────┐          ┌─────────────────────┐
│   customers     │          │    warehouse_stocks   │          │   purchase_orders   │
│   (客户表)       │          │   (仓库商品库存表)     │          │    (采购订单表)      │
└────────┬────────┘          └─────────────┬─────────┘          └───────────┬─────────┘
         │                                 │                                  │
         │                                 ▼                                  ▼
         │                      ┌───────────────────────┐          ┌─────────────────────┐
         │                      │   inventory_batches   │          │   purchase_items    │
         │                      │   (库存批次表)         │          │    (采购订单项表)    │
         │                      └───────────────────────┘          └─────────────────────┘
         │
         ▼
┌─────────────────┐          ┌───────────────────────┐          ┌─────────────────────┐
│    orders       │◄─────────│    order_items        │          │    order_gifts      │
│   (订单表)       │          │   (订单项表)           │          │   (订单礼品表)       │
└────────┬────────┘          └───────────────────────┘          └─────────────────────┘
         │
         ├─────────────────────────────┐
         │                             │
         ▼                             ▼
┌─────────────────┐          ┌─────────────────────┐
│   payments      │          │   commissions       │
│   (回款记录表)    │          │    (提成记录表)      │
└─────────────────┘          └─────────────────────┘
```

### 1.3 表结构总览

| 序号 | 表名 | 中文名称 | 核心用途 |
|:---:|------|----------|----------|
| 1 | users | 用户表 | 存储员工账号信息 |
| 2 | roles | 角色表 | 定义系统角色 |
| 3 | permissions | 权限表 | 定义系统权限 |
| 4 | user_roles | 用户角色关联表 | 用户与角色的多对多关系 |
| 5 | role_permissions | 角色权限关联表 | 角色与权限的多对多关系 |
| 6 | stores | 门店表 | 存储门店信息 |
| 7 | products | 商品表 | 商品基础信息 |
| 8 | product_skus | 商品SKU表 | 商品规格信息 |
| 9 | categories | 品类表 | 商品分类 |
| 10 | orders | 订单表 | 订单主表 |
| 11 | order_items | 订单项表 | 订单商品明细 |
| 12 | order_gifts | 订单礼品表 | 订单礼品明细 |
| 13 | payments | 回款记录表 | 订单回款记录 |
| 14 | commissions | 提成记录表 | 提成计算记录 |
| 15 | customers | 客户表 | 客户信息 |
| 16 | suppliers | 供应商表 | 供应商信息 |
| 17 | warehouses | 仓库表 | 仓库信息 |
| 18 | warehouse_stocks | 仓库商品库存表 | 商品库存记录 |
| 19 | warehouse_gift_stocks | 仓库礼品库存表 | 礼品库存记录 |
| 20 | gifts | 礼品表 | 礼品基础信息 |
| 21 | inventory_batches | 库存批次表 | 商品入库批次 |
| 22 | gift_inventory_batches | 礼品库存批次表 | 礼品入库批次 |
| 23 | inventory_transactions | 库存事务表 | 库存变动记录 |
| 24 | purchase_orders | 采购订单表 | 采购订单主表 |
| 25 | purchase_items | 采购订单项表 | 采购商品明细 |
| 26 | transfer_orders | 调拨订单表 | 仓库调拨订单 |
| 27 | transfer_items | 调拨订单项表 | 调拨商品明细 |
| 28 | fund_pools | 基金池表 | 团队奖励基金池 |
| 29 | fund_pool_shares | 基金池分红表 | 基金池分红记录 |
| 30 | salary_records | 工资记录表 | 员工工资记录 |
| 31 | salary_items | 工资项表 | 工资组成明细 |
| 32 | referral_relations | 引荐关系表 | 老带新关系记录 |
| 33 | peers | 同行表 | 同行人员信息 |
| 34 | stock_alerts | 库存预警表 | 库存预警记录 |
| 35 | customer_follow_ups | 客户跟进表 | 客户跟进记录 |
| 36 | system_configs | 系统配置表 | 系统参数配置 |

---

## 2. 表结构详细设计

### 2.1 用户管理模块

#### 2.1.1 users（用户表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | NULL | 门店ID |
| employee_no | VARCHAR(32) | UNIQUE | 员工编号 |
| username | VARCHAR(64) | UNIQUE, NOT NULL | 登录用户名 |
| password | VARCHAR(128) | NOT NULL | 密码（加密存储） |
| real_name | VARCHAR(64) | | 真实姓名 |
| phone | VARCHAR(20) | UNIQUE | 手机号 |
| email | VARCHAR(128) | | 邮箱 |
| department_id | BIGINT | NULL | 部门ID |
| role | INT | DEFAULT 0, NOT NULL | 角色类型：0-业务员,1-主管,2-店长,3-仓管,4-老板,5-管理员 |
| status | TINYINT | DEFAULT 1, NOT NULL | 状态：0-禁用,1-启用 |
| entry_date | DATETIME | NULL | 入职日期 |
| probation_end_date | DATETIME | NULL | 试用期结束日期 |
| is_formal | TINYINT | DEFAULT 0 | 是否正式员工：0-试用,1-正式 |
| parent_id | BIGINT | NULL | 上级ID（主管/店长） |
| referrer_id | BIGINT | NULL | 引荐人ID（老带新） |
| base_salary | DECIMAL(12,2) | DEFAULT 0 | 基本工资 |
| id_card | VARCHAR(18) | | 身份证号 |
| bank_account | VARCHAR(32) | | 银行账号 |
| bank_name | VARCHAR(64) | | 开户行 |
| avatar | VARCHAR(255) | | 头像URL |
| last_login_at | DATETIME | NULL | 最后登录时间 |
| last_login_ip | VARCHAR(45) | | 最后登录IP |
| created_by | BIGINT | NULL | 创建人ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `idx_store_id` (store_id)
- `idx_role` (role)
- `idx_parent_id` (parent_id)

---

#### 2.1.2 roles（角色表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| name | VARCHAR(64) | NOT NULL | 角色名称 |
| code | VARCHAR(32) | UNIQUE, NOT NULL | 角色编码 |
| description | VARCHAR(255) | | 角色描述 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

#### 2.1.3 permissions（权限表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| name | VARCHAR(64) | NOT NULL | 权限名称 |
| code | VARCHAR(64) | UNIQUE, NOT NULL | 权限编码 |
| module | VARCHAR(64) | | 所属模块 |
| parent_id | BIGINT | NULL | 父权限ID |
| sort | INT | DEFAULT 0 | 排序号 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

#### 2.1.4 user_roles（用户角色关联表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| user_id | BIGINT | NOT NULL | 用户ID |
| role_id | BIGINT | NOT NULL | 角色ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `uk_user_role` (user_id, role_id) UNIQUE

---

#### 2.1.5 role_permissions（角色权限关联表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| role_id | BIGINT | NOT NULL | 角色ID |
| permission_id | BIGINT | NOT NULL | 权限ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `uk_role_permission` (role_id, permission_id) UNIQUE

---

### 2.2 商品管理模块

#### 2.2.1 categories（品类表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | DEFAULT 1 | 门店ID |
| name | VARCHAR(64) | NOT NULL | 品类名称 |
| code | VARCHAR(32) | UNIQUE | 品类编码 |
| parent_id | BIGINT | NULL | 父品类ID |
| sort | INT | DEFAULT 0 | 排序号 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

#### 2.2.2 products（商品表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | DEFAULT 1 | 门店ID |
| category_id | BIGINT | NULL | 品类ID |
| product_code | VARCHAR(32) | UNIQUE | 商品编码 |
| product_name | VARCHAR(100) | NOT NULL | 商品名称 |
| brand | VARCHAR(50) | | 品牌 |
| product_image | VARCHAR(255) | | 商品图片URL |
| description | TEXT | | 商品描述 |
| list_price | DECIMAL(12,2) | DEFAULT 0.00 | 挂牌售价 |
| min_price | DECIMAL(12,2) | DEFAULT 0.00 | 最低销售价 |
| reference_cost | DECIMAL(12,2) | DEFAULT 0.00 | 参考成本价 |
| cost_price | DECIMAL(12,2) | DEFAULT 0.00 | 当前成本价 |
| total_cost_rate | DECIMAL(5,4) | DEFAULT 1.2000 | 综合成本率 |
| warning_stock | INT | DEFAULT 10 | 预警库存 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_by | BIGINT | NULL | 创建人ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `idx_category_id` (category_id)

---

#### 2.2.3 product_skus（商品SKU表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| product_id | BIGINT | NOT NULL | 商品ID |
| sku_code | VARCHAR(32) | UNIQUE | SKU编码 |
| spec_name | VARCHAR(100) | | 规格名称（如颜色、尺寸） |
| spec_value | VARCHAR(200) | | 规格值 |
| list_price | DECIMAL(12,2) | DEFAULT 0.00 | 挂牌售价 |
| cost_price | DECIMAL(12,2) | DEFAULT 0.00 | 成本价 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `idx_product_id` (product_id)

---

### 2.3 订单管理模块

#### 2.3.1 orders（订单表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | DEFAULT 1 | 门店ID |
| order_no | VARCHAR(32) | UNIQUE | 订单号 |
| salesman_id | BIGINT | NOT NULL | 业务员ID |
| customer_id | BIGINT | NULL | 客户ID |
| customer_name | VARCHAR(100) | | 客户姓名 |
| customer_phone | VARCHAR(20) | | 客户电话 |
| customer_address | VARCHAR(500) | | 客户地址 |
| source | TINYINT | DEFAULT 0 | 订单来源：0-手机端,1-PC端 |
| delivery_status | TINYINT | DEFAULT 0 | 配送状态：0-未发货,1-已发货,2-已签收 |
| order_type | TINYINT | DEFAULT 1 | 订单类型：1-单品,2-多品,3-特殊审批,4-同行带单 |
| order_status | TINYINT | DEFAULT 0 | 订单状态：0-待审核,1-已生效,2-已驳回,3-已取消,4-已完成 |
| payment_status | TINYINT | DEFAULT 0 | 付款状态：0-未付款,1-部分付款,2-已付清 |
| total_list_price | DECIMAL(12,2) | DEFAULT 0.00 | 总挂牌价 |
| total_sale_price | DECIMAL(12,2) | DEFAULT 0.00 | 总成交价 |
| discount_amount | DECIMAL(12,2) | DEFAULT 0.00 | 优惠金额 |
| final_amount | DECIMAL(12,2) | DEFAULT 0.00 | 最终金额 |
| total_cost | DECIMAL(12,2) | DEFAULT 0.00 | 总成本 |
| gift_cost | DECIMAL(12,2) | DEFAULT 0.00 | 礼品成本 |
| actual_profit | DECIMAL(12,2) | DEFAULT 0.00 | 实际利润 |
| category_count | INT | DEFAULT 0 | 品类数量 |
| sku_count | INT | DEFAULT 0 | SKU数量 |
| total_quantity | INT | DEFAULT 0 | 商品总数 |
| paid_amount | DECIMAL(12,2) | DEFAULT 0.00 | 已回款金额 |
| remaining_amount | DECIMAL(12,2) | DEFAULT 0.00 | 剩余金额 |
| is_peer_order | TINYINT | DEFAULT 0 | 是否同行带单：0-否,1-是 |
| peer_id | BIGINT | NULL | 同行ID |
| is_special_approved | TINYINT | DEFAULT 0 | 是否特殊审批：0-否,1-是 |
| approval_remark | VARCHAR(500) | | 审批备注 |
| approved_by | BIGINT | NULL | 审批人ID |
| approved_at | DATETIME | NULL | 审批时间 |
| is_returned | TINYINT | DEFAULT 0 | 是否退货：0-否,1-是 |
| return_amount | DECIMAL(12,2) | DEFAULT 0.00 | 退货金额 |
| return_profit | DECIMAL(12,2) | DEFAULT 0.00 | 退货利润调整 |
| remark | VARCHAR(500) | | 订单备注 |
| order_date | DATETIME | NULL | 下单日期 |
| created_by | BIGINT | NULL | 创建人ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `idx_salesman_id` (salesman_id)
- `idx_customer_id` (customer_id)
- `idx_order_status` (order_status)
- `idx_order_date` (order_date)

---

#### 2.3.2 order_items（订单项表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| order_id | BIGINT | NOT NULL | 订单ID |
| sku_id | BIGINT | NOT NULL | SKU ID |
| product_id | BIGINT | NOT NULL | 商品ID |
| product_name | VARCHAR(100) | NOT NULL | 商品名称 |
| sku_code | VARCHAR(32) | | SKU编码 |
| spec_info | VARCHAR(200) | | 规格信息 |
| quantity | INT | NOT NULL | 数量 |
| list_price | DECIMAL(12,2) | NOT NULL | 挂牌售价 |
| sale_price | DECIMAL(12,2) | NOT NULL | 成交单价 |
| cost_price | DECIMAL(12,2) | DEFAULT 0.00 | 成本价 |
| amount | DECIMAL(12,2) | NOT NULL | 小计金额 |
| profit | DECIMAL(12,2) | DEFAULT 0.00 | 单项利润 |
| remark | VARCHAR(200) | | 备注 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `idx_order_id` (order_id)
- `idx_sku_id` (sku_id)

---

#### 2.3.3 order_gifts（订单礼品表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| order_id | BIGINT | NOT NULL | 订单ID |
| gift_id | BIGINT | NOT NULL | 礼品ID |
| gift_name | VARCHAR(100) | NOT NULL | 礼品名称 |
| quantity | INT | NOT NULL | 数量 |
| cost_price | DECIMAL(12,2) | DEFAULT 0.00 | 成本价 |
| amount | DECIMAL(12,2) | DEFAULT 0.00 | 总成本 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `idx_order_id` (order_id)
- `idx_gift_id` (gift_id)

---

#### 2.3.4 payments（回款记录表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| order_id | BIGINT | NOT NULL | 订单ID |
| amount | DECIMAL(12,2) | NOT NULL | 回款金额 |
| payment_method | TINYINT | DEFAULT 0 | 回款方式：0-现金,1-银行卡,2-转账,3-支付宝,4-微信 |
| payment_no | VARCHAR(64) | | 回款单号 |
| status | TINYINT | DEFAULT 0 | 状态：0-待审核,1-已审核,2-已驳回 |
| approved_by | BIGINT | NULL | 审核人ID |
| approved_at | DATETIME | NULL | 审核时间 |
| remark | VARCHAR(200) | | 备注 |
| created_by | BIGINT | NULL | 创建人ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `idx_order_id` (order_id)
- `idx_status` (status)

---

### 2.4 库存管理模块

#### 2.4.1 warehouses（仓库表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | DEFAULT 1 | 门店ID |
| name | VARCHAR(64) | NOT NULL | 仓库名称 |
| code | VARCHAR(32) | UNIQUE | 仓库编码 |
| address | VARCHAR(500) | | 仓库地址 |
| manager_id | BIGINT | NULL | 仓库管理员ID |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

#### 2.4.2 warehouse_stocks（仓库商品库存表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| warehouse_id | BIGINT | NOT NULL | 仓库ID |
| sku_id | BIGINT | NOT NULL | SKU ID |
| stock_quantity | INT | DEFAULT 0 | 当前库存数量 |
| available_quantity | INT | DEFAULT 0 | 可用库存数量 |
| locked_quantity | INT | DEFAULT 0 | 锁定库存数量（已下单未出库） |
| warning_stock | INT | DEFAULT 10 | 预警库存 |
| version | INT | DEFAULT 0 | 乐观锁版本号 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `uk_warehouse_sku` (warehouse_id, sku_id) UNIQUE
- `idx_sku_id` (sku_id)

---

#### 2.4.3 gifts（礼品表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | DEFAULT 1 | 门店ID |
| gift_code | VARCHAR(32) | UNIQUE | 礼品编码 |
| gift_name | VARCHAR(100) | NOT NULL | 礼品名称 |
| cost_price | DECIMAL(12,2) | DEFAULT 0.00 | 成本价 |
| description | VARCHAR(500) | | 礼品描述 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

#### 2.4.4 warehouse_gift_stocks（仓库礼品库存表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| warehouse_id | BIGINT | NOT NULL | 仓库ID |
| gift_id | BIGINT | NOT NULL | 礼品ID |
| stock_quantity | INT | DEFAULT 0 | 当前库存数量 |
| available_quantity | INT | DEFAULT 0 | 可用库存数量 |
| locked_quantity | INT | DEFAULT 0 | 锁定库存数量 |
| warning_stock | INT | DEFAULT 10 | 预警库存 |
| version | INT | DEFAULT 0 | 乐观锁版本号 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `uk_warehouse_gift` (warehouse_id, gift_id) UNIQUE
- `idx_gift_id` (gift_id)

---

#### 2.4.5 inventory_batches（库存批次表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| warehouse_id | BIGINT | NOT NULL | 仓库ID |
| sku_id | BIGINT | NOT NULL | SKU ID |
| batch_no | VARCHAR(32) | UNIQUE | 批次号 |
| quantity | INT | NOT NULL | 批次数量 |
| cost_price | DECIMAL(12,2) | NOT NULL | 入库成本价 |
| remaining_quantity | INT | NOT NULL | 剩余数量 |
| source_type | TINYINT | DEFAULT 1 | 来源类型：1-采购入库,2-调拨入库 |
| source_id | BIGINT | NULL | 来源单据ID |
| expiry_date | DATETIME | NULL | 有效期 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `idx_warehouse_sku` (warehouse_id, sku_id)
- `idx_batch_no` (batch_no)

---

#### 2.4.6 inventory_transactions（库存事务表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| warehouse_id | BIGINT | NOT NULL | 仓库ID |
| sku_id | BIGINT | NULL | SKU ID（商品事务） |
| gift_id | BIGINT | NULL | 礼品ID（礼品事务） |
| transaction_type | TINYINT | NOT NULL | 事务类型：1-采购入库,2-销售出库,3-调拨出库,4-调拨入库,5-盘点调整,6-退货入库 |
| quantity | INT | NOT NULL | 变动数量（正数入库，负数出库） |
| cost_price | DECIMAL(12,2) | DEFAULT 0.00 | 成本价 |
| amount | DECIMAL(12,2) | DEFAULT 0.00 | 变动金额 |
| source_type | TINYINT | DEFAULT 0 | 来源类型：1-采购单,2-订单,3-调拨单,4-盘点单 |
| source_id | BIGINT | NULL | 来源单据ID |
| remark | VARCHAR(500) | | 备注 |
| created_by | BIGINT | NULL | 创建人ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `idx_warehouse_id` (warehouse_id)
- `idx_source_type_id` (source_type, source_id)

---

### 2.5 提成管理模块

#### 2.5.1 commissions（提成记录表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | NOT NULL | 门店ID |
| order_id | BIGINT | NOT NULL | 订单ID |
| employee_id | BIGINT | NULL | 员工ID（业务员/主管/店长） |
| peer_id | BIGINT | NULL | 同行ID |
| commission_type | TINYINT | NOT NULL | 提成类型：1-业务员提成,2-同行分成,3-主管团队分润,4-店长团队分润,5-基金池奖励,6-老带新奖励 |
| period_value | VARCHAR(10) | NOT NULL | 核算周期（如202605） |
| base_amount | DECIMAL(12,2) | DEFAULT 0.00 | 计算基数 |
| rate | DECIMAL(6,4) | DEFAULT 0.00 | 提成比例 |
| amount | DECIMAL(12,2) | DEFAULT 0.00 | 提成金额 |
| status | TINYINT | DEFAULT 0 | 状态：0-待回款,1-可发放,2-已发放 |
| settled_at | DATETIME | NULL | 发放时间 |
| remark | VARCHAR(500) | | 备注 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `idx_order_id` (order_id)
- `idx_employee_id` (employee_id)
- `idx_period_type` (period_value, commission_type)
- `idx_status` (status)

---

#### 2.5.2 fund_pools（基金池表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | NOT NULL | 门店ID |
| name | VARCHAR(64) | NOT NULL | 基金池名称 |
| code | VARCHAR(32) | UNIQUE | 基金池编码 |
| balance | DECIMAL(12,2) | DEFAULT 0.00 | 基金池余额 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

#### 2.5.3 fund_pool_shares（基金池分红表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| fund_pool_id | BIGINT | NOT NULL | 基金池ID |
| period_value | VARCHAR(10) | NOT NULL | 分红周期 |
| employee_id | BIGINT | NOT NULL | 员工ID |
| amount | DECIMAL(12,2) | DEFAULT 0.00 | 分红金额 |
| status | TINYINT | DEFAULT 0 | 状态：0-待发放,1-已发放 |
| settled_at | DATETIME | NULL | 发放时间 |
| remark | VARCHAR(200) | | 备注 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `idx_fund_pool_id` (fund_pool_id)
- `idx_employee_id` (employee_id)

---

#### 2.5.4 referral_relations（引荐关系表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| referrer_id | BIGINT | NOT NULL | 引荐人ID |
| referee_id | BIGINT | NOT NULL | 被引荐人ID |
| status | TINYINT | DEFAULT 1 | 关系状态：0-已终止,1-有效 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `uk_referral` (referrer_id, referee_id) UNIQUE

---

### 2.6 工资管理模块

#### 2.6.1 salary_records（工资记录表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | NOT NULL | 门店ID |
| employee_id | BIGINT | NOT NULL | 员工ID |
| period_value | VARCHAR(10) | NOT NULL | 工资周期（如202605） |
| base_salary | DECIMAL(12,2) | DEFAULT 0.00 | 基本工资 |
| commission_amount | DECIMAL(12,2) | DEFAULT 0.00 | 提成金额 |
| bonus_amount | DECIMAL(12,2) | DEFAULT 0.00 | 奖金金额 |
| deduction_amount | DECIMAL(12,2) | DEFAULT 0.00 | 扣款金额 |
| net_amount | DECIMAL(12,2) | DEFAULT 0.00 | 实发工资 |
| status | TINYINT | DEFAULT 0 | 状态：0-待核算,1-已核算,2-已发放 |
| paid_at | DATETIME | NULL | 发放时间 |
| remark | VARCHAR(500) | | 备注 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `idx_employee_period` (employee_id, period_value) UNIQUE
- `idx_status` (status)

---

#### 2.6.2 salary_items（工资项表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| salary_record_id | BIGINT | NOT NULL | 工资记录ID |
| item_type | TINYINT | NOT NULL | 工资项类型：1-基本工资,2-提成,3-奖金,4-扣款,5-补贴 |
| item_name | VARCHAR(64) | NOT NULL | 工资项名称 |
| amount | DECIMAL(12,2) | NOT NULL | 金额（正数收入，负数扣款） |
| remark | VARCHAR(200) | | 备注 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `idx_salary_record_id` (salary_record_id)

---

### 2.7 采购管理模块

#### 2.7.1 purchase_orders（采购订单表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | DEFAULT 1 | 门店ID |
| supplier_id | BIGINT | NOT NULL | 供应商ID |
| purchase_no | VARCHAR(32) | UNIQUE | 采购单号 |
| status | TINYINT | DEFAULT 0 | 状态：0-待审核,1-已审核,2-已入库,3-已取消 |
| total_amount | DECIMAL(12,2) | DEFAULT 0.00 | 采购总金额 |
| remark | VARCHAR(500) | | 备注 |
| approved_by | BIGINT | NULL | 审核人ID |
| approved_at | DATETIME | NULL | 审核时间 |
| created_by | BIGINT | NULL | 创建人ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `idx_supplier_id` (supplier_id)
- `idx_status` (status)

---

#### 2.7.2 purchase_items（采购订单项表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| purchase_order_id | BIGINT | NOT NULL | 采购订单ID |
| sku_id | BIGINT | NOT NULL | SKU ID |
| product_name | VARCHAR(100) | NOT NULL | 商品名称 |
| sku_code | VARCHAR(32) | | SKU编码 |
| spec_info | VARCHAR(200) | | 规格信息 |
| quantity | INT | NOT NULL | 采购数量 |
| unit_price | DECIMAL(12,2) | NOT NULL | 单价 |
| amount | DECIMAL(12,2) | NOT NULL | 小计金额 |
| received_quantity | INT | DEFAULT 0 | 已入库数量 |
| remark | VARCHAR(200) | | 备注 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `idx_purchase_order_id` (purchase_order_id)

---

### 2.8 客户管理模块

#### 2.8.1 customers（客户表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | DEFAULT 1 | 门店ID |
| customer_code | VARCHAR(32) | UNIQUE | 客户编码 |
| customer_name | VARCHAR(100) | NOT NULL | 客户姓名 |
| phone | VARCHAR(20) | | 联系电话 |
| email | VARCHAR(128) | | 邮箱 |
| address | VARCHAR(500) | | 地址 |
| gender | TINYINT | DEFAULT 0 | 性别：0-未知,1-男,2-女 |
| birthday | DATETIME | NULL | 生日 |
| level | TINYINT | DEFAULT 0 | 客户等级：0-普通,1-银卡,2-金卡,3-钻石 |
| total_orders | INT | DEFAULT 0 | 订单总数 |
| total_amount | DECIMAL(12,2) | DEFAULT 0.00 | 累计消费金额 |
| total_profit | DECIMAL(12,2) | DEFAULT 0.00 | 累计贡献利润 |
| last_order_at | DATETIME | NULL | 最后下单时间 |
| remark | VARCHAR(500) | | 备注 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_by | BIGINT | NULL | 创建人ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

**索引设计**：
- `idx_phone` (phone)
- `idx_level` (level)

---

#### 2.8.2 customer_follow_ups（客户跟进表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| customer_id | BIGINT | NOT NULL | 客户ID |
| user_id | BIGINT | NOT NULL | 跟进人ID |
| follow_type | TINYINT | NOT NULL | 跟进类型：1-电话,2-微信,3-短信,4-拜访,5-其他 |
| content | TEXT | | 跟进内容 |
| next_follow_at | DATETIME | NULL | 下次跟进时间 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `idx_customer_id` (customer_id)
- `idx_user_id` (user_id)

---

### 2.9 供应商管理模块

#### 2.9.1 suppliers（供应商表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | DEFAULT 1 | 门店ID |
| supplier_code | VARCHAR(32) | UNIQUE | 供应商编码 |
| supplier_name | VARCHAR(100) | NOT NULL | 供应商名称 |
| contact_name | VARCHAR(64) | | 联系人 |
| phone | VARCHAR(20) | | 联系电话 |
| email | VARCHAR(128) | | 邮箱 |
| address | VARCHAR(500) | | 地址 |
| bank_account | VARCHAR(32) | | 银行账号 |
| bank_name | VARCHAR(64) | | 开户行 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| remark | VARCHAR(500) | | 备注 |
| created_by | BIGINT | NULL | 创建人ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

### 2.10 调拨管理模块

#### 2.10.1 transfer_orders（调拨订单表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | DEFAULT 1 | 门店ID |
| transfer_no | VARCHAR(32) | UNIQUE | 调拨单号 |
| from_warehouse_id | BIGINT | NOT NULL | 调出仓库ID |
| to_warehouse_id | BIGINT | NOT NULL | 调入仓库ID |
| status | TINYINT | DEFAULT 0 | 状态：0-待审核,1-已审核,2-已出库,3-已入库,4-已取消 |
| remark | VARCHAR(500) | | 备注 |
| approved_by | BIGINT | NULL | 审核人ID |
| approved_at | DATETIME | NULL | 审核时间 |
| created_by | BIGINT | NULL | 创建人ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

#### 2.10.2 transfer_items（调拨订单项表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| transfer_order_id | BIGINT | NOT NULL | 调拨订单ID |
| sku_id | BIGINT | NOT NULL | SKU ID |
| product_name | VARCHAR(100) | NOT NULL | 商品名称 |
| sku_code | VARCHAR(32) | | SKU编码 |
| spec_info | VARCHAR(200) | | 规格信息 |
| quantity | INT | NOT NULL | 调拨数量 |
| cost_price | DECIMAL(12,2) | DEFAULT 0.00 | 成本价 |
| shipped_quantity | INT | DEFAULT 0 | 已出库数量 |
| received_quantity | INT | DEFAULT 0 | 已入库数量 |
| remark | VARCHAR(200) | | 备注 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**索引设计**：
- `idx_transfer_order_id` (transfer_order_id)

---

### 2.11 其他模块

#### 2.11.1 peers（同行表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| store_id | BIGINT | DEFAULT 1 | 门店ID |
| peer_code | VARCHAR(32) | UNIQUE | 同行编码 |
| real_name | VARCHAR(64) | NOT NULL | 姓名 |
| phone | VARCHAR(20) | UNIQUE | 联系电话 |
| wechat | VARCHAR(64) | | 微信号 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| remark | VARCHAR(500) | | 备注 |
| created_by | BIGINT | NULL | 创建人ID |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

#### 2.11.2 stock_alerts（库存预警表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| warehouse_id | BIGINT | NOT NULL | 仓库ID |
| sku_id | BIGINT | NULL | SKU ID（商品预警） |
| gift_id | BIGINT | NULL | 礼品ID（礼品预警） |
| alert_type | TINYINT | NOT NULL | 预警类型：1-低于预警库存,2-库存为零 |
| current_quantity | INT | NOT NULL | 当前库存 |
| warning_stock | INT | NOT NULL | 预警阈值 |
| status | TINYINT | DEFAULT 0 | 状态：0-未处理,1-已处理 |
| handled_by | BIGINT | NULL | 处理人ID |
| handled_at | DATETIME | NULL | 处理时间 |
| remark | VARCHAR(200) | | 备注 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

#### 2.11.3 stores（门店表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| name | VARCHAR(64) | NOT NULL | 门店名称 |
| code | VARCHAR(32) | UNIQUE | 门店编码 |
| address | VARCHAR(500) | | 门店地址 |
| phone | VARCHAR(20) | | 联系电话 |
| manager_id | BIGINT | NULL | 店长ID |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

#### 2.11.4 system_configs（系统配置表）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| config_key | VARCHAR(64) | UNIQUE, NOT NULL | 配置键 |
| config_value | TEXT | | 配置值 |
| config_type | VARCHAR(32) | DEFAULT 'string' | 配置类型：string, int, decimal, json |
| description | VARCHAR(255) | | 配置说明 |
| sort | INT | DEFAULT 0 | 排序号 |
| status | TINYINT | DEFAULT 1 | 状态：0-禁用,1-启用 |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

---

## 3. 表关系图

### 3.1 用户与权限关系

```
users ──(1:N)──► user_roles ◄──(N:1)── roles ──(1:N)──► role_permissions ◄──(N:1)── permissions
```

### 3.2 订单与关联关系

```
orders ──(1:N)──► order_items
orders ──(1:N)──► order_gifts
orders ──(1:N)──► payments
orders ──(1:N)──► commissions
orders ──(1:N)──► order_items ──(N:1)──► product_skus ──(N:1)──► products
```

### 3.3 库存与关联关系

```
warehouses ──(1:N)──► warehouse_stocks ──(N:1)──► product_skus
warehouses ──(1:N)──► warehouse_gift_stocks ──(N:1)──► gifts
warehouse_stocks ──(1:N)──► inventory_batches
warehouses ──(1:N)──► inventory_transactions
```

### 3.4 提成与工资关系

```
commissions ──(N:1)──► orders
commissions ──(N:1)──► users
salary_records ──(1:N)──► salary_items
salary_records ──(N:1)──► users
fund_pools ──(1:N)──► fund_pool_shares ──(N:1)──► users
```

---

## 4. 数据字典

### 4.1 枚举值定义

#### 4.1.1 用户角色（user.role）

| 值 | 名称 | 说明 |
|----|------|------|
| 0 | 业务员 | 一线销售人员 |
| 1 | 主管 | 团队主管 |
| 2 | 店长 | 门店负责人 |
| 3 | 仓管 | 仓库管理员 |
| 4 | 老板 | 最高权限 |
| 5 | 管理员 | 系统管理员 |

#### 4.1.2 订单状态（orders.order_status）

| 值 | 名称 | 说明 |
|----|------|------|
| 0 | 待审核 | 订单创建后等待审核 |
| 1 | 已生效 | 审核通过，订单生效 |
| 2 | 已驳回 | 审核未通过 |
| 3 | 已取消 | 订单取消 |
| 4 | 已完成 | 订单完成（已发货、已回款） |

#### 4.1.3 订单类型（orders.order_type）

| 值 | 名称 | 说明 |
|----|------|------|
| 1 | 单品订单 | 单一品类商品订单 |
| 2 | 多品订单 | 多个品类商品订单 |
| 3 | 特殊审批订单 | 需要特殊审批的订单 |
| 4 | 同行带单 | 同行引荐的订单 |

#### 4.1.4 提成类型（commissions.commission_type）

| 值 | 名称 | 说明 |
|----|------|------|
| 1 | 业务员提成 | 业务员销售提成 |
| 2 | 同行分成 | 同行引荐分成 |
| 3 | 主管团队分润 | 主管团队奖励 |
| 4 | 店长团队分润 | 店长团队奖励 |
| 5 | 基金池奖励 | 基金池分配奖励 |
| 6 | 老带新奖励 | 引荐新员工奖励 |

#### 4.1.5 回款方式（payments.payment_method）

| 值 | 名称 | 说明 |
|----|------|------|
| 0 | 现金 | 现金支付 |
| 1 | 银行卡 | POS刷卡 |
| 2 | 转账 | 银行转账 |
| 3 | 支付宝 | 支付宝支付 |
| 4 | 微信 | 微信支付 |

---

## 附录：索引优化建议

| 表名 | 建议索引 | 用途 |
|------|----------|------|
| orders | (salesman_id, order_date) | 业务员业绩查询 |
| orders | (order_status, created_at) | 订单状态筛选 |
| commissions | (employee_id, period_value) | 员工提成查询 |
| commissions | (order_id, commission_type) | 订单提成明细 |
| warehouse_stocks | (warehouse_id, sku_id) | 库存查询（已有） |
| payments | (order_id, status) | 回款查询 |
| customers | (phone) | 客户搜索 |
| inventory_transactions | (warehouse_id, created_at) | 库存流水查询 |

---

**文档审批**

| 角色 | 姓名 | 审批日期 |
|------|------|----------|
| 产品负责人 | | |
| 技术负责人 | | |
| DBA | | |