import request from './request'

// 认证相关API
export const authApi = {
  login(data: { username: string; password: string }) {
    return request.post('/auth/login', data)
  },
  getUserInfo() {
    return request.get('/auth/userinfo')
  },
  updateProfile(data: { nickname?: string; email?: string; phone?: string }) {
    return request.put('/auth/profile', data)
  },
  changePassword(data: { old_password: string; new_password: string }) {
    return request.put('/auth/password', data)
  },
}

// 用户管理API
export const userApi = {
  list(params?: any) {
    return request.get('/users', { params })
  },
  create(data: any) {
    return request.post('/users', data)
  },
  update(id: number, data: any) {
    return request.put(`/users/${id}`, data)
  },
  delete(id: number) {
    return request.delete(`/users/${id}`)
  },
  resetPassword(id: number, data: { password: string }) {
    return request.put(`/users/${id}/reset-password`, data)
  },
  updateStatus(id: number, data: { status: number }) {
    return request.put(`/users/${id}/status`, data)
  },
}

// 角色管理API
export const roleApi = {
  list(params?: any) {
    return request.get('/roles', { params })
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
  assignMenus(id: number, menu_ids: number[]) {
    return request.put(`/roles/${id}/menus`, { menu_ids })
  },
}

// 部门管理API
export const deptApi = {
  list() {
    return request.get('/depts')
  },
  create(data: any) {
    return request.post('/depts', data)
  },
  update(id: number, data: any) {
    return request.put(`/depts/${id}`, data)
  },
  delete(id: number) {
    return request.delete(`/depts/${id}`)
  },
}

// 岗位管理API
export const postApi = {
  list(params?: any) {
    return request.get('/posts', { params })
  },
  create(data: any) {
    return request.post('/posts', data)
  },
  update(id: number, data: any) {
    return request.put(`/posts/${id}`, data)
  },
  delete(id: number) {
    return request.delete(`/posts/${id}`)
  },
}

// 菜单管理API
export const menuApi = {
  list() {
    return request.get('/menus')
  },
  userMenus() {
    return request.get('/menus/user')
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

// 系统监控API
export const monitorApi = {
  getServerInfo() {
    return request.get('/monitor/server')
  },
}

// 登录日志API
export const loginLogApi = {
  list(params?: any) {
    return request.get('/loginlog', { params })
  },
  delete(ids: number[]) {
    return request.delete('/loginlog', { data: { ids } })
  },
  clean() {
    return request.delete('/loginlog/clean')
  },
}

// 操作日志API
export const operLogApi = {
  list(params?: any) {
    return request.get('/operlog', { params })
  },
  delete(ids: number[]) {
    return request.delete('/operlog', { data: { ids } })
  },
}

// 在线用户API
export const onlineApi = {
  list() {
    return request.get('/online')
  },
  forceLogout(id: string) {
    return request.delete(`/online/${id}`)
  },
}

// 日志库API
export const logStoreApi = {
  list(params?: any) {
    return request.get('/logstore', { params })
  },
  create(data: any) {
    return request.post('/logstore', data)
  },
  update(id: number, data: any) {
    return request.put(`/logstore/${id}`, data)
  },
  delete(id: number) {
    return request.delete(`/logstore/${id}`)
  },
}

// ES日志查询API
export const esLogApi = {
  search(params: any) {
    return request.get('/eslog', { params })
  },
  fields(params: any) {
    return request.get('/eslog/fields', { params })
  },
  histogram(params: any) {
    return request.get('/eslog/histogram', { params })
  },
}

// AI智能告警API
export const aiAlertApi = {
  toggle(storeId: number, enabled: boolean) {
    return request.put(`/logstore/${storeId}/ai-alert`, { enabled })
  },
  getConfig(storeId: number) {
    return request.get(`/logstore/${storeId}/ai-alert/config`)
  },
  updateConfig(storeId: number, config: string) {
    return request.put(`/logstore/${storeId}/ai-alert/config`, { config })
  },
  test(storeId: number) {
    return request.post(`/logstore/${storeId}/ai-alert/test`)
  },
  history(params?: any) {
    return request.get('/ai-alert/history', { params })
  },
}

// 首页统计API
export const dashboardApi = {
  getStats() {
    return request.get('/dashboard')
  },
}

// 备份快照API
export const backupApi = {
  listSnapshots(params?: any) {
    return request.get('/backup/snapshots', { params })
  },
  deleteSnapshot(name: string, repo?: string) {
    return request.delete(`/backup/snapshots/${name}`, { params: repo ? { repo } : {} })
  },
  restoreSnapshot(name: string, repo?: string) {
    return request.post(`/backup/snapshots/${name}/restore`, {}, { params: repo ? { repo } : {} })
  },
}

// SLM策略API
export const slmApi = {
  list() {
    return request.get('/backup/policies')
  },
  create(data: any) {
    return request.post('/backup/policies', data)
  },
  update(id: number, data: any) {
    return request.put(`/backup/policies/${id}`, data)
  },
  delete(id: number) {
    return request.delete(`/backup/policies/${id}`)
  },
  execute(id: number) {
    return request.post(`/backup/policies/${id}/execute`)
  },
}

// 采集任务API
export const collectApi = {
  listTasks() {
    return request.get('/collect/tasks')
  },
  createTask(data: any) {
    return request.post('/collect/tasks', data)
  },
  updateTask(id: number, data: any) {
    return request.put(`/collect/tasks/${id}`, data)
  },
  deleteTask(id: number) {
    return request.delete(`/collect/tasks/${id}`)
  },
  listAgents() {
    return request.get('/agents')
  },
  deleteAgent(id: number) {
    return request.delete(`/agents/${id}`)
  },
}

// 授权码API
export const licenseApi = {
  getStatus() {
    return request.get('/license/status')
  },
  activate(licenseKey: string) {
    return request.post('/license/activate', { license_key: licenseKey })
  },
  deactivate() {
    return request.delete('/license')
  },
}
