<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '220px'" class="aside">
      <div class="logo">
        <div class="logo-icon">L</div>
        <transition name="fade">
          <span v-if="!isCollapse" class="logo-text">LCA</span>
        </transition>
      </div>
      <div class="menu-wrapper">
        <el-menu
          :default-active="route.path"
          :collapse="isCollapse"
          router
          background-color="transparent"
          text-color="var(--tech-text-secondary)"
          active-text-color="var(--tech-primary)"
        >
          <el-menu-item index="/dashboard">
            <el-icon><Odometer /></el-icon>
            <template #title>首页</template>
          </el-menu-item>

          <el-sub-menu index="permission">
            <template #title>
              <el-icon><Key /></el-icon>
              <span>权限管理</span>
            </template>
            <el-menu-item index="/permission/users">用户管理</el-menu-item>
            <el-menu-item index="/permission/roles">角色管理</el-menu-item>
            <el-menu-item index="/permission/menus">菜单管理</el-menu-item>
            <el-menu-item index="/permission/dept">部门管理</el-menu-item>
            <el-menu-item index="/permission/post">岗位管理</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="monitor">
            <template #title>
              <el-icon><Monitor /></el-icon>
              <span>系统监控</span>
            </template>
            <el-menu-item index="/monitor/loginlog">登录日志</el-menu-item>
            <el-menu-item index="/monitor/operlog">操作日志</el-menu-item>
            <el-menu-item index="/monitor/online">在线用户</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="log">
            <template #title>
              <el-icon><Document /></el-icon>
              <span>日志管理</span>
            </template>
            <el-menu-item index="/log/store">日志库</el-menu-item>
            <el-menu-item index="/log/list">日志查询</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="backup">
            <template #title>
              <el-icon><FolderOpened /></el-icon>
              <span>备份管理</span>
            </template>
            <el-menu-item index="/backup/snapshots">备份列表</el-menu-item>
            <el-menu-item index="/backup/policies">备份策略</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="collect">
            <template #title>
              <el-icon><Connection /></el-icon>
              <span>日志采集</span>
            </template>
            <el-menu-item index="/collect/tasks">采集任务</el-menu-item>
            <el-menu-item index="/collect/agents">Agent管理</el-menu-item>
          </el-sub-menu>
        </el-menu>
      </div>
    </el-aside>

    <!-- 主内容 -->
    <el-container class="main-container">
      <el-header class="header">
        <div class="header-left">
          <div class="collapse-btn" @click="isCollapse = !isCollapse">
            <el-icon :size="18">
              <Fold v-if="!isCollapse" />
              <Expand v-else />
            </el-icon>
          </div>
          <div class="breadcrumb-wrapper">
            <el-breadcrumb separator="/">
              <el-breadcrumb-item v-if="route.meta.parent">{{ route.meta.parent }}</el-breadcrumb-item>
              <el-breadcrumb-item>{{ route.meta.title || '首页' }}</el-breadcrumb-item>
            </el-breadcrumb>
          </div>
        </div>
        <div class="header-right">
          <el-dropdown trigger="click">
            <div class="user-info">
              <div class="user-avatar">
                {{ userStore.userInfo?.nickname?.[0] || 'U' }}
              </div>
              <span class="username">{{ userStore.userInfo?.nickname || userStore.userInfo?.username }}</span>
              <el-icon :size="12"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push('/profile')">
                  <el-icon><User /></el-icon>
                  个人中心
                </el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>

    <!-- 授权码绑定弹窗 -->
    <LicenseDialog v-if="!licenseActivated" :machine-id="machineId" @activated="licenseActivated = true" />
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../store/user'
import { licenseApi } from '../api'
import LicenseDialog from '../components/LicenseDialog.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const isCollapse = ref(false)
const licenseActivated = ref(true) // 默认true，检测后更新
const machineId = ref('')

// 检查授权状态
onMounted(async () => {
  try {
    const res: any = await licenseApi.getStatus()
    licenseActivated.value = res.data.activated === true
    machineId.value = res.data.machine_id || ''
  } catch {
    licenseActivated.value = false
  }
})

function handleLogout() {
  userStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  overflow: hidden;
}

/* ========== 侧边栏 ========== */
.aside {
  background: var(--tech-gradient-sidebar);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  border-right: 1px solid var(--tech-border);
  display: flex;
  flex-direction: column;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 0 16px;
  border-bottom: 1px solid var(--tech-border);
  flex-shrink: 0;
}
.logo-icon {
  width: 32px;
  height: 32px;
  color: #0a1628;
  background: var(--tech-gradient-primary);
  border-radius: var(--tech-radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 800;
  flex-shrink: 0;
}
.logo-text {
  font-size: 20px;
  font-weight: 800;
  background: var(--tech-gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: 3px;
  white-space: nowrap;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 菜单区域 */
.menu-wrapper {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px 0;
}

/* Element Plus 菜单覆盖 */
.menu-wrapper :deep(.el-menu) {
  border-right: none;
  background-color: transparent;
}
.menu-wrapper :deep(.el-menu-item),
.menu-wrapper :deep(.el-sub-menu__title) {
  height: 44px;
  line-height: 44px;
  margin: 2px 8px;
  border-radius: var(--tech-radius-sm);
  transition: all 0.2s ease;
  color: var(--tech-text-secondary);
}
.menu-wrapper :deep(.el-menu-item:hover),
.menu-wrapper :deep(.el-sub-menu__title:hover) {
  background-color: var(--tech-bg-hover) !important;
  color: var(--tech-text-primary);
}
.menu-wrapper :deep(.el-menu-item.is-active) {
  background: var(--tech-primary-bg) !important;
  color: var(--tech-primary) !important;
  position: relative;
}
.menu-wrapper :deep(.el-menu-item.is-active)::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: var(--tech-primary);
  border-radius: 0 3px 3px 0;
  box-shadow: 0 0 8px var(--tech-primary-glow);
}
.menu-wrapper :deep(.el-sub-menu .el-menu-item) {
  height: 40px;
  line-height: 40px;
  padding-left: 52px !important;
  font-size: 13px;
  margin: 1px 8px;
}
.menu-wrapper :deep(.el-sub-menu__icon-arrow) {
  color: var(--tech-text-placeholder);
}
.menu-wrapper :deep(.el-menu-item .el-icon),
.menu-wrapper :deep(.el-sub-menu__title .el-icon) {
  color: inherit;
  font-size: 18px;
}

/* ========== 主容器 ========== */
.main-container {
  background-color: var(--tech-bg-page);
  display: flex;
  flex-direction: column;
}

/* ========== 顶部导航 ========== */
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 64px;
  background: var(--tech-bg-card);
  border-bottom: 1px solid var(--tech-border);
  flex-shrink: 0;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}
.collapse-btn {
  cursor: pointer;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--tech-radius-sm);
  color: var(--tech-text-regular);
  transition: all 0.2s ease;
}
.collapse-btn:hover {
  background-color: var(--tech-bg-hover);
  color: var(--tech-primary);
}
.breadcrumb-wrapper {
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: var(--tech-radius-md);
  transition: all 0.2s ease;
}
.user-info:hover {
  background-color: var(--tech-bg-hover);
}
.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--tech-gradient-primary);
  color: #0a1628;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
}
.username {
  font-size: 14px;
  color: var(--tech-text-primary);
  font-weight: 500;
}

/* ========== 主内容 ========== */
.main {
  background-color: var(--tech-bg-page);
  padding: 20px;
  overflow-y: auto;
}
</style>
