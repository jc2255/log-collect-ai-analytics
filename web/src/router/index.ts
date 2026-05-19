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
        component: () => import('../views/dashboard/Index.vue'),
        meta: { title: '首页' },
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('../views/system/Profile.vue'),
        meta: { title: '个人中心' },
      },
      // 权限管理
      {
        path: 'permission/users',
        name: 'Users',
        component: () => import('../views/permission/Users.vue'),
        meta: { title: '用户管理', parent: '权限管理' },
      },
      {
        path: 'permission/roles',
        name: 'Roles',
        component: () => import('../views/permission/Roles.vue'),
        meta: { title: '角色管理', parent: '权限管理' },
      },
      {
        path: 'permission/menus',
        name: 'Menus',
        component: () => import('../views/permission/Menus.vue'),
        meta: { title: '菜单管理', parent: '权限管理' },
      },
      {
        path: 'permission/dept',
        name: 'Dept',
        component: () => import('../views/permission/Dept.vue'),
        meta: { title: '部门管理', parent: '权限管理' },
      },
      {
        path: 'permission/post',
        name: 'Post',
        component: () => import('../views/permission/Post.vue'),
        meta: { title: '岗位管理', parent: '权限管理' },
      },
      // 系统监控
      {
        path: 'monitor/loginlog',
        name: 'LoginLog',
        component: () => import('../views/monitor/LoginLog.vue'),
        meta: { title: '登录日志', parent: '系统监控' },
      },
      {
        path: 'monitor/operlog',
        name: 'OperLog',
        component: () => import('../views/monitor/OperLog.vue'),
        meta: { title: '操作日志', parent: '系统监控' },
      },
      {
        path: 'monitor/online',
        name: 'Online',
        component: () => import('../views/monitor/Online.vue'),
        meta: { title: '在线用户', parent: '系统监控' },
      },
      // 日志管理
      {
        path: 'log/store',
        name: 'LogStore',
        component: () => import('../views/log/Store.vue'),
        meta: { title: '日志库', parent: '日志管理' },
      },
      {
        path: 'log/list',
        name: 'LogList',
        component: () => import('../views/log/List.vue'),
        meta: { title: '日志查询', parent: '日志管理' },
      },
      // 备份管理
      {
        path: 'backup/snapshots',
        name: 'Snapshots',
        component: () => import('../views/backup/Snapshots.vue'),
        meta: { title: '备份列表', parent: '备份管理' },
      },
      {
        path: 'backup/policies',
        name: 'Policies',
        component: () => import('../views/backup/Policies.vue'),
        meta: { title: '备份策略', parent: '备份管理' },
      },
      // 日志采集
      {
        path: 'collect/tasks',
        name: 'CollectTasks',
        component: () => import('../views/collect/Tasks.vue'),
        meta: { title: '采集任务', parent: '日志采集' },
      },
      {
        path: 'collect/agents',
        name: 'CollectAgents',
        component: () => import('../views/collect/Agents.vue'),
        meta: { title: 'Agent管理', parent: '日志采集' },
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
