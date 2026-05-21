package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/handler"
	"github.com/cj/log-collect-ai-analytics/internal/middleware"
	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/config"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/license"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/logger"
	redispkg "github.com/cj/log-collect-ai-analytics/internal/pkg/redis"
	"github.com/cj/log-collect-ai-analytics/internal/service"
)

func main() {
	configFile := flag.String("config", "configs/admin.yaml", "config file path")
	flag.Parse()

	// 1. 加载配置
	if err := config.Init(*configFile); err != nil {
		fmt.Printf("init config failed: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	cfg := config.Get()
	logger.InitWithConfig(logger.LogConfig{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	})

	// 3. 初始化数据库
	dsn := cfg.MySQL.DSN()
	if err := model.InitDB(dsn,
		model.WithMaxOpenConns(cfg.MySQL.MaxOpenConns),
		model.WithMaxIdleConns(cfg.MySQL.MaxIdleConns),
		model.WithConnMaxLifetime(cfg.MySQL.ConnMaxLifetime),
	); err != nil {
		fmt.Printf("init db failed: %v\n", err)
		os.Exit(1)
	}

	// 4. 自动迁移
	if err := model.AutoMigrate(); err != nil {
		fmt.Printf("auto migrate failed: %v\n", err)
		os.Exit(1)
	}

	// 4.5 初始化读副本（可选）
	if readDSN := cfg.MySQL.ReadDSN(); readDSN != "" {
		if err := model.InitReadDB(readDSN,
			model.WithMaxOpenConns(cfg.MySQL.MaxOpenConns),
			model.WithMaxIdleConns(cfg.MySQL.MaxIdleConns),
			model.WithConnMaxLifetime(cfg.MySQL.ConnMaxLifetime),
		); err != nil {
			fmt.Printf("warning: init read replica failed: %v\n", err)
		}
	}

	// 5. 初始化Redis
	redisOpts := []redispkg.Option{
		redispkg.WithPoolSize(cfg.Redis.PoolSize),
		redispkg.WithDialTimeout(cfg.Redis.DialTimeout),
		redispkg.WithReadTimeout(cfg.Redis.ReadTimeout),
		redispkg.WithWriteTimeout(cfg.Redis.WriteTimeout),
		redispkg.WithMaxRetries(cfg.Redis.MaxRetries),
		redispkg.WithMode(cfg.Redis.Mode),
	}
	if cfg.Redis.MasterName != "" {
		redisOpts = append(redisOpts, redispkg.WithSentinel(cfg.Redis.MasterName, cfg.Redis.SentinelAddrs))
	}
	if err := redispkg.Init(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB, redisOpts...); err != nil {
		fmt.Printf("warning: init redis failed: %v\n", err)
	}

	// 6. 初始化JWT
	middleware.InitJWT(cfg.JWT.Secret, cfg.JWT.ExpireHour, cfg.JWT.Issuer)
	middleware.InitLicenseDB(model.GetDB())

	// 7. 初始化Casbin
	if err := middleware.InitCasbin(model.GetDB()); err != nil {
		fmt.Printf("init casbin failed: %v\n", err)
		os.Exit(1)
	}
	// 同步用户-角色关系到 Casbin
	middleware.SyncUserRoles(model.GetDB())
	// 启动 Casbin 策略同步订阅（多实例间 Redis Pub/Sub）
	middleware.StartCasbinSubscriber()

	// 6.5 初始化License验签器
	var licenseVerifier *license.Verifier
	if cfg.License.PublicKey != "" {
		var err error
		licenseVerifier, err = license.NewVerifier(cfg.License.PublicKey)
		if err != nil {
			fmt.Printf("init license verifier failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("warning: license public_key not configured, license check disabled")
	}

	// 将验签器注入 LicenseCheck 中间件，每个请求都从 license_key 密文重新验签
	// 防御：客户手改 MySQL licenses 表的 status / expires_at 不生效
	middleware.InitLicenseVerifier(licenseVerifier)

	// 7. 初始化ES客户端（可选，连不上不影响启动）
	var esClient *elastic.Client
	if len(cfg.ES.Addresses) > 0 {
		var err error
		esClient, err = elastic.NewClient(
			elastic.SetURL(cfg.ES.Addresses...),
			elastic.SetSniff(false),
		)
		if err != nil {
			fmt.Printf("warning: ES connection failed: %v\n", err)
		}
	}

	// 8. 初始化Handlers
	db := model.GetDB()
	captchaHandler := handler.NewCaptchaHandler()
	authHandler := handler.NewAuthHandler(db, captchaHandler)
	userHandler := handler.NewUserHandler(db)
	roleHandler := handler.NewRoleHandler(db)
	deptHandler := handler.NewDeptHandler(db)
	postHandler := handler.NewPostHandler(db)
	menuHandler := handler.NewMenuHandler(db)
	monitorHandler := handler.NewMonitorHandler()
	loginLogHandler := handler.NewLoginLogHandler(db)
	operLogHandler := handler.NewOperLogHandler(db)
	onlineHandler := handler.NewOnlineHandler(db)
	logStoreHandler := handler.NewLogStoreHandler(db, esClient)
	esLogHandler := handler.NewESLogHandler(db, esClient)
	dashboardHandler := handler.NewDashboardHandler(db, esClient)
	backupHandler := handler.NewBackupHandler(db, esClient)
	slmHandler := handler.NewSLMHandler(db, esClient)
	collectHandler := handler.NewCollectHandler(db)
	aiAlertHandler := handler.NewAIAlertHandler(db, esClient)
	licenseHandler := handler.NewLicenseHandler(db, licenseVerifier, cfg.LCATopURL)

	// 9. 设置路由
	mode := cfg.Server.Mode
	if mode == "" {
		mode = "debug"
	}
	gin.SetMode(mode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Cors())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 公开路由
	r.GET("/api/v1/captcha", captchaHandler.Generate)
	r.POST("/api/v1/auth/login", authHandler.Login)

	// Agent公共接口（无需认证，Agent使用api_key鉴权）
	r.POST("/api/v1/agents/heartbeat", collectHandler.Heartbeat)
	r.GET("/api/v1/agents/tasks", collectHandler.GetTasksForAgent)
	r.GET("/api/v1/agents/offsets", collectHandler.GetOffsets)
	r.POST("/api/v1/agents/offsets", collectHandler.UpdateOffsets)

	// 需要认证的路由（不需要授权码）
	auth := r.Group("/api/v1").Use(middleware.JWTAuth(), middleware.AuditLog())
	{
		// 授权码接口
		auth.GET("/license/status", licenseHandler.Status)
		auth.POST("/license/activate", licenseHandler.Activate)
		auth.DELETE("/license", licenseHandler.Deactivate)

		// 用户基本信息（登录后必须能获取，不受授权码限制）
		auth.GET("/auth/userinfo", authHandler.GetUserInfo)
		auth.PUT("/auth/profile", authHandler.UpdateProfile)
		auth.PUT("/auth/password", authHandler.ChangePassword)
	}

	// 需要认证+授权码的路由
	authWithLicense := r.Group("/api/v1").Use(middleware.JWTAuth(), middleware.AuditLog(), middleware.LicenseCheck())
	{

		// 用户管理
		authWithLicense.GET("/users", userHandler.ListUsers)
		authWithLicense.POST("/users", userHandler.CreateUser)
		authWithLicense.PUT("/users/:id", userHandler.UpdateUser)
		authWithLicense.DELETE("/users/:id", userHandler.DeleteUser)
		authWithLicense.PUT("/users/:id/reset-password", userHandler.ResetPassword)
		authWithLicense.PUT("/users/:id/status", userHandler.UpdateStatus)

		// 角色管理
		authWithLicense.GET("/roles", roleHandler.List)
		authWithLicense.POST("/roles", roleHandler.Create)
		authWithLicense.PUT("/roles/:id", roleHandler.Update)
		authWithLicense.DELETE("/roles/:id", roleHandler.Delete)
		authWithLicense.PUT("/roles/:id/menus", roleHandler.AssignMenus)
		authWithLicense.GET("/roles/:id/menus", roleHandler.GetMenus)

		// 部门管理
		authWithLicense.GET("/depts", deptHandler.List)
		authWithLicense.POST("/depts", deptHandler.Create)
		authWithLicense.PUT("/depts/:id", deptHandler.Update)
		authWithLicense.DELETE("/depts/:id", deptHandler.Delete)

		// 岗位管理
		authWithLicense.GET("/posts", postHandler.List)
		authWithLicense.POST("/posts", postHandler.Create)
		authWithLicense.PUT("/posts/:id", postHandler.Update)
		authWithLicense.DELETE("/posts/:id", postHandler.Delete)

		// 菜单管理
		authWithLicense.GET("/menus", menuHandler.List)
		authWithLicense.GET("/menus/user", menuHandler.UserMenus)
		authWithLicense.POST("/menus", menuHandler.Create)
		authWithLicense.PUT("/menus/:id", menuHandler.Update)
		authWithLicense.DELETE("/menus/:id", menuHandler.Delete)

		// 系统监控
		authWithLicense.GET("/monitor/server", monitorHandler.ServerInfo)

		// 登录日志
		authWithLicense.GET("/loginlog", loginLogHandler.List)
		authWithLicense.DELETE("/loginlog", loginLogHandler.Delete)
		authWithLicense.DELETE("/loginlog/clean", loginLogHandler.Clean)

		// 操作日志
		authWithLicense.GET("/operlog", operLogHandler.List)
		authWithLicense.DELETE("/operlog", operLogHandler.Delete)

		// 在线用户
		authWithLicense.GET("/online", onlineHandler.List)
		authWithLicense.DELETE("/online/:id", onlineHandler.ForceLogout)

		// 日志库
		authWithLicense.GET("/logstore", logStoreHandler.List)
		authWithLicense.POST("/logstore", logStoreHandler.Create)
		authWithLicense.PUT("/logstore/:id", logStoreHandler.Update)
		authWithLicense.DELETE("/logstore/:id", logStoreHandler.Delete)

		// AI 智能告警
		authWithLicense.PUT("/logstore/:id/ai-alert", aiAlertHandler.Toggle)
		authWithLicense.GET("/logstore/:id/ai-alert/config", aiAlertHandler.GetConfig)
		authWithLicense.PUT("/logstore/:id/ai-alert/config", aiAlertHandler.UpdateConfig)
		authWithLicense.POST("/logstore/:id/ai-alert/test", aiAlertHandler.Test)
		authWithLicense.GET("/ai-alert/history", aiAlertHandler.History)
		authWithLicense.DELETE("/ai-alert/history/:id", aiAlertHandler.DeleteHistory)
		authWithLicense.DELETE("/ai-alert/history", aiAlertHandler.ClearHistory)

		// 日志查询
		authWithLicense.GET("/eslog", esLogHandler.Search)
		authWithLicense.GET("/eslog/fields", esLogHandler.Fields)
		authWithLicense.GET("/eslog/histogram", esLogHandler.Histogram)

		// 首页统计
		authWithLicense.GET("/dashboard", dashboardHandler.Stats)

		// 备份管理
		authWithLicense.GET("/backup/snapshots", backupHandler.ListSnapshots)
		authWithLicense.DELETE("/backup/snapshots/:name", backupHandler.DeleteSnapshot)
		authWithLicense.POST("/backup/snapshots/:name/restore", backupHandler.RestoreSnapshot)

		// 备份策略
		authWithLicense.GET("/backup/policies", slmHandler.List)
		authWithLicense.POST("/backup/policies", slmHandler.Create)
		authWithLicense.PUT("/backup/policies/:id", slmHandler.Update)
		authWithLicense.DELETE("/backup/policies/:id", slmHandler.Delete)
		authWithLicense.POST("/backup/policies/:id/execute", slmHandler.Execute)

		// 采集任务管理
		authWithLicense.GET("/collect/tasks", collectHandler.ListTasks)
		authWithLicense.POST("/collect/tasks", collectHandler.CreateTask)
		authWithLicense.PUT("/collect/tasks/:id", collectHandler.UpdateTask)
		authWithLicense.DELETE("/collect/tasks/:id", collectHandler.DeleteTask)

		// Agent管理
		authWithLicense.GET("/agents", collectHandler.ListAgents)
		authWithLicense.DELETE("/agents/:id", collectHandler.DeleteAgent)
	}

	// 10. 初始化默认数据
	initDefaultData(db)

	// 10.5 启动快照定时同步
	backupHandler.StartSnapshotSync()

	// 启动AI智能告警调度器
	aiAlertScheduler := service.NewAIAlertScheduler(db, esClient)
	aiAlertScheduler.Start()

	// 启动Agent离线监控通知器
	agentOfflineNotifier := service.NewAgentOfflineNotifier(db)
	agentOfflineNotifier.Start()

	// 11. 启动服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8080
	}
	logger.Info(fmt.Sprintf("Admin server starting on :%d", port))

	go func() {
		if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
			logger.Fatal(fmt.Sprintf("server start failed: %v", err))
		}
	}()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")
	aiAlertScheduler.Stop()
	agentOfflineNotifier.Stop()
}

func initDefaultData(db *gorm.DB) {
	var count int64

	// === 部门 ===
	db.Model(&model.Department{}).Count(&count)
	if count == 0 {
		depts := []model.Department{
			{BaseModel: model.BaseModel{ID: 1}, ParentID: 0, Name: "LCA科技", Sort: 1, Status: 1, Leader: "张总", Phone: "13800000001", Email: "ceo@lca.com"},
			{BaseModel: model.BaseModel{ID: 2}, ParentID: 1, Name: "研发部", Sort: 1, Status: 1, Leader: "李明", Phone: "13800000002", Email: "dev@lca.com"},
			{BaseModel: model.BaseModel{ID: 3}, ParentID: 1, Name: "运维部", Sort: 2, Status: 1, Leader: "王强", Phone: "13800000003", Email: "ops@lca.com"},
			{BaseModel: model.BaseModel{ID: 4}, ParentID: 1, Name: "测试部", Sort: 3, Status: 1, Leader: "赵芳", Phone: "13800000004", Email: "qa@lca.com"},
			{BaseModel: model.BaseModel{ID: 5}, ParentID: 2, Name: "前端组", Sort: 1, Status: 1},
			{BaseModel: model.BaseModel{ID: 6}, ParentID: 2, Name: "后端组", Sort: 2, Status: 1},
		}
		for i := range depts {
			db.Create(&depts[i])
		}
	}

	// === 岗位 ===
	db.Model(&model.Post{}).Count(&count)
	if count == 0 {
		posts := []model.Post{
			{BaseModel: model.BaseModel{ID: 1}, Name: "董事长", Code: "ceo", Sort: 1, Status: 1},
			{BaseModel: model.BaseModel{ID: 2}, Name: "项目经理", Code: "pm", Sort: 2, Status: 1},
			{BaseModel: model.BaseModel{ID: 3}, Name: "高级开发", Code: "senior_dev", Sort: 3, Status: 1},
			{BaseModel: model.BaseModel{ID: 4}, Name: "开发工程师", Code: "dev", Sort: 4, Status: 1},
			{BaseModel: model.BaseModel{ID: 5}, Name: "运维工程师", Code: "ops", Sort: 5, Status: 1},
			{BaseModel: model.BaseModel{ID: 6}, Name: "测试工程师", Code: "qa", Sort: 6, Status: 1},
		}
		for i := range posts {
			db.Create(&posts[i])
		}
	}

	// === 菜单 ===
	db.Model(&model.Menu{}).Count(&count)
	if count == 0 {
		menus := []model.Menu{
			// 首页
			{BaseModel: model.BaseModel{ID: 1}, ParentID: 0, Name: "首页", Path: "/dashboard", Icon: "Odometer", Sort: 1, MenuType: "C", Visible: 1, Status: 1},
			// 权限管理
			{BaseModel: model.BaseModel{ID: 2}, ParentID: 0, Name: "权限管理", Path: "/permission", Icon: "Key", Sort: 2, MenuType: "M", Visible: 1, Status: 1},
			{BaseModel: model.BaseModel{ID: 21}, ParentID: 2, Name: "用户管理", Path: "/permission/users", Icon: "User", Sort: 1, MenuType: "C", Visible: 1, Status: 1, Perms: "permission:user:list"},
			{BaseModel: model.BaseModel{ID: 22}, ParentID: 2, Name: "角色管理", Path: "/permission/roles", Icon: "UserFilled", Sort: 2, MenuType: "C", Visible: 1, Status: 1, Perms: "permission:role:list"},
			{BaseModel: model.BaseModel{ID: 23}, ParentID: 2, Name: "菜单管理", Path: "/permission/menus", Icon: "Menu", Sort: 3, MenuType: "C", Visible: 1, Status: 1, Perms: "permission:menu:list"},
			{BaseModel: model.BaseModel{ID: 24}, ParentID: 2, Name: "部门管理", Path: "/permission/dept", Icon: "OfficeBuilding", Sort: 4, MenuType: "C", Visible: 1, Status: 1, Perms: "permission:dept:list"},
			{BaseModel: model.BaseModel{ID: 25}, ParentID: 2, Name: "岗位管理", Path: "/permission/post", Icon: "Postcard", Sort: 5, MenuType: "C", Visible: 1, Status: 1, Perms: "permission:post:list"},
			// 系统监控
			{BaseModel: model.BaseModel{ID: 3}, ParentID: 0, Name: "系统监控", Path: "/monitor", Icon: "Monitor", Sort: 3, MenuType: "M", Visible: 1, Status: 1},
			{BaseModel: model.BaseModel{ID: 32}, ParentID: 3, Name: "登录日志", Path: "/monitor/loginlog", Icon: "Tickets", Sort: 1, MenuType: "C", Visible: 1, Status: 1, Perms: "monitor:loginlog:list"},
			{BaseModel: model.BaseModel{ID: 33}, ParentID: 3, Name: "操作日志", Path: "/monitor/operlog", Icon: "Document", Sort: 2, MenuType: "C", Visible: 1, Status: 1, Perms: "monitor:operlog:list"},
			{BaseModel: model.BaseModel{ID: 34}, ParentID: 3, Name: "在线用户", Path: "/monitor/online", Icon: "Connection", Sort: 3, MenuType: "C", Visible: 1, Status: 1, Perms: "monitor:online:list"},
			// 日志管理
			{BaseModel: model.BaseModel{ID: 4}, ParentID: 0, Name: "日志管理", Path: "/log", Icon: "Document", Sort: 4, MenuType: "M", Visible: 1, Status: 1},
			{BaseModel: model.BaseModel{ID: 41}, ParentID: 4, Name: "日志库", Path: "/log/store", Icon: "Folder", Sort: 1, MenuType: "C", Visible: 1, Status: 1, Perms: "log:store:list"},
			{BaseModel: model.BaseModel{ID: 42}, ParentID: 4, Name: "日志查询", Path: "/log/list", Icon: "List", Sort: 2, MenuType: "C", Visible: 1, Status: 1, Perms: "log:list:view"},
			// 备份管理
			{BaseModel: model.BaseModel{ID: 5}, ParentID: 0, Name: "备份管理", Path: "/backup", Icon: "FolderOpened", Sort: 5, MenuType: "M", Visible: 1, Status: 1},
			{BaseModel: model.BaseModel{ID: 51}, ParentID: 5, Name: "备份列表", Path: "/backup/snapshots", Icon: "Files", Sort: 1, MenuType: "C", Visible: 1, Status: 1, Perms: "backup:snapshot:list"},
			{BaseModel: model.BaseModel{ID: 52}, ParentID: 5, Name: "备份策略", Path: "/backup/policies", Icon: "SetUp", Sort: 2, MenuType: "C", Visible: 1, Status: 1, Perms: "backup:policy:list"},
			// 日志采集
			{BaseModel: model.BaseModel{ID: 6}, ParentID: 0, Name: "日志采集", Path: "/collect", Icon: "Connection", Sort: 6, MenuType: "M", Visible: 1, Status: 1},
			{BaseModel: model.BaseModel{ID: 61}, ParentID: 6, Name: "采集任务", Path: "/collect/tasks", Icon: "Position", Sort: 1, MenuType: "C", Visible: 1, Status: 1, Perms: "collect:task:list"},
			{BaseModel: model.BaseModel{ID: 62}, ParentID: 6, Name: "Agent管理", Path: "/collect/agents", Icon: "Monitor", Sort: 2, MenuType: "C", Visible: 1, Status: 1, Perms: "collect:agent:list"},
		}
		for i := range menus {
			db.Create(&menus[i])
		}
	}

	// === 角色 ===
	db.Model(&model.Role{}).Count(&count)
	if count == 0 {
		roles := []model.Role{
			{BaseModel: model.BaseModel{ID: 1}, Name: "管理员", Code: "admin", Sort: 1, Status: 1, Description: "系统管理员，拥有所有权限"},
			{BaseModel: model.BaseModel{ID: 2}, Name: "运维人员", Code: "ops", Sort: 2, Status: 1, Description: "运维人员，可管理日志和备份"},
			{BaseModel: model.BaseModel{ID: 3}, Name: "开发人员", Code: "dev", Sort: 3, Status: 1, Description: "开发人员，可查看日志"},
			{BaseModel: model.BaseModel{ID: 4}, Name: "只读用户", Code: "viewer", Sort: 4, Status: 1, Description: "只读用户，只能查看"},
		}
		for i := range roles {
			db.Create(&roles[i])
		}
		// 给管理员角色分配所有菜单
		var allMenus []model.Menu
		db.Find(&allMenus)
		db.Model(&roles[0]).Association("Menus").Replace(allMenus)
		// 运维角色：首页+日志采集+日志管理+备份管理+系统监控(无服务监控)
		var opsMenus []model.Menu
		db.Where("id IN ?", []uint{1, 3, 32, 33, 34, 4, 41, 42, 5, 51, 52, 6, 61, 62}).Find(&opsMenus)
		db.Model(&roles[1]).Association("Menus").Replace(opsMenus)
		// 开发角色：首页+日志查看
		var devMenus []model.Menu
		db.Where("id IN ?", []uint{1, 4, 41, 42}).Find(&devMenus)
		db.Model(&roles[2]).Association("Menus").Replace(devMenus)
		// 只读：首页
		var viewerMenus []model.Menu
		db.Where("id IN ?", []uint{1}).Find(&viewerMenus)
		db.Model(&roles[3]).Association("Menus").Replace(viewerMenus)
	}

	// === 用户 ===
	db.Model(&model.User{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		admin := model.User{
			Username:     "admin",
			PasswordHash: string(hash),
			Nickname:     "超级管理员",
			Status:       1,
			DeptID:       1,
			PostID:       1,
		}
		db.Create(&admin)
		// 分配管理员角色
		var adminRole model.Role
		db.First(&adminRole, 1)
		db.Model(&admin).Association("Roles").Replace([]model.Role{adminRole})
	}

	// 创建测试用户
	db.Model(&model.User{}).Where("username = ?", "zhangsan").Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		testUsers := []struct {
			Username string
			Nickname string
			DeptID   uint
			PostID   uint
			RoleID   uint
		}{
			{"zhangsan", "张三", 2, 3, 2},
			{"lisi", "李四", 3, 5, 2},
			{"wangwu", "王五", 5, 4, 3},
			{"zhaoliu", "赵六", 6, 4, 3},
			{"sunqi", "孙七", 4, 6, 4},
		}
		for _, u := range testUsers {
			user := model.User{
				Username:     u.Username,
				PasswordHash: string(hash),
				Nickname:     u.Nickname,
				Status:       1,
				DeptID:       u.DeptID,
				PostID:       u.PostID,
			}
			db.Create(&user)
			var role model.Role
			db.First(&role, u.RoleID)
			db.Model(&user).Association("Roles").Replace([]model.Role{role})
		}
	}

	// === 增量菜单迁移：确保 Agent管理 菜单存在 ===
	var agentMenuCount int64
	db.Model(&model.Menu{}).Where("id = ?", 62).Count(&agentMenuCount)
	if agentMenuCount == 0 {
		agentMenu := model.Menu{
			BaseModel: model.BaseModel{ID: 62},
			ParentID:  6,
			Name:      "Agent管理",
			Path:      "/collect/agents",
			Icon:      "Monitor",
			Sort:      2,
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "collect:agent:list",
		}
		db.Create(&agentMenu)
		// 追加到管理员角色的菜单中
		var adminRole model.Role
		if err := db.Preload("Menus").First(&adminRole, 1).Error; err == nil {
			db.Model(&adminRole).Association("Menus").Append(agentMenu)
		}
	}

	// === 增量菜单迁移：移除 服务监控 菜单 ===
	db.Where("id = ?", 31).Delete(&model.Menu{})
	// 从角色菜单关联中清理
	db.Exec("DELETE FROM role_menus WHERE menu_id = ?", 31)

	// === 增量菜单迁移：移除 Kibana iframe 菜单，更新日志列表→日志查询 ===
	db.Where("id = ?", 43).Delete(&model.Menu{})
	db.Exec("DELETE FROM role_menus WHERE menu_id = ?", 43)
	db.Model(&model.Menu{}).Where("id = ?", 42).Update("name", "日志查询")

	// === 增量菜单迁移：确保 告警历史 菜单存在（日志管理下）===
	var alertHistoryMenuCount int64
	db.Model(&model.Menu{}).Where("id = ?", 44).Count(&alertHistoryMenuCount)
	if alertHistoryMenuCount == 0 {
		alertHistoryMenu := model.Menu{
			BaseModel: model.BaseModel{ID: 44},
			ParentID:  4,
			Name:      "告警历史",
			Path:      "/log/alert-history",
			Icon:      "Bell",
			Sort:      3,
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "log:alert:history",
		}
		db.Create(&alertHistoryMenu)
		// 追加到管理员和运维角色的菜单中
		var adminRole model.Role
		if err := db.Preload("Menus").First(&adminRole, 1).Error; err == nil {
			db.Model(&adminRole).Association("Menus").Append(alertHistoryMenu)
		}
		var opsRole model.Role
		if err := db.Preload("Menus").First(&opsRole, 2).Error; err == nil {
			db.Model(&opsRole).Association("Menus").Append(alertHistoryMenu)
		}
	}

	// === 日志库测试数据 ===
	db.Model(&model.LogStore{}).Count(&count)
	if count == 0 {
		stores := []model.LogStore{
			{Name: "app-nginx", IndexPattern: "app-nginx-*", DeleteDays: 90, Compress: true, RollMaxDays: 7, RollMaxSizeGB: 50, ColdDays: 30, Description: "Nginx访问日志", APIKey: "ak_nginx_001", KafkaTopic: "lca_app-nginx", OSSRepository: "my_oss_backup", OSSEndpoint: "http://oss-cn-beijing.aliyuncs.com", OSSBucket: "max-standard", OSSPath: "lca/", OSSChunkSize: "500mb"},
			{Name: "app-backend", IndexPattern: "app-backend-*", DeleteDays: 60, Compress: true, RollMaxDays: 7, ColdDays: 15, Description: "后端应用日志", APIKey: "ak_backend_002", KafkaTopic: "lca_app-backend"},
			{Name: "app-security", IndexPattern: "app-security-*", DeleteDays: 180, Compress: true, Description: "安全审计日志", APIKey: "ak_security_003", KafkaTopic: "lca_app-security"},
		}
		for i := range stores {
			db.Create(&stores[i])
		}
	}
}
