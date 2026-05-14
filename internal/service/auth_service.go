package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/cj/log-collect-ai-analytics/internal/dao"
	"github.com/cj/log-collect-ai-analytics/internal/middleware"
	"github.com/cj/log-collect-ai-analytics/internal/model"
)

// AuthService 认证服务
type AuthService struct {
	userDAO *dao.UserDAO
}

func NewAuthService(userDAO *dao.UserDAO) *AuthService {
	return &AuthService{userDAO: userDAO}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string    `json:"token"`
	UserInfo *UserInfo `json:"user_info"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID           uint       `json:"id"`
	TenantID     uint       `json:"tenant_id"`
	TenantName   string     `json:"tenant_name"`
	Username     string     `json:"username"`
	Nickname     string     `json:"nickname"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Avatar       string     `json:"avatar"`
	IsSuperAdmin bool       `json:"is_super_admin"`
	Roles        []RoleInfo `json:"roles"`
}

// RoleInfo 角色信息
type RoleInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// Login 登录
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userDAO.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("用户已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成Token
	token, err := middleware.GenerateToken(user.ID, user.TenantID, user.Username, user.IsSuperAdmin)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	// 组装用户信息
	userInfo := &UserInfo{
		ID:           user.ID,
		TenantID:     user.TenantID,
		TenantName:   user.Tenant.Name,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Email:        user.Email,
		Phone:        user.Phone,
		Avatar:       user.Avatar,
		IsSuperAdmin: user.IsSuperAdmin,
	}
	for _, role := range user.Roles {
		userInfo.Roles = append(userInfo.Roles, RoleInfo{
			ID:   role.ID,
			Name: role.Name,
			Code: role.Code,
		})
	}

	return &LoginResponse{Token: token, UserInfo: userInfo}, nil
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(userID uint) (*UserInfo, error) {
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	userInfo := &UserInfo{
		ID:           user.ID,
		TenantID:     user.TenantID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Email:        user.Email,
		Phone:        user.Phone,
		Avatar:       user.Avatar,
		IsSuperAdmin: user.IsSuperAdmin,
	}
	for _, role := range user.Roles {
		userInfo.Roles = append(userInfo.Roles, RoleInfo{
			ID:   role.ID,
			Name: role.Name,
			Code: role.Code,
		})
	}
	return userInfo, nil
}

// ChangePassword 修改密码
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (s *AuthService) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	user.PasswordHash = string(hash)
	return s.userDAO.Update(user)
}

// CreateUser 创建用户
type CreateUserRequest struct {
	TenantID uint   `json:"tenant_id" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

func (s *AuthService) CreateUser(req *CreateUserRequest) (*model.User, error) {
	// 检查用户名是否已存在
	existing, _ := s.userDAO.GetByUsername(req.TenantID, req.Username)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("用户名已存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &model.User{
		TenantID:     req.TenantID,
		Username:     req.Username,
		PasswordHash: string(hash),
		Nickname:     req.Nickname,
		Email:        req.Email,
		Phone:        req.Phone,
		Status:       1,
	}

	if err := s.userDAO.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// ListUsers 用户列表
func (s *AuthService) ListUsers(tenantID uint, page, pageSize int) ([]model.User, int64, error) {
	return s.userDAO.List(tenantID, page, pageSize)
}
