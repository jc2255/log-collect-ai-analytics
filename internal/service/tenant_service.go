package service

import (
	"errors"

	"github.com/cj/log-collect-ai-analytics/internal/dao"
	"github.com/cj/log-collect-ai-analytics/internal/model"
)

// TenantService 租户管理服务
type TenantService struct {
	tenantDAO *dao.TenantDAO
}

func NewTenantService(tenantDAO *dao.TenantDAO) *TenantService {
	return &TenantService{tenantDAO: tenantDAO}
}

type CreateTenantRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	QuotaConfig string `json:"quota_config"`
	Description string `json:"description"`
}

func (s *TenantService) Create(req *CreateTenantRequest) (*model.Tenant, error) {
	// 检查code唯一
	existing, _ := s.tenantDAO.GetByCode(req.Code)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("租户编码已存在")
	}

	tenant := &model.Tenant{
		Name:        req.Name,
		Code:        req.Code,
		QuotaConfig: req.QuotaConfig,
		Description: req.Description,
		Status:      1,
	}
	err := s.tenantDAO.Create(tenant)
	return tenant, err
}

func (s *TenantService) GetByID(id uint) (*model.Tenant, error) {
	return s.tenantDAO.GetByID(id)
}

type UpdateTenantRequest struct {
	Name        string `json:"name"`
	QuotaConfig string `json:"quota_config"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

func (s *TenantService) Update(id uint, req *UpdateTenantRequest) error {
	tenant, err := s.tenantDAO.GetByID(id)
	if err != nil {
		return err
	}
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.QuotaConfig != "" {
		tenant.QuotaConfig = req.QuotaConfig
	}
	if req.Description != "" {
		tenant.Description = req.Description
	}
	if req.Status != 0 {
		tenant.Status = req.Status
	}
	return s.tenantDAO.Update(tenant)
}

func (s *TenantService) Delete(id uint) error {
	return s.tenantDAO.Delete(id)
}

func (s *TenantService) List(page, pageSize int) ([]model.Tenant, int64, error) {
	return s.tenantDAO.List(page, pageSize)
}
