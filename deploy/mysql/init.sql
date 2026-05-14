-- LCA 数据库初始化
CREATE DATABASE IF NOT EXISTS lca DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE lca;

-- 表由GORM AutoMigrate自动创建，此处只做基础设置
SET GLOBAL max_connections = 500;
SET GLOBAL innodb_buffer_pool_size = 268435456;
