package handler

import (
	"testing"
	"time"

	"github.com/cj/log-collect-ai-analytics/internal/model"
)

func TestServerInfo(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	w := doRequest(r, "GET", "/api/v1/monitor/server", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["go_version"] == nil {
		t.Fatal("go_version should be present")
	}
	if data["os"] == nil {
		t.Fatal("os should be present")
	}
	if data["cpu_num"] == nil {
		t.Fatal("cpu_num should be present")
	}
}

func TestLoginLogList(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	// 插入测试登录日志
	db.Create(&model.LoginLog{Username: "admin", IP: "127.0.0.1", Status: 1, Msg: "OK"})
	db.Create(&model.LoginLog{Username: "test", IP: "192.168.1.1", Status: 0, Msg: "密码错误"})

	w := doRequest(r, "GET", "/api/v1/loginlog?page=1&page_size=10", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 2 {
		t.Fatalf("expected 2 login logs, got %v", data["total"])
	}
}

func TestLoginLogListWithFilter(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	db.Create(&model.LoginLog{Username: "admin", IP: "127.0.0.1", Status: 1, Msg: "OK"})
	db.Create(&model.LoginLog{Username: "test", IP: "192.168.1.1", Status: 0, Msg: "错误"})

	// 按用户名过滤
	w := doRequest(r, "GET", "/api/v1/loginlog?page=1&page_size=10&username=admin", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Fatalf("expected 1 login log for admin, got %v", data["total"])
	}
}

func TestLoginLogDelete(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	db.Create(&model.LoginLog{Username: "admin", IP: "127.0.0.1", Status: 1, Msg: "OK"})

	body := map[string]interface{}{"ids": []int{1}}
	w := doRequest(r, "DELETE", "/api/v1/loginlog", body, token)
	assertSuccess(t, w)
}

func TestOperLogList(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	// 插入操作日志
	db.Create(&model.AuditLog{UserID: 1, Username: "admin", Action: "POST /api/v1/users", Resource: "/api/v1/users"})
	db.Create(&model.AuditLog{UserID: 1, Username: "admin", Action: "DELETE /api/v1/roles/1", Resource: "/api/v1/roles/1"})

	w := doRequest(r, "GET", "/api/v1/operlog?page=1&page_size=10", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 2 {
		t.Fatalf("expected 2 oper logs, got %v", data["total"])
	}
}

func TestOperLogListWithFilter(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	db.Create(&model.AuditLog{UserID: 1, Username: "admin", Action: "POST /api/v1/users", Resource: "/api/v1/users"})
	db.Create(&model.AuditLog{UserID: 2, Username: "test", Action: "GET /api/v1/roles", Resource: "/api/v1/roles"})

	w := doRequest(r, "GET", "/api/v1/operlog?page=1&page_size=10&resource=users", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Fatalf("expected 1 oper log, got %v", data["total"])
	}
}

func TestOperLogDelete(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	db.Create(&model.AuditLog{UserID: 1, Username: "admin", Action: "POST", Resource: "/test"})

	body := map[string]interface{}{"ids": []int{1}}
	w := doRequest(r, "DELETE", "/api/v1/operlog", body, token)
	assertSuccess(t, w)
}

func TestOnlineList(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	// 插入最近24小时内的登录成功记录
	db.Create(&model.LoginLog{
		Username: "admin",
		IP:       "127.0.0.1",
		Status:   1,
		Msg:      "OK",
		Browser:  "Chrome",
		OS:       "macOS",
	})

	w := doRequest(r, "GET", "/api/v1/online", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	list := data["list"].([]interface{})
	if len(list) != 1 {
		t.Fatalf("expected 1 online user, got %d", len(list))
	}
}

func TestOnlineDedup(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	// 同一用户登录2次
	db.Create(&model.LoginLog{Username: "admin", IP: "127.0.0.1", Status: 1, Msg: "OK"})
	db.Create(&model.LoginLog{Username: "admin", IP: "127.0.0.2", Status: 1, Msg: "OK"})

	w := doRequest(r, "GET", "/api/v1/online", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	total := data["total"].(float64)
	if total != 1 {
		t.Fatalf("expected 1 (deduped), got %v", total)
	}
}

func TestForceLogout(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	w := doRequest(r, "DELETE", "/api/v1/online/1", nil, token)
	assertSuccess(t, w)
}

// 确保 time 包被使用
var _ = time.Now
