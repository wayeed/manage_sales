# 数据库迁移规范

## 目录结构

```
sql/
├── init.sql              # 完整初始化脚本（基准，全新环境部署使用）
├── seed.sql              # 初始数据（角色、权限、配置）
├── migrations/           # 增量迁移脚本（已有环境升级使用）
│   ├── 001_add_delivery_tables.sql
│   ├── 002_add_order_stock_fields.sql
│   ├── 003_add_customer_fields.sql
│   ├── 004_add_stock_queue.sql
│   ├── 005_add_warehouse_stock_fields.sql
│   ├── 006_add_stocktake_tables.sql
│   └── 007_add_misc_fields.sql
└── README.md             # 本文件
```

## 同步规范

### 新增表时必须修改的文件

1. `backend/internal/models/xxx.go` - GORM 模型定义
2. `sql/init.sql` - 完整初始化脚本中添加 CREATE TABLE
3. `sql/migrations/xxx_xxx.sql` - 增量迁移脚本
4. `backend/internal/pkg/database/mysql.go` - AutoMigrate 列表中添加模型

### 新增字段时必须修改的文件

1. `backend/internal/models/xxx.go` - GORM 模型添加字段
2. `sql/init.sql` - 完整初始化脚本中添加 ALTER TABLE 或在 CREATE TABLE 中添加列
3. `sql/migrations/xxx_xxx.sql` - 增量迁移脚本（ALTER TABLE ADD COLUMN）
4. `backend/internal/pkg/database/mysql.go` - migrateColumns() 中添加 HasColumn 检查

### 迁移脚本命名规范

- 格式：`{三位序号}_{英文描述}.sql`
- 序号从 001 开始，顺序递增，不跳号
- 描述使用英文小写，下划线分隔
- 示例：`008_add_order_source_type.sql`

### 迁移脚本内容规范

```sql
-- Migration: 008_add_order_source_type
-- Description: 订单表添加客户来源字段
-- CreatedAt: 2025-01-25

ALTER TABLE `orders` ADD COLUMN `source_type` TINYINT DEFAULT 0 COMMENT '客户来源' AFTER `source`;

-- 回滚SQL:
-- ALTER TABLE `orders` DROP COLUMN `source_type`;
```

每个迁移脚本必须包含：
- 头部注释：Migration编号、Description、CreatedAt
- 正向SQL
- 回滚SQL（注释形式）

## 提交检查清单

新增表或字段时，提交前确认：

- [ ] GORM 模型已定义
- [ ] init.sql 已同步更新
- [ ] migrations/ 下有对应的增量脚本
- [ ] mysql.go 的 AutoMigrate 或 migrateColumns 已添加
- [ ] 迁移脚本编号连续不跳号
- [ ] 回滚SQL已注释保留

## 部署方式

### 全新环境

```bash
mysql -u root -p < sql/init.sql
mysql -u root -p < sql/seed.sql
```

### 已有环境升级

按序号顺序执行增量脚本：

```bash
for f in sql/migrations/*.sql; do
  echo "Executing $f..."
  mysql -u root -p < "$f"
done
```

### 开发环境

服务启动时 GORM AutoMigrate + migrateColumns 会自动处理，无需手动执行 SQL。
