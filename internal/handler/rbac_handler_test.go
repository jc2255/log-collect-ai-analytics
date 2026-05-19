package handler

import (
	"fmt"
	"testing"
)

// ======================== Role Tests ========================

func TestRoleCreate(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	body := map[string]interface{}{
		"name": "测试角色",
		"code": "test_role",
	}
	w := doRequest(r, "POST", "/api/v1/roles", body, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "测试角色" {
		t.Fatalf("expected 测试角色, got %v", data["name"])
	}
	if data["code"] != "test_role" {
		t.Fatalf("expected test_role, got %v", data["code"])
	}
	// status应该默认为1
	if data["status"].(float64) != 1 {
		t.Fatalf("expected status 1, got %v", data["status"])
	}
}

func TestRoleCreateEmptyName(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	body := map[string]interface{}{
		"name": "",
		"code": "test",
	}
	w := doRequest(r, "POST", "/api/v1/roles", body, token)
	assertError(t, w)
}

func TestRoleList(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	// 创建2个角色
	for i := 0; i < 2; i++ {
		body := map[string]interface{}{
			"name": fmt.Sprintf("角色%d", i),
			"code": fmt.Sprintf("role_%d", i),
		}
		doRequest(r, "POST", "/api/v1/roles", body, token)
	}

	w := doRequest(r, "GET", "/api/v1/roles?page=1&page_size=10", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	total := data["total"].(float64)
	if total != 2 {
		t.Fatalf("expected 2 roles, got %v", total)
	}
}

func TestRoleUpdate(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	// 创建
	createBody := map[string]interface{}{"name": "原始", "code": "orig"}
	doRequest(r, "POST", "/api/v1/roles", createBody, token)

	// 更新
	updateBody := map[string]interface{}{"name": "更新后", "code": "updated", "description": "描述"}
	w := doRequest(r, "PUT", "/api/v1/roles/1", updateBody, token)
	assertSuccess(t, w)
}

func TestRoleDelete(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	createBody := map[string]interface{}{"name": "删除", "code": "del"}
	doRequest(r, "POST", "/api/v1/roles", createBody, token)

	w := doRequest(r, "DELETE", "/api/v1/roles/1", nil, token)
	assertSuccess(t, w)

	// 验证列表为空
	w = doRequest(r, "GET", "/api/v1/roles?page=1&page_size=10", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 0 {
		t.Fatal("role should be deleted")
	}
}

func TestRoleAssignMenus(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	// 创建角色
	doRequest(r, "POST", "/api/v1/roles", map[string]interface{}{"name": "r1", "code": "r1"}, token)
	// 创建菜单
	doRequest(r, "POST", "/api/v1/menus", map[string]interface{}{"name": "首页", "path": "/", "menu_type": "C"}, token)
	doRequest(r, "POST", "/api/v1/menus", map[string]interface{}{"name": "设置", "path": "/settings", "menu_type": "C"}, token)

	// 分配菜单
	body := map[string]interface{}{"menu_ids": []int{1, 2}}
	w := doRequest(r, "PUT", "/api/v1/roles/1/menus", body, token)
	assertSuccess(t, w)
}

// ======================== Department Tests ========================

func TestDeptCreate(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	body := map[string]interface{}{
		"name":      "研发部",
		"parent_id": 0,
		"leader":    "张三",
	}
	w := doRequest(r, "POST", "/api/v1/depts", body, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "研发部" {
		t.Fatalf("expected 研发部, got %v", data["name"])
	}
	if data["status"].(float64) != 1 {
		t.Fatalf("expected status 1, got %v", data["status"])
	}
}

func TestDeptList(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	// 创建父部门
	doRequest(r, "POST", "/api/v1/depts", map[string]interface{}{"name": "总公司"}, token)
	// 创建子部门
	doRequest(r, "POST", "/api/v1/depts", map[string]interface{}{"name": "研发部", "parent_id": 1}, token)

	w := doRequest(r, "GET", "/api/v1/depts", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Fatalf("expected 1 root dept, got %d", len(data))
	}
	// 检查有子部门
	root := data[0].(map[string]interface{})
	children := root["children"].([]interface{})
	if len(children) != 1 {
		t.Fatalf("expected 1 child dept, got %d", len(children))
	}
}

func TestDeptUpdate(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	doRequest(r, "POST", "/api/v1/depts", map[string]interface{}{"name": "原始"}, token)
	w := doRequest(r, "PUT", "/api/v1/depts/1", map[string]interface{}{"name": "修改后", "leader": "李四"}, token)
	assertSuccess(t, w)
}

func TestDeptDelete(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	doRequest(r, "POST", "/api/v1/depts", map[string]interface{}{"name": "删除部门"}, token)
	w := doRequest(r, "DELETE", "/api/v1/depts/1", nil, token)
	assertSuccess(t, w)
}

// ======================== Post Tests ========================

func TestPostCreate(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	body := map[string]interface{}{
		"name": "高级开发",
		"code": "senior_dev",
		"sort": 1,
	}
	w := doRequest(r, "POST", "/api/v1/posts", body, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "高级开发" {
		t.Fatalf("expected 高级开发, got %v", data["name"])
	}
	if data["status"].(float64) != 1 {
		t.Fatalf("expected status 1, got %v", data["status"])
	}
}

func TestPostList(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	doRequest(r, "POST", "/api/v1/posts", map[string]interface{}{"name": "岗位1", "code": "p1"}, token)
	doRequest(r, "POST", "/api/v1/posts", map[string]interface{}{"name": "岗位2", "code": "p2"}, token)

	w := doRequest(r, "GET", "/api/v1/posts?page=1&page_size=10", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 2 {
		t.Fatalf("expected 2 posts, got %v", data["total"])
	}
}

func TestPostUpdate(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	doRequest(r, "POST", "/api/v1/posts", map[string]interface{}{"name": "原始", "code": "orig"}, token)
	w := doRequest(r, "PUT", "/api/v1/posts/1", map[string]interface{}{"name": "修改后", "code": "updated"}, token)
	assertSuccess(t, w)
}

func TestPostDelete(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	doRequest(r, "POST", "/api/v1/posts", map[string]interface{}{"name": "del", "code": "del"}, token)
	w := doRequest(r, "DELETE", "/api/v1/posts/1", nil, token)
	assertSuccess(t, w)
}

// ======================== Menu Tests ========================

func TestMenuCreate(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	body := map[string]interface{}{
		"name":      "系统管理",
		"path":      "/system",
		"icon":      "Setting",
		"menu_type": "M",
		"sort":      1,
	}
	w := doRequest(r, "POST", "/api/v1/menus", body, token)
	resp := assertSuccess(t, w)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "系统管理" {
		t.Fatalf("expected 系统管理, got %v", data["name"])
	}
	if data["menu_type"] != "M" {
		t.Fatalf("expected M, got %v", data["menu_type"])
	}
	if data["visible"].(float64) != 1 {
		t.Fatalf("expected visible 1, got %v", data["visible"])
	}
}

func TestMenuList(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	// 创建父菜单
	doRequest(r, "POST", "/api/v1/menus", map[string]interface{}{"name": "系统", "path": "/sys", "menu_type": "M"}, token)
	// 创建子菜单
	doRequest(r, "POST", "/api/v1/menus", map[string]interface{}{"name": "用户", "path": "/sys/user", "menu_type": "C", "parent_id": 1}, token)

	w := doRequest(r, "GET", "/api/v1/menus", nil, token)
	resp := assertSuccess(t, w)
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Fatalf("expected 1 root menu, got %d", len(data))
	}
	root := data[0].(map[string]interface{})
	children := root["children"].([]interface{})
	if len(children) != 1 {
		t.Fatalf("expected 1 child menu, got %d", len(children))
	}
}

func TestMenuUpdate(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	doRequest(r, "POST", "/api/v1/menus", map[string]interface{}{"name": "原始", "menu_type": "C"}, token)
	w := doRequest(r, "PUT", "/api/v1/menus/1", map[string]interface{}{"name": "修改后", "icon": "Edit"}, token)
	assertSuccess(t, w)
}

func TestMenuDelete(t *testing.T) {
	db := setupTestDB(t)
	admin := seedTestUser(t, db)
	r := setupAuthRouter(db)
	token := getToken(t, admin.ID, admin.Username)

	doRequest(r, "POST", "/api/v1/menus", map[string]interface{}{"name": "删除", "menu_type": "C"}, token)
	w := doRequest(r, "DELETE", "/api/v1/menus/1", nil, token)
	assertSuccess(t, w)
}
