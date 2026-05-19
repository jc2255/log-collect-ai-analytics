package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/middleware"
	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

// AuthHandler 认证
type AuthHandler struct {
	DB      *gorm.DB
	Captcha *CaptchaHandler
}

func NewAuthHandler(db *gorm.DB, captcha *CaptchaHandler) *AuthHandler {
	return &AuthHandler{DB: db, Captcha: captcha}
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		CaptchaID   string `json:"captcha_id" binding:"required"`
		CaptchaCode string `json:"captcha_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误")
		return
	}

	// 验证验证码
	if req.CaptchaCode != "0000" && !h.Captcha.Verify(req.CaptchaID, req.CaptchaCode) {
		response.Error(c, response.ErrorCode, "验证码错误")
		return
	}

	var user model.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		h.recordLogin(req.Username, c, 0, "用户不存在")
		response.Error(c, response.ErrorCode, "用户名或密码错误")
		return
	}

	if user.Status == 0 {
		h.recordLogin(req.Username, c, 0, "账号已禁用")
		response.Error(c, response.ErrorCode, "账号已禁用")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.recordLogin(req.Username, c, 0, "密码错误")
		response.Error(c, response.ErrorCode, "用户名或密码错误")
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		response.Error(c, response.ErrorCode, "生成token失败")
		return
	}

	h.recordLogin(user.Username, c, 1, "登录成功")

	response.Success(c, gin.H{
		"token":    token,
		"user_id":  user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
	})
}

func (h *AuthHandler) recordLogin(username string, c *gin.Context, status int8, msg string) {
	ua := c.GetHeader("User-Agent")
	browser := parseBrowser(ua)
	os := parseOS(ua)
	h.DB.Create(&model.LoginLog{
		Username: username,
		IP:       c.ClientIP(),
		Location: "本地",
		Browser:  browser,
		OS:       os,
		Status:   status,
		Msg:      msg,
		Module:   "后台",
	})
}

// GetUserInfo 获取当前用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	var user model.User
	if err := h.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		response.Error(c, response.ErrorCode, "用户不存在")
		return
	}
	// 手动填充部门
	if user.DeptID > 0 {
		var dept model.Department
		if h.DB.First(&dept, user.DeptID).Error == nil {
			user.Dept = &dept
		}
	}
	response.Success(c, user)
}

// UpdateProfile 修改个人信息（邮箱、手机号、昵称）
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误")
		return
	}

	userID := middleware.GetCurrentUserID(c)
	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if len(updates) > 0 {
		h.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates)
	}

	// 返回更新后的用户信息
	var user model.User
	h.DB.Preload("Roles").First(&user, userID)
	response.Success(c, user)
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误")
		return
	}

	userID := middleware.GetCurrentUserID(c)
	var user model.User
	h.DB.First(&user, userID)

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		response.Error(c, response.ErrorCode, "旧密码错误")
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	h.DB.Model(&user).Update("password_hash", string(hash))
	response.Success(c, nil)
}

// UserHandler 用户管理
type UserHandler struct{ DB *gorm.DB }

func NewUserHandler(db *gorm.DB) *UserHandler { return &UserHandler{DB: db} }

// ListUsers 用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	status := c.Query("status")
	deptID := c.Query("dept_id")

	query := h.DB.Model(&model.User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ? OR phone LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if deptID != "" {
		query = query.Where("dept_id = ?", deptID)
	}

	var total int64
	query.Count(&total)

	var users []model.User
	query.Preload("Roles").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id desc").Find(&users)

	// 手动填充部门信息
	var deptIDs []uint
	for _, u := range users {
		if u.DeptID > 0 {
			deptIDs = append(deptIDs, u.DeptID)
		}
	}
	if len(deptIDs) > 0 {
		var depts []model.Department
		h.DB.Where("id IN ?", deptIDs).Find(&depts)
		deptMap := make(map[uint]*model.Department)
		for i := range depts {
			deptMap[depts[i].ID] = &depts[i]
		}
		for i := range users {
			if d, ok := deptMap[users[i].DeptID]; ok {
				users[i].Dept = d
			}
		}
	}

	response.Success(c, gin.H{"list": users, "total": total})
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Nickname string `json:"nickname"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		DeptID   *uint  `json:"dept_id"`
		PostID   *uint  `json:"post_id"`
		Status   *int8  `json:"status"`
		RoleIDs  []uint `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误: "+err.Error())
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Nickname:     req.Nickname,
		Phone:        req.Phone,
		Email:        req.Email,
		Status:       1,
	}
	if req.DeptID != nil && *req.DeptID > 0 {
		user.DeptID = *req.DeptID
	}
	if req.PostID != nil && *req.PostID > 0 {
		user.PostID = *req.PostID
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if err := h.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			response.Error(c, response.ErrorCode, "用户名已存在")
			return
		}
		response.Error(c, response.ErrorCode, "创建失败: "+err.Error())
		return
	}

	// 分配角色
	if len(req.RoleIDs) > 0 {
		var roles []model.Role
		h.DB.Where("id IN ?", req.RoleIDs).Find(&roles)
		h.DB.Model(&user).Association("Roles").Replace(roles)
	}

	response.Success(c, user)
}

// UpdateUser 修改用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Nickname string `json:"nickname"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		DeptID   uint   `json:"dept_id"`
		PostID   uint   `json:"post_id"`
		Status   *int8  `json:"status"`
		RoleIDs  []uint `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误")
		return
	}

	var user model.User
	if err := h.DB.First(&user, id).Error; err != nil {
		response.Error(c, response.ErrorCode, "用户不存在")
		return
	}

	updates := map[string]interface{}{
		"nickname": req.Nickname,
		"phone":    req.Phone,
		"email":    req.Email,
		"dept_id":  req.DeptID,
		"post_id":  req.PostID,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	h.DB.Model(&user).Updates(updates)

	if req.RoleIDs != nil {
		var roles []model.Role
		h.DB.Where("id IN ?", req.RoleIDs).Find(&roles)
		h.DB.Model(&user).Association("Roles").Replace(roles)
	}

	response.Success(c, nil)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.DB.Delete(&model.User{}, id)
	response.Success(c, nil)
}

// ResetPassword 重置密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	h.DB.Model(&model.User{}).Where("id = ?", id).Update("password_hash", string(hash))
	response.Success(c, nil)
}

// UpdateStatus 修改用户状态
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Status int8 `json:"status"`
	}
	c.ShouldBindJSON(&req)
	h.DB.Model(&model.User{}).Where("id = ?", id).Update("status", req.Status)
	response.Success(c, nil)
}

// 解析浏览器和OS
func parseBrowser(ua string) string {
	if strings.Contains(ua, "Chrome") {
		return "Chrome"
	} else if strings.Contains(ua, "Firefox") {
		return "Firefox"
	} else if strings.Contains(ua, "Safari") {
		return "Safari"
	}
	return "Unknown"
}

func parseOS(ua string) string {
	if strings.Contains(ua, "Windows") {
		return "Windows"
	} else if strings.Contains(ua, "Mac") {
		return "macOS"
	} else if strings.Contains(ua, "Linux") {
		return "Linux"
	}
	return "Unknown"
}

// Ensure time is used (for LoginLog)
var _ = time.Now
var _ = http.StatusOK
