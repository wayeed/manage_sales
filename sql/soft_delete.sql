-- ============================================================
-- 软删除：为业务表添加 deleted_at 字段
-- 执行前请备份数据库
-- ============================================================

ALTER TABLE `stores` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;
ALTER TABLE `products` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;
ALTER TABLE `product_skus` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;
ALTER TABLE `categories` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;
ALTER TABLE `peers` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;
ALTER TABLE `customers` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;
ALTER TABLE `warehouses` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;
ALTER TABLE `gifts` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;
ALTER TABLE `suppliers` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;
ALTER TABLE `roles` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;

-- 添加索引（GORM 软删除查询需要）
ALTER TABLE `stores` ADD INDEX `idx_stores_deleted_at` (`deleted_at`);
ALTER TABLE `products` ADD INDEX `idx_products_deleted_at` (`deleted_at`);
ALTER TABLE `product_skus` ADD INDEX `idx_product_skus_deleted_at` (`deleted_at`);
ALTER TABLE `categories` ADD INDEX `idx_categories_deleted_at` (`deleted_at`);
ALTER TABLE `peers` ADD INDEX `idx_peers_deleted_at` (`deleted_at`);
ALTER TABLE `customers` ADD INDEX `idx_customers_deleted_at` (`deleted_at`);
ALTER TABLE `warehouses` ADD INDEX `idx_warehouses_deleted_at` (`deleted_at`);
ALTER TABLE `gifts` ADD INDEX `idx_gifts_deleted_at` (`deleted_at`);
ALTER TABLE `suppliers` ADD INDEX `idx_suppliers_deleted_at` (`deleted_at`);
ALTER TABLE `roles` ADD INDEX `idx_roles_deleted_at` (`deleted_at`);
