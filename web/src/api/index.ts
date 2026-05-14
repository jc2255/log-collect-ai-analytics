import request from './request'

// 认证相关API
export const authApi = {
  login(data: { username: string; password: string }) {
    return request.post('/auth/login', data)
  },
  getUserInfo() {
    return request.get('/auth/userinfo')
  },
  changePassword(data: { old_password: string; new_password: string }) {
    return request.put('/auth/password', data)
  },
}

// 用户管理API
export const userApi = {
  list(params: { page: number; page_size: number; tenant_id?: number }) {
    return request.get('/users', { params })
  },
  create(data: any) {
    return request.post('/users', data)
  },
  assignRoles(id: number, role_ids: number[]) {
    return request.put(`/users/${id}/roles`, { role_ids })
  },
}

// 角色管理API
export const roleApi = {
  list(params: { page: number; page_size: number }) {
    return request.get('/roles', { params })
  },
  get(id: number) {
    return request.get(`/roles/${id}`)
  },
  create(data: any) {
    return request.post('/roles', data)
  },
  update(id: number, data: any) {
    return request.put(`/roles/${id}`, data)
  },
  delete(id: number) {
    return request.delete(`/roles/${id}`)
  },
  assignPermissions(id: number, permission_ids: number[]) {
    return request.put(`/roles/${id}/permissions`, { permission_ids })
  },
}

// 租户管理API
export const tenantApi = {
  list(params: { page: number; page_size: number }) {
    return request.get('/tenants', { params })
  },
  get(id: number) {
    return request.get(`/tenants/${id}`)
  },
  create(data: any) {
    return request.post('/tenants', data)
  },
  update(id: number, data: any) {
    return request.put(`/tenants/${id}`, data)
  },
  delete(id: number) {
    return request.delete(`/tenants/${id}`)
  },
}

// 菜单管理API
export const menuApi = {
  list() {
    return request.get('/menus')
  },
  create(data: any) {
    return request.post('/menus', data)
  },
  update(id: number, data: any) {
    return request.put(`/menus/${id}`, data)
  },
  delete(id: number) {
    return request.delete(`/menus/${id}`)
  },
}
