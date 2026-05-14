import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '../store/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/system/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/system/Dashboard.vue'),
        meta: { title: '仪表盘' },
      },
      // 系统管理
      {
        path: 'system/users',
        name: 'Users',
        component: () => import('../views/system/Users.vue'),
        meta: { title: '用户管理' },
      },
      {
        path: 'system/roles',
        name: 'Roles',
        component: () => import('../views/system/Roles.vue'),
        meta: { title: '角色管理' },
      },
      {
        path: 'system/tenants',
        name: 'Tenants',
        component: () => import('../views/system/Tenants.vue'),
        meta: { title: '租户管理' },
      },
      {
        path: 'system/menus',
        name: 'Menus',
        component: () => import('../views/system/Menus.vue'),
        meta: { title: '菜单管理' },
      },
      // 日志库管理
      {
        path: 'logstore',
        name: 'LogStore',
        component: () => import('../views/logstore/Index.vue'),
        meta: { title: '日志库管理' },
      },
      // 日志查询
      {
        path: 'search',
        name: 'LogSearch',
        component: () => import('../views/search/Index.vue'),
        meta: { title: '日志查询' },
      },
      // 聚合分析
      {
        path: 'analysis',
        name: 'Analysis',
        component: () => import('../views/analysis/Index.vue'),
        meta: { title: '聚合分析' },
      },
      // 告警管理
      {
        path: 'alert',
        name: 'Alert',
        component: () => import('../views/alert/Index.vue'),
        meta: { title: '告警管理' },
      },
      // 生命周期管理
      {
        path: 'lifecycle',
        name: 'Lifecycle',
        component: () => import('../views/lifecycle/Index.vue'),
        meta: { title: '生命周期管理' },
      },
      // Agent管理
      {
        path: 'agent',
        name: 'Agent',
        component: () => import('../views/agent/Index.vue'),
        meta: { title: 'Agent管理' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth === false) {
    next()
    return
  }

  if (!userStore.token) {
    next('/login')
    return
  }

  if (!userStore.userInfo) {
    try {
      await userStore.fetchUserInfo()
    } catch {
      userStore.logout()
      next('/login')
      return
    }
  }

  next()
})

export default router
