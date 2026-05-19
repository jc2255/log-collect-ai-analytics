package handler

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestCaptchaGenerate(t *testing.T) {
	db := setupTestDB(t)
	r := setupAuthRouter(db)

	w := doRequest(r, "GET", "/api/v1/captcha", nil, "")
	resp := assertSuccess(t, w)

	data := resp["data"].(map[string]interface{})
	if data["captcha_id"] == nil || data["captcha_id"] == "" {
		t.Fatal("captcha_id should not be empty")
	}
	if data["captcha_image"] == nil || data["captcha_image"] == "" {
		t.Fatal("captcha_image should not be empty")
	}
}

func TestLoginSuccess(t *testing.T) {
	db := setupTestDB(t)
	seedTestUser(t, db)
	r := setupAuthRouter(db)

	// 先获取验证码
	w := doRequest(r, "GET", "/api/v1/captcha", nil, "")
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["captcha_id"] == nil || data["captcha_id"] == "" {
		t.Fatal("captcha_id should not be empty")
	}

	// 使用已知验证码创建登录
	captcha := NewCaptchaHandler()
	captcha.store.Store("test-captcha-id", captchaEntry{
		Code:      "1234",
		ExpiredAt: time.Now().Add(5 * time.Minute),
	})

	r2 := setupRouterWithCaptcha(db, captcha)

	body := map[string]string{
		"username":     "admin",
		"password":     "123456",
		"captcha_id":   "test-captcha-id",
		"captcha_code": "1234",
	}
	w = doRequest(r2, "POST", "/api/v1/auth/login", body, "")
	resp = assertSuccess(t, w)
	loginData := resp["data"].(map[string]interface{})
	if loginData["token"] == nil || loginData["token"] == "" {
		t.Fatal("login should return token")
	}
	if loginData["username"] != "admin" {
		t.Fatalf("expected username admin, got %v", loginData["username"])
	}
}

func TestLoginWrongPassword(t *testing.T) {
	db := setupTestDB(t)
	seedTestUser(t, db)
	captcha := NewCaptchaHandler()
	captcha.store.Store("test-id", captchaEntry{Code: "1234", ExpiredAt: time.Now().Add(5 * time.Minute)})
	r := setupRouterWithCaptcha(db, captcha)

	body := map[string]string{
		"username":     "admin",
		"password":     "wrong",
		"captcha_id":   "test-id",
		"captcha_code": "1234",
	}
	w := doRequest(r, "POST", "/api/v1/auth/login", body, "")
	assertError(t, w)
}

func TestLoginWrongCaptcha(t *testing.T) {
	db := setupTestDB(t)
	seedTestUser(t, db)
	captcha := NewCaptchaHandler()
	captcha.store.Store("test-id", captchaEntry{Code: "1234", ExpiredAt: time.Now().Add(5 * time.Minute)})
	r := setupRouterWithCaptcha(db, captcha)

	body := map[string]string{
		"username":     "admin",
		"password":     "123456",
		"captcha_id":   "test-id",
		"captcha_code": "9999",
	}
	w := doRequest(r, "POST", "/api/v1/auth/login", body, "")
	resp := assertError(t, w)
	if resp["message"] != "验证码错误" {
		t.Fatalf("expected '验证码错误', got %v", resp["message"])
	}
}

func TestGetUserInfo(t *testing.T) {
	db := setupTestDB(t)
	user := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, user.ID, user.Username)

	w := doRequest(r, "GET", "/api/v1/auth/userinfo", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["username"] != "admin" {
		t.Fatalf("expected admin, got %v", data["username"])
	}
}

func TestGetUserInfoUnauthorized(t *testing.T) {
	db := setupTestDB(t)
	r := setupAuthRouter(db)

	w := doRequest(r, "GET", "/api/v1/auth/userinfo", nil, "")
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestChangePassword(t *testing.T) {
	db := setupTestDB(t)
	user := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, user.ID, user.Username)

	body := map[string]string{
		"old_password": "123456",
		"new_password": "654321",
	}
	w := doRequest(r, "PUT", "/api/v1/auth/password", body, token)
	assertSuccess(t, w)

	// 用旧密码登录应该失败
	captcha := NewCaptchaHandler()
	captcha.store.Store("cap1", captchaEntry{Code: "1111", ExpiredAt: time.Now().Add(5 * time.Minute)})
	r2 := setupRouterWithCaptcha(db, captcha)
	loginBody := map[string]string{
		"username": "admin", "password": "123456",
		"captcha_id": "cap1", "captcha_code": "1111",
	}
	w = doRequest(r2, "POST", "/api/v1/auth/login", loginBody, "")
	assertError(t, w)
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	body := map[string]interface{}{
		"username": "testuser",
		"password": "123456",
		"nickname": "测试用户",
		"status":   1,
	}
	w := doRequest(r, "POST", "/api/v1/users", body, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["username"] != "testuser" {
		t.Fatalf("expected testuser, got %v", data["username"])
	}
	if data["id"] == nil || data["id"].(float64) == 0 {
		t.Fatal("user id should be set")
	}
}

func TestCreateUserDuplicate(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	body := map[string]interface{}{
		"username": "admin",
		"password": "123456",
	}
	w := doRequest(r, "POST", "/api/v1/users", body, token)
	assertError(t, w)
}

func TestListUsers(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	w := doRequest(r, "GET", "/api/v1/users?page=1&page_size=10", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	total := data["total"].(float64)
	if total != 1 {
		t.Fatalf("expected 1 user, got %v", total)
	}
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	body := map[string]interface{}{
		"nickname": "新昵称",
		"phone":    "13800138000",
	}
	w := doRequest(r, "PUT", "/api/v1/users/1", body, token)
	assertSuccess(t, w)
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	// 先创建一个用户
	createBody := map[string]interface{}{"username": "toDelete", "password": "123456"}
	doRequest(r, "POST", "/api/v1/users", createBody, token)

	w := doRequest(r, "DELETE", "/api/v1/users/2", nil, token)
	assertSuccess(t, w)
}

func TestResetPassword(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	w := doRequest(r, "PUT", "/api/v1/users/1/reset-password", nil, token)
	assertSuccess(t, w)
}

func TestUpdateUserStatus(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	body := map[string]interface{}{"status": 0}
	w := doRequest(r, "PUT", "/api/v1/users/1/status", body, token)
	assertSuccess(t, w)
}

// setupRouterWithCaptcha 使用指定captcha创建路由
func setupRouterWithCaptcha(db *gorm.DB, captcha *CaptchaHandler) *gin.Engine {
	r := gin.New()
	authHandler := NewAuthHandler(db, captcha)
	r.POST("/api/v1/auth/login", authHandler.Login)
	return r
}
