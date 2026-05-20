-- LCA 日志收集智能分析系统 - 数据库初始化
-- 包含：完整建表 DDL + 基础参数 + 种子数据
-- 使用 CREATE TABLE IF NOT EXISTS，可重复执行（幂等）

CREATE DATABASE IF NOT EXISTS lca DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE lca;

SET GLOBAL max_connections = 500;
SET GLOBAL innodb_buffer_pool_size = 268435456;

-- =====================================================================
-- 建表 DDL（与 GORM AutoMigrate 保持一致）
-- =====================================================================

CREATE TABLE IF NOT EXISTS `agents` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `hostname` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ip` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `os_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'linux/windows',
  `version` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'offline' COMMENT 'online/offline',
  `last_heartbeat` bigint DEFAULT NULL,
  `labels` text COLLATE utf8mb4_unicode_ci COMMENT '标签JSON',
  PRIMARY KEY (`id`),
  KEY `idx_agents_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `alert_actions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `rule_id` bigint unsigned NOT NULL,
  `action_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'webhook/email/wecom/dingtalk',
  `config` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '动作配置JSON',
  `status` tinyint DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_alert_actions_deleted_at` (`deleted_at`),
  KEY `idx_alert_actions_rule_id` (`rule_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `alert_histories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `rule_id` bigint unsigned NOT NULL,
  `content` text COLLATE utf8mb4_unicode_ci,
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'firing' COMMENT 'firing/resolved',
  `rule_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `severity` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `store_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '触发告警的日志库名称',
  `raw_logs` longtext COLLATE utf8mb4_unicode_ci COMMENT '触发告警时的原始日志样本JSON',
  `root_cause` text COLLATE utf8mb4_unicode_ci COMMENT '根因分析',
  `resolution` text COLLATE utf8mb4_unicode_ci COMMENT '修复步骤(markdown)',
  `category` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '问题分类',
  PRIMARY KEY (`id`),
  KEY `idx_alert_histories_deleted_at` (`deleted_at`),
  KEY `idx_alert_histories_rule_id` (`rule_id`),
  KEY `idx_alert_histories_store_name` (`store_name`),
  KEY `idx_alert_histories_category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `alert_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `store_id` bigint unsigned NOT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `query_condition` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '查询条件DSL',
  `trigger_condition` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '触发条件JSON',
  `severity` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'warning' COMMENT 'critical/warning/info',
  `cron_expr` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行cron表达式',
  `silence_minutes` bigint DEFAULT '5',
  `status` tinyint DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_alert_rules_deleted_at` (`deleted_at`),
  KEY `idx_alert_rules_store_id` (`store_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `audit_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned NOT NULL,
  `username` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `action` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `resource` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `detail` text COLLATE utf8mb4_unicode_ci,
  `ip` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `user_agent` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `module` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_audit_logs_deleted_at` (`deleted_at`),
  KEY `idx_audit_logs_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `backup_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `slm_policy_id` bigint unsigned NOT NULL,
  `tenant_id` bigint unsigned NOT NULL,
  `snapshot_name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `indices` text COLLATE utf8mb4_unicode_ci COMMENT '备份的索引列表',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'in_progress' COMMENT 'in_progress/success/failed',
  `oss_path` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `size_bytes` bigint DEFAULT '0',
  `started_at` bigint DEFAULT NULL,
  `finished_at` bigint DEFAULT NULL,
  `error_msg` varchar(1000) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_backup_records_deleted_at` (`deleted_at`),
  KEY `idx_backup_records_slm_policy_id` (`slm_policy_id`),
  KEY `idx_backup_records_tenant_id` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `casbin_rule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ptype` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v0` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v1` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v2` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v3` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v4` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `v5` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `collect_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `agent_id` bigint unsigned NOT NULL,
  `store_id` bigint unsigned NOT NULL,
  `log_path_pattern` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '日志路径模式',
  `multiline_pattern` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '多行合并正则',
  `parse_mode` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'json' COMMENT 'json/regex/delimiter/raw',
  `parse_config` text COLLATE utf8mb4_unicode_ci COMMENT '解析配置JSON',
  `status` tinyint DEFAULT '1' COMMENT '1-启用 0-禁用',
  `store_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '日志库名称',
  PRIMARY KEY (`id`),
  KEY `idx_collect_tasks_deleted_at` (`deleted_at`),
  KEY `idx_collect_tasks_agent_id` (`agent_id`),
  KEY `idx_collect_tasks_store_id` (`store_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `departments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `parent_id` bigint unsigned DEFAULT '0',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sort` bigint DEFAULT '0',
  `leader` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_departments_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `file_offsets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `agent_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_id` bigint unsigned NOT NULL,
  `file_path` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL,
  `file_inode` bigint unsigned NOT NULL COMMENT '文件inode',
  `offset` bigint NOT NULL DEFAULT '0' COMMENT '采集偏移量',
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_file_offsets_agent_id` (`agent_id`),
  KEY `idx_file_offsets_task_id` (`task_id`),
  KEY `idx_file_offsets_file_path` (`file_path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ilm_policies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `tenant_id` bigint unsigned NOT NULL,
  `store_id` bigint unsigned NOT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `hot_days` bigint DEFAULT '7',
  `warm_days` bigint DEFAULT '30',
  `cold_days` bigint DEFAULT '90',
  `delete_days` bigint DEFAULT '180',
  `hot_config` text COLLATE utf8mb4_unicode_ci COMMENT 'hot阶段配置JSON',
  `warm_config` text COLLATE utf8mb4_unicode_ci COMMENT 'warm阶段配置JSON',
  `cold_config` text COLLATE utf8mb4_unicode_ci COMMENT 'cold阶段配置JSON',
  `status` tinyint DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_ilm_policies_deleted_at` (`deleted_at`),
  KEY `idx_ilm_policies_tenant_id` (`tenant_id`),
  KEY `idx_ilm_policies_store_id` (`store_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `licenses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `license_key` varchar(1000) COLLATE utf8mb4_unicode_ci NOT NULL,
  `machine_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `license_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'monthly/yearly/permanent',
  `expires_at` datetime(3) DEFAULT NULL,
  `bound_at` datetime(3) DEFAULT NULL,
  `status` tinyint DEFAULT '0' COMMENT '0-未激活 1-已激活 2-已过期',
  PRIMARY KEY (`id`),
  KEY `idx_licenses_deleted_at` (`deleted_at`),
  KEY `idx_licenses_machine_id` (`machine_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `log_store_fields` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `store_id` bigint unsigned NOT NULL,
  `field_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `field_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'keyword/text/long/double/date/ip',
  `is_indexed` tinyint(1) DEFAULT '1',
  `is_keyword` tinyint(1) DEFAULT '0',
  `analyzer` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分词器',
  `sort` bigint DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_log_store_fields_deleted_at` (`deleted_at`),
  KEY `idx_log_store_fields_store_id` (`store_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `log_stores` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `index_pattern` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ES索引模式',
  `api_key` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '日志推送API Key',
  `kafka_topic` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Kafka topic名称',
  `status` tinyint DEFAULT '1',
  `compress` tinyint(1) DEFAULT '1' COMMENT '是否压缩存储',
  `roll_max_days` bigint DEFAULT '0' COMMENT '滚动最大天数(0不限)',
  `roll_max_size_gb` bigint DEFAULT '0' COMMENT '滚动最大容量GB(0不限)',
  `cold_days` bigint DEFAULT '0' COMMENT '冷存储天数(0不限)',
  `delete_days` bigint DEFAULT '90' COMMENT '删除数据天数',
  `oss_repository` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'OSS仓库名称',
  `oss_endpoint` varchar(300) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'OSS Endpoint',
  `oss_bucket` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'OSS Bucket名称',
  `oss_access_key_id` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'OSS AccessKeyID',
  `oss_access_key_secret` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'OSS AccessKeySecret',
  `oss_path` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'OSS存储路径',
  `oss_chunk_size` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT '500mb' COMMENT '分块大小',
  `ai_alert_enabled` tinyint(1) DEFAULT '0' COMMENT '是否启用AI智能告警',
  `ai_alert_config` text COLLATE utf8mb4_unicode_ci COMMENT 'AI告警配置JSON',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_log_stores_name` (`name`),
  KEY `idx_log_stores_deleted_at` (`deleted_at`),
  KEY `idx_log_stores_api_key` (`api_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `login_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `username` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ip` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `location` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `browser` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `os` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint DEFAULT '1' COMMENT '1-成功 0-失败',
  `msg` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `module` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '后台',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_login_logs_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `parent_id` bigint unsigned DEFAULT '0',
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `path` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `component` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `icon` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sort` bigint DEFAULT '0',
  `status` tinyint DEFAULT '1',
  `menu_type` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT 'M' COMMENT 'M-目录 C-菜单 F-按钮',
  `visible` tinyint DEFAULT '1' COMMENT '1-显示 0-隐藏',
  `perms` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '权限标识',
  `api_path` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'API接口路径',
  PRIMARY KEY (`id`),
  KEY `idx_menus_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'menu/button/api',
  `parent_id` bigint unsigned DEFAULT '0',
  `path` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'API路径或前端路由',
  `method` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'HTTP方法',
  `sort` bigint DEFAULT '0',
  `icon` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint DEFAULT '1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_permissions_code` (`code`),
  KEY `idx_permissions_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `posts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sort` bigint DEFAULT '0',
  `status` tinyint DEFAULT '1',
  `remark` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_posts_code` (`code`),
  KEY `idx_posts_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `role_menus` (
  `role_id` bigint unsigned NOT NULL,
  `menu_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`role_id`,`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `role_permissions` (
  `role_id` bigint unsigned NOT NULL,
  `permission_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`role_id`,`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint DEFAULT '1',
  `sort` bigint DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_roles_code` (`code`),
  KEY `idx_roles_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `slm_policies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `repository` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT 'lca_backup',
  `retention_days` bigint DEFAULT '30',
  `status` tinyint DEFAULT '1',
  `log_store` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '日志库名称',
  `frequency` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'every_day/every_week/every_month',
  `min_count` bigint DEFAULT '5',
  `max_count` bigint DEFAULT '100',
  `cron_expression` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_slm_policies_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `snapshot_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `snapshot_name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `uuid` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `state` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'SUCCESS/IN_PROGRESS/FAILED',
  `repository` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `indices` text COLLATE utf8mb4_unicode_ci COMMENT '索引列表JSON',
  `start_time` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `end_time` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `duration_ms` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_snapshot_records_snapshot_name` (`snapshot_name`),
  KEY `idx_snapshot_records_deleted_at` (`deleted_at`),
  KEY `idx_snapshot_records_repository` (`repository`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `tenants` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` tinyint DEFAULT '1' COMMENT '1-启用 0-禁用',
  `quota_config` text COLLATE utf8mb4_unicode_ci COMMENT '配额配置JSON',
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_tenants_name` (`name`),
  UNIQUE KEY `idx_tenants_code` (`code`),
  KEY `idx_tenants_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_roles` (
  `user_id` bigint unsigned NOT NULL,
  `role_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`user_id`,`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `username` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `nickname` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `avatar` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint DEFAULT '1' COMMENT '1-启用 0-禁用',
  `dept_id` bigint unsigned DEFAULT '0',
  `post_id` bigint unsigned DEFAULT '0',
  `remark` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`),
  KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================================
-- 种子数据（INSERT IGNORE 保证幂等，重复执行不报错）
-- =====================================================================

-- 部门
INSERT IGNORE INTO `departments` (`id`,`created_at`,`updated_at`,`deleted_at`,`parent_id`,`name`,`sort`,`leader`,`phone`,`email`,`status`) VALUES
(1,NOW(),NOW(),NULL,0,'LCA科技',1,'张总','13800000001','ceo@lca.com',1),
(2,NOW(),NOW(),NULL,1,'研发部',1,'李明','13800000002','dev@lca.com',1),
(3,NOW(),NOW(),NULL,1,'运维部',2,'王强','13800000003','ops@lca.com',1),
(4,NOW(),NOW(),NULL,1,'测试部',3,'赵芳','13800000004','qa@lca.com',1),
(5,NOW(),NOW(),NULL,2,'前端组',1,'','','',1),
(6,NOW(),NOW(),NULL,2,'后端组',2,'','','',1);

-- 岗位
INSERT IGNORE INTO `posts` (`id`,`created_at`,`updated_at`,`deleted_at`,`name`,`code`,`sort`,`status`,`remark`) VALUES
(1,NOW(),NOW(),NULL,'董事长','ceo',1,1,''),
(2,NOW(),NOW(),NULL,'项目经理','pm',2,1,''),
(3,NOW(),NOW(),NULL,'高级开发','senior_dev',3,1,''),
(4,NOW(),NOW(),NULL,'开发工程师','dev',4,1,''),
(5,NOW(),NOW(),NULL,'运维工程师','ops',5,1,''),
(6,NOW(),NOW(),NULL,'测试工程师','qa',6,1,'');

-- 菜单（visible=1 全部显示，admin 角色可见所有菜单）
INSERT IGNORE INTO `menus` (`id`,`created_at`,`updated_at`,`deleted_at`,`parent_id`,`name`,`path`,`component`,`icon`,`sort`,`status`,`menu_type`,`visible`,`perms`,`api_path`) VALUES
-- 一级目录
(1, NOW(),NOW(),NULL,0,'首页','/dashboard','','Odometer',1,1,'C',1,'',''),
(2, NOW(),NOW(),NULL,0,'权限管理','/permission','','Key',2,1,'M',1,'',''),
(3, NOW(),NOW(),NULL,0,'系统监控','/monitor','','Monitor',3,1,'M',1,'',''),
(4, NOW(),NOW(),NULL,0,'日志管理','/log','','Document',4,1,'M',1,'',''),
(5, NOW(),NOW(),NULL,0,'备份管理','/backup','','FolderOpened',5,1,'M',1,'',''),
(6, NOW(),NOW(),NULL,0,'日志采集','/collect','','Connection',6,1,'M',1,'',''),
-- 权限管理子菜单
(21,NOW(),NOW(),NULL,2,'用户管理','/permission/users','','User',1,1,'C',1,'permission:user:list',''),
(22,NOW(),NOW(),NULL,2,'角色管理','/permission/roles','','UserFilled',2,1,'C',1,'permission:role:list',''),
(23,NOW(),NOW(),NULL,2,'菜单管理','/permission/menus','','Menu',3,1,'C',1,'permission:menu:list',''),
(24,NOW(),NOW(),NULL,2,'部门管理','/permission/dept','','OfficeBuilding',4,1,'C',1,'permission:dept:list',''),
(25,NOW(),NOW(),NULL,2,'岗位管理','/permission/post','','Postcard',5,1,'C',1,'permission:post:list',''),
-- 系统监控子菜单
(32,NOW(),NOW(),NULL,3,'登录日志','/monitor/loginlog','','Tickets',1,1,'C',1,'monitor:loginlog:list',''),
(33,NOW(),NOW(),NULL,3,'操作日志','/monitor/operlog','','Document',2,1,'C',1,'monitor:operlog:list',''),
(34,NOW(),NOW(),NULL,3,'在线用户','/monitor/online','','Connection',3,1,'C',1,'monitor:online:list',''),
-- 日志管理子菜单
(41,NOW(),NOW(),NULL,4,'日志库','/log/store','','Folder',1,1,'C',1,'log:store:list',''),
(42,NOW(),NOW(),NULL,4,'日志查询','/log/list','','List',2,1,'C',1,'log:list:view',''),
(44,NOW(),NOW(),NULL,4,'告警历史','/log/alert-history','','Bell',3,1,'C',1,'log:alert:history',''),
-- 备份管理子菜单
(51,NOW(),NOW(),NULL,5,'备份列表','/backup/snapshots','','Files',1,1,'C',1,'backup:snapshot:list',''),
(52,NOW(),NOW(),NULL,5,'备份策略','/backup/policies','','SetUp',2,1,'C',1,'backup:policy:list',''),
-- 日志采集子菜单
(61,NOW(),NOW(),NULL,6,'采集任务','/collect/tasks','','Position',1,1,'C',1,'collect:task:list',''),
(62,NOW(),NOW(),NULL,6,'Agent管理','/collect/agents','','Monitor',2,1,'C',1,'collect:agent:list','');

-- 角色
INSERT IGNORE INTO `roles` (`id`,`created_at`,`updated_at`,`deleted_at`,`name`,`code`,`description`,`sort`,`status`) VALUES
(1,NOW(),NOW(),NULL,'管理员','admin','系统管理员，拥有所有权限',1,1),
(2,NOW(),NOW(),NULL,'运维人员','ops','运维人员，可管理日志和备份',2,1),
(3,NOW(),NOW(),NULL,'开发人员','dev','开发人员，可查看日志',3,1),
(4,NOW(),NOW(),NULL,'测试','test','测试人员',4,1);

-- 用户（admin/admin123，其余用户密码 lca@2026，生产环境请修改）
INSERT IGNORE INTO `users` (`id`,`created_at`,`updated_at`,`deleted_at`,`username`,`password_hash`,`nickname`,`email`,`phone`,`avatar`,`status`,`dept_id`,`post_id`,`remark`) VALUES
(1,NOW(),NOW(),NULL,'admin',   '$2a$10$g0o/BZSRVs4gPO6vyOhMvee2QAMmF1PnQu5h1DkOivqvLuKIFViKG','超级管理员','','','',1,1,1,''),
(2,NOW(),NOW(),NULL,'zhangsan','$2a$10$g9gg2x6dLr/YEDDmxWb6y.Ak6zYxO7uZDorClnOUa3DmfdogNLim2','张三',       '','','',1,2,3,''),
(3,NOW(),NOW(),NULL,'lisi',    '$2a$10$CFiOPW5rIZ8W6qFrri4W2uJAsLM3dEHyEXekYRQkspi1rvLn1LX0a','李四',       '','','',1,3,5,''),
(4,NOW(),NOW(),NULL,'wangwu',  '$2a$10$g9gg2x6dLr/YEDDmxWb6y.Ak6zYxO7uZDorClnOUa3DmfdogNLim2','王五',       '','','',1,5,4,''),
(5,NOW(),NOW(),NULL,'zhaoliu', '$2a$10$g9gg2x6dLr/YEDDmxWb6y.Ak6zYxO7uZDorClnOUa3DmfdogNLim2','赵六',       '','','',1,6,4,''),
(6,NOW(),NOW(),NULL,'sunqi',   '$2a$10$63.fupXGwSBN8v0vudnYbu5Jh3E47H2EtqeFXHJCcJoHW3/EdKWj6','孙七',       '','','',1,4,6,'');

-- 用户-角色关联
INSERT IGNORE INTO `user_roles` (`user_id`,`role_id`) VALUES
(1,1),(2,2),(3,2),(4,3),(5,3),(6,4);

-- 角色-菜单关联（管理员拥有全部菜单）
INSERT IGNORE INTO `role_menus` (`role_id`,`menu_id`) VALUES
-- 管理员：全部
(1,1),(1,2),(1,3),(1,4),(1,5),(1,6),
(1,21),(1,22),(1,23),(1,24),(1,25),
(1,32),(1,33),(1,34),
(1,41),(1,42),(1,44),
(1,51),(1,52),
(1,61),(1,62),
-- 运维：日志+备份+采集
(2,4),(2,5),(2,6),(2,41),(2,42),(2,44),(2,51),(2,52),(2,61),(2,62),
-- 开发：日志查询+告警历史
(3,42),(3,44),
-- 测试：日志查询
(4,42);

-- 日志库（3个示例，敏感配置留空，部署后在界面填写）
INSERT IGNORE INTO `log_stores` (`id`,`created_at`,`updated_at`,`deleted_at`,`name`,`description`,`index_pattern`,`api_key`,`kafka_topic`,`status`,`compress`,`roll_max_days`,`roll_max_size_gb`,`cold_days`,`delete_days`,`ai_alert_enabled`,`ai_alert_config`,`oss_repository`,`oss_endpoint`,`oss_bucket`,`oss_access_key_id`,`oss_access_key_secret`,`oss_path`,`oss_chunk_size`) VALUES
(1,NOW(),NOW(),NULL,'app-nginx',   'Nginx访问日志',   'app-nginx-*',   'ak_nginx_001',   'lca_app-nginx',   1,1,7, 50,30, 90,0,NULL,'','','','','','lca/','500mb'),
(2,NOW(),NOW(),NULL,'app-backend', '后端应用日志',   'app-backend-*', 'ak_backend_002', 'lca_app-backend', 1,1,7,  0,15, 60,0,NULL,'','','','','','',    '500mb'),
(3,NOW(),NOW(),NULL,'app-security','安全审计日志',   'app-security-*','ak_security_003','lca_app-security',1,1,0,  0, 0,180,0,NULL,'','','','','','',    '500mb');
