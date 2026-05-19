package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/middleware"
	"github.com/cj/log-collect-ai-analytics/internal/model"
)

func init() {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret-key", 24, "test")
}

// setupTestDB 创建内存SQLite测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Department{},
		&model.Post{},
		&model.Menu{},
		&model.AuditLog{},
		&model.LoginLog{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

// seedTestUser 创建测试用户并返回
func seedTestUser(t *testing.T, db *gorm.DB) model.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := model.User{
		Username:     "admin",
		PasswordHash: string(hash),
		Nickname:     "管理员",
		Status:       1,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	return user
}

// setupAuthRouter 创建带认证的测试路由
func setupAuthRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	captcha := NewCaptchaHandler()
	authHandler := NewAuthHandler(db, captcha)
	userHandler := NewUserHandler(db)
	roleHandler := NewRoleHandler(db)
	deptHandler := NewDeptHandler(db)
	postHandler := NewPostHandler(db)
	menuHandler := NewMenuHandler(db)
	operLogHandler := NewOperLogHandler(db)
	loginLogHandler := NewLoginLogHandler(db)
	onlineHandler := NewOnlineHandler(db)
	monitorHandler := NewMonitorHandler()

	// 公开路由
	r.GET("/api/v1/captcha", captcha.Generate)
	r.POST("/api/v1/auth/login", authHandler.Login)

	// 认证路由
	auth := r.Group("/api/v1").Use(middleware.JWTAuth())
	{
		auth.GET("/auth/userinfo", authHandler.GetUserInfo)
		auth.PUT("/auth/password", authHandler.ChangePassword)

		auth.GET("/users", userHandler.ListUsers)
		auth.POST("/users", userHandler.CreateUser)
		auth.PUT("/users/:id", userHandler.UpdateUser)
		auth.DELETE("/users/:id", userHandler.DeleteUser)
		auth.PUT("/users/:id/reset-password", userHandler.ResetPassword)
		auth.PUT("/users/:id/status", userHandler.UpdateStatus)

		auth.GET("/roles", roleHandler.List)
		auth.POST("/roles", roleHandler.Create)
		auth.PUT("/roles/:id", roleHandler.Update)
		auth.DELETE("/roles/:id", roleHandler.Delete)
		auth.PUT("/roles/:id/menus", roleHandler.AssignMenus)

		auth.GET("/depts", deptHandler.List)
		auth.POST("/depts", deptHandler.Create)
		auth.PUT("/depts/:id", deptHandler.Update)
		auth.DELETE("/depts/:id", deptHandler.Delete)

		auth.GET("/posts", postHandler.List)
		auth.POST("/posts", postHandler.Create)
		auth.PUT("/posts/:id", postHandler.Update)
		auth.DELETE("/posts/:id", postHandler.Delete)

		auth.GET("/menus", menuHandler.List)
		auth.POST("/menus", menuHandler.Create)
		auth.PUT("/menus/:id", menuHandler.Update)
		auth.DELETE("/menus/:id", menuHandler.Delete)

		auth.GET("/monitor/server", monitorHandler.ServerInfo)
		auth.GET("/loginlog", loginLogHandler.List)
		auth.DELETE("/loginlog", loginLogHandler.Delete)
		auth.GET("/operlog", operLogHandler.List)
		auth.DELETE("/operlog", operLogHandler.Delete)
		auth.GET("/online", onlineHandler.List)
		auth.DELETE("/online/:id", onlineHandler.ForceLogout)
	}
	return r
}

// doRequest 执行HTTP请求
func doRequest(r *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// parseResponse 解析响应JSON
func parseResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response failed: %v, body: %s", err, w.Body.String())
	}
	return resp
}

// getToken 获取测试token
func getToken(t *testing.T, userID uint, username string) string {
	token, err := middleware.GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}
	return token
}

// assertSuccess 断言请求成功
func assertSuccess(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
	resp := parseResponse(t, w)
	code, _ := resp["code"].(float64)
	if code != 0 {
		t.Fatalf("expected code 0, got %v, message: %v", code, resp["message"])
	}
	return resp
}

// assertError 断言请求失败
func assertError(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	resp := parseResponse(t, w)
	code, _ := resp["code"].(float64)
	if code == 0 {
		t.Fatalf("expected error code, got 0")
	}
	return resp
}
