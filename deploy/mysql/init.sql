-- LCA 日志收集智能分析系统 - 数据库初始化
-- 表结构由 GORM AutoMigrate 自动创建，此文件负责：基础参数 + 种子数据

CREATE DATABASE IF NOT EXISTS lca DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE lca;

SET GLOBAL max_connections = 500;
SET GLOBAL innodb_buffer_pool_size = 268435456;

-- =====================================================================
-- 以下种子数据在服务首次启动、AutoMigrate 完成后执行
-- 使用 INSERT IGNORE 保证幂等，重复执行不报错
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

-- 用户（密码均为 admin123，生产环境请修改）
-- admin123 的 bcrypt hash:
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
