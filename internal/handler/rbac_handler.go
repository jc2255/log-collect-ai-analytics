package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/middleware"
	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

// RoleHandler 角色管理
type RoleHandler struct{ DB *gorm.DB }

func NewRoleHandler(db *gorm.DB) *RoleHandler { return &RoleHandler{DB: db} }

func (h *RoleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	name := c.Query("name")
	status := c.Query("status")

	query := h.DB.Model(&model.Role{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var roles []model.Role
	query.Order("sort asc, id asc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles)
	response.Success(c, gin.H{"list": roles, "total": total})
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		Sort        int    `json:"sort"`
		Status      *int8  `json:"status"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误: "+err.Error())
		return
	}
	if req.Name == "" || req.Code == "" {
		response.Error(c, response.ErrorCode, "角色名称和标识不能为空")
		return
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	role := model.Role{
		Name: req.Name, Code: req.Code, Sort: req.Sort,
		Status: status, Description: req.Description,
	}
	if err := h.DB.Create(&role).Error; err != nil {
		response.Error(c, response.ErrorCode, "创建失败: "+err.Error())
		return
	}
	response.Success(c, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误")
		return
	}
	h.DB.Model(&model.Role{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name": req.Name, "code": req.Code, "sort": req.Sort,
		"status": req.Status, "description": req.Description,
	})
	response.Success(c, nil)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.DB.Delete(&model.Role{}, id)
	response.Success(c, nil)
}

// AssignMenus 角色分配菜单权限
func (h *RoleHandler) AssignMenus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		MenuIDs []uint `json:"menu_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误")
		return
	}
	var role model.Role
	if err := h.DB.First(&role, id).Error; err != nil {
		response.Error(c, response.ErrorCode, "角色不存在")
		return
	}
	var menus []model.Menu
	h.DB.Where("id IN ?", req.MenuIDs).Find(&menus)
	h.DB.Model(&role).Association("Menus").Replace(menus)
	response.Success(c, nil)
}

// DeptHandler 部门管理
type DeptHandler struct{ DB *gorm.DB }

func NewDeptHandler(db *gorm.DB) *DeptHandler { return &DeptHandler{DB: db} }

func (h *DeptHandler) List(c *gin.Context) {
	var depts []model.Department
	h.DB.Order("sort asc, id asc").Find(&depts)
	tree := buildDeptTree(depts, 0)
	response.Success(c, tree)
}

func buildDeptTree(depts []model.Department, parentID uint) []model.Department {
	var result []model.Department
	for _, d := range depts {
		if d.ParentID == parentID {
			d.Children = buildDeptTree(depts, d.ID)
			result = append(result, d)
		}
	}
	return result
}

func (h *DeptHandler) Create(c *gin.Context) {
	var req struct {
		ParentID uint   `json:"parent_id"`
		Name     string `json:"name"`
		Sort     int    `json:"sort"`
		Leader   string `json:"leader"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Status   *int8  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误: "+err.Error())
		return
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	dept := model.Department{
		ParentID: req.ParentID, Name: req.Name, Sort: req.Sort,
		Leader: req.Leader, Phone: req.Phone, Email: req.Email, Status: status,
	}
	if err := h.DB.Create(&dept).Error; err != nil {
		response.Error(c, response.ErrorCode, "创建失败: "+err.Error())
		return
	}
	response.Success(c, dept)
}

func (h *DeptHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.Department
	c.ShouldBindJSON(&req)
	h.DB.Model(&model.Department{}).Where("id = ?", id).Updates(map[string]interface{}{
		"parent_id": req.ParentID, "name": req.Name, "sort": req.Sort,
		"leader": req.Leader, "phone": req.Phone, "email": req.Email, "status": req.Status,
	})
	response.Success(c, nil)
}

func (h *DeptHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.DB.Delete(&model.Department{}, id)
	response.Success(c, nil)
}

// PostHandler 岗位管理
type PostHandler struct{ DB *gorm.DB }

func NewPostHandler(db *gorm.DB) *PostHandler { return &PostHandler{DB: db} }

func (h *PostHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var total int64
	h.DB.Model(&model.Post{}).Count(&total)

	var posts []model.Post
	h.DB.Order("sort asc, id asc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&posts)
	response.Success(c, gin.H{"list": posts, "total": total})
}

func (h *PostHandler) Create(c *gin.Context) {
	var req struct {
		Name   string `json:"name"`
		Code   string `json:"code"`
		Sort   int    `json:"sort"`
		Status *int8  `json:"status"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误: "+err.Error())
		return
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	post := model.Post{
		Name: req.Name, Code: req.Code, Sort: req.Sort, Status: status, Remark: req.Remark,
	}
	if err := h.DB.Create(&post).Error; err != nil {
		response.Error(c, response.ErrorCode, "创建失败: "+err.Error())
		return
	}
	response.Success(c, post)
}

func (h *PostHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.Post
	c.ShouldBindJSON(&req)
	h.DB.Model(&model.Post{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name": req.Name, "code": req.Code, "sort": req.Sort, "status": req.Status, "remark": req.Remark,
	})
	response.Success(c, nil)
}

func (h *PostHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.DB.Delete(&model.Post{}, id)
	response.Success(c, nil)
}

// MenuHandler 菜单管理
type MenuHandler struct{ DB *gorm.DB }

func NewMenuHandler(db *gorm.DB) *MenuHandler { return &MenuHandler{DB: db} }

func (h *MenuHandler) List(c *gin.Context) {
	var menus []model.Menu
	h.DB.Order("sort asc, id asc").Find(&menus)
	tree := buildMenuTree(menus, 0)
	response.Success(c, tree)
}

func buildMenuTree(menus []model.Menu, parentID uint) []model.Menu {
	var result []model.Menu
	for _, m := range menus {
		if m.ParentID == parentID {
			m.Children = buildMenuTree(menus, m.ID)
			result = append(result, m)
		}
	}
	return result
}

// UserMenus 获取当前用户有权限的菜单树
func (h *MenuHandler) UserMenus(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	// 查询用户关联的角色ID
	var roleIDs []uint
	h.DB.Table("user_roles").Where("user_id = ?", userID).Pluck("role_id", &roleIDs)
	if len(roleIDs) == 0 {
		response.Success(c, []model.Menu{})
		return
	}

	// 查询角色关联的菜单ID（去重）
	var menuIDs []uint
	h.DB.Table("role_menus").Where("role_id IN ?", roleIDs).Distinct().Pluck("menu_id", &menuIDs)
	if len(menuIDs) == 0 {
		response.Success(c, []model.Menu{})
		return
	}

	// 查询菜单（只返回可见且启用的）
	var menus []model.Menu
	h.DB.Where("id IN ? AND visible = 1 AND status = 1", menuIDs).Order("sort asc, id asc").Find(&menus)

	// 需要包含父级菜单以构建树
	var allMenuIDs = make(map[uint]bool)
	for _, m := range menus {
		allMenuIDs[m.ID] = true
		// 向上追溯父级
		pid := m.ParentID
		for pid > 0 {
			if allMenuIDs[pid] {
				break
			}
			allMenuIDs[pid] = true
			var parent model.Menu
			if err := h.DB.Where("id = ?", pid).First(&parent).Error; err != nil {
				break
			}
			menus = append(menus, parent)
			pid = parent.ParentID
		}
	}

	// 去重并排序
	uniqueMenus := make([]model.Menu, 0, len(allMenuIDs))
	seen := make(map[uint]bool)
	for _, m := range menus {
		if !seen[m.ID] && allMenuIDs[m.ID] {
			seen[m.ID] = true
			uniqueMenus = append(uniqueMenus, m)
		}
	}

	// 按 sort 和 id 排序
	for i := 0; i < len(uniqueMenus); i++ {
		for j := i + 1; j < len(uniqueMenus); j++ {
			if uniqueMenus[i].Sort > uniqueMenus[j].Sort ||
				(uniqueMenus[i].Sort == uniqueMenus[j].Sort && uniqueMenus[i].ID > uniqueMenus[j].ID) {
				uniqueMenus[i], uniqueMenus[j] = uniqueMenus[j], uniqueMenus[i]
			}
		}
	}

	tree := buildMenuTree(uniqueMenus, 0)
	response.Success(c, tree)
}

func (h *MenuHandler) Create(c *gin.Context) {
	var req struct {
		ParentID  uint   `json:"parent_id"`
		Name      string `json:"name"`
		Path      string `json:"path"`
		Component string `json:"component"`
		Icon      string `json:"icon"`
		Sort      int    `json:"sort"`
		MenuType  string `json:"menu_type"`
		Visible   *int8  `json:"visible"`
		Status    *int8  `json:"status"`
		Perms     string `json:"perms"`
		ApiPath   string `json:"api_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误: "+err.Error())
		return
	}
	visible := int8(1)
	if req.Visible != nil {
		visible = *req.Visible
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	menuType := req.MenuType
	if menuType == "" {
		menuType = "C"
	}
	menu := model.Menu{
		ParentID: req.ParentID, Name: req.Name, Path: req.Path,
		Component: req.Component, Icon: req.Icon, Sort: req.Sort,
		MenuType: menuType, Visible: visible, Status: status,
		Perms: req.Perms, ApiPath: req.ApiPath,
	}
	if err := h.DB.Create(&menu).Error; err != nil {
		response.Error(c, response.ErrorCode, "创建失败: "+err.Error())
		return
	}
	response.Success(c, menu)
}

func (h *MenuHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.Menu
	c.ShouldBindJSON(&req)
	h.DB.Model(&model.Menu{}).Where("id = ?", id).Updates(map[string]interface{}{
		"parent_id": req.ParentID, "name": req.Name, "path": req.Path,
		"component": req.Component, "icon": req.Icon, "sort": req.Sort,
		"menu_type": req.MenuType, "visible": req.Visible, "status": req.Status,
		"perms": req.Perms, "api_path": req.ApiPath,
	})
	response.Success(c, nil)
}

func (h *MenuHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.DB.Delete(&model.Menu{}, id)
	response.Success(c, nil)
}
