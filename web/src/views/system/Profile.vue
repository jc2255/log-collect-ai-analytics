<template>
  <div class="profile-page">
    <div class="profile-card">
      <div class="card-header">
        <el-icon :size="22" color="var(--tech-primary)"><User /></el-icon>
        <span>个人中心</span>
      </div>

      <!-- 用户头像区 -->
      <div class="profile-banner">
        <div class="avatar-large">{{ userStore.userInfo?.nickname?.[0] || 'U' }}</div>
        <div class="banner-info">
          <div class="banner-name">{{ userStore.userInfo?.nickname || '-' }}</div>
          <div class="banner-role">{{ roleText }}</div>
        </div>
      </div>

      <el-tabs v-model="activeTab" class="profile-tabs">
        <!-- 基本信息 Tab -->
        <el-tab-pane label="基本信息" name="info">
          <el-form
            ref="infoFormRef"
            :model="infoForm"
            :rules="infoRules"
            label-width="80px"
            class="profile-form"
          >
            <el-form-item label="用户名">
              <el-input :model-value="userStore.userInfo?.username" disabled />
            </el-form-item>
            <el-form-item label="昵称" prop="nickname">
              <el-input v-model="infoForm.nickname" placeholder="请输入昵称" />
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="infoForm.email" placeholder="请输入邮箱" />
            </el-form-item>
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="infoForm.phone" placeholder="请输入手机号" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleUpdateInfo" :loading="infoLoading">保存修改</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 修改密码 Tab -->
        <el-tab-pane label="修改密码" name="password">
          <el-form
            ref="pwdFormRef"
            :model="pwdForm"
            :rules="pwdRules"
            label-width="80px"
            class="profile-form"
          >
            <el-form-item label="旧密码" prop="old_password">
              <el-input v-model="pwdForm.old_password" type="password" show-password placeholder="请输入旧密码" />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input v-model="pwdForm.new_password" type="password" show-password placeholder="请输入新密码（至少6位）" />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm_password">
              <el-input v-model="pwdForm.confirm_password" type="password" show-password placeholder="请再次输入新密码" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleChangePassword" :loading="pwdLoading">修改密码</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useUserStore } from '../../store/user'
import { authApi } from '../../api'

const userStore = useUserStore()
const activeTab = ref('info')
const infoLoading = ref(false)
const pwdLoading = ref(false)
const infoFormRef = ref<FormInstance>()
const pwdFormRef = ref<FormInstance>()

const roleText = computed(() => {
  const roles = userStore.userInfo?.roles
  if (roles && roles.length > 0) {
    return roles.map((r: any) => r.name).join('、')
  }
  return '未分配角色'
})

// 基本信息表单
const infoForm = reactive({
  nickname: '',
  email: '',
  phone: '',
})

const infoRules: FormRules = {
  email: [
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' },
  ],
}

// 修改密码表单
const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const pwdRules: FormRules = {
  old_password: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: any) => {
        if (value !== pwdForm.new_password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

// 初始化表单数据
onMounted(async () => {
  if (!userStore.userInfo) {
    await userStore.fetchUserInfo()
  }
  syncFormFromStore()
})

function syncFormFromStore() {
  infoForm.nickname = userStore.userInfo?.nickname || ''
  infoForm.email = userStore.userInfo?.email || ''
  infoForm.phone = userStore.userInfo?.phone || ''
}

// 保存基本信息
async function handleUpdateInfo() {
  const valid = await infoFormRef.value?.validate().catch(() => false)
  if (!valid) return

  infoLoading.value = true
  try {
    const res: any = await authApi.updateProfile({
      nickname: infoForm.nickname,
      email: infoForm.email,
      phone: infoForm.phone,
    })
    // 更新 store 中的用户信息
    userStore.userInfo = res.data
    ElMessage.success('个人信息修改成功')
  } catch {
    // 错误已在拦截器中处理
  } finally {
    infoLoading.value = false
  }
}

// 修改密码
async function handleChangePassword() {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return

  pwdLoading.value = true
  try {
    await authApi.changePassword({
      old_password: pwdForm.old_password,
      new_password: pwdForm.new_password,
    })
    ElMessage.success('密码修改成功，请重新登录')
    // 清空表单
    pwdForm.old_password = ''
    pwdForm.new_password = ''
    pwdForm.confirm_password = ''
    pwdFormRef.value?.resetFields()
    // 退出登录
    setTimeout(() => {
      userStore.logout()
      window.location.href = '/login'
    }, 1500)
  } catch {
    // 错误已在拦截器中处理
  } finally {
    pwdLoading.value = false
  }
}
</script>

<style scoped>
.profile-page {
  max-width: 720px;
  margin: 0 auto;
}

.profile-card {
  background: var(--tech-bg-card);
  border-radius: var(--tech-radius-lg);
  border: 1px solid var(--tech-border);
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 24px;
  border-bottom: 1px solid var(--tech-border);
  font-size: 16px;
  font-weight: 600;
  color: var(--tech-text-primary);
}

.profile-banner {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 28px 24px;
  border-bottom: 1px solid var(--tech-border);
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.05), rgba(138, 43, 226, 0.05));
}

.avatar-large {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--tech-gradient-primary);
  color: #0a1628;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 700;
  flex-shrink: 0;
  box-shadow: 0 0 20px var(--tech-primary-glow);
}

.banner-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.banner-name {
  font-size: 20px;
  font-weight: 700;
  color: var(--tech-text-primary);
}

.banner-role {
  font-size: 13px;
  color: var(--tech-text-secondary);
  background: var(--tech-primary-bg);
  padding: 2px 12px;
  border-radius: 10px;
  display: inline-block;
  width: fit-content;
}

/* Tabs 样式 */
.profile-tabs {
  padding: 0 24px 24px;
}
.profile-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: var(--tech-border);
}
.profile-tabs :deep(.el-tabs__item) {
  color: var(--tech-text-secondary);
  font-size: 14px;
}
.profile-tabs :deep(.el-tabs__item.is-active) {
  color: var(--tech-primary);
}
.profile-tabs :deep(.el-tabs__active-bar) {
  background-color: var(--tech-primary);
}

/* Form 样式 */
.profile-form {
  max-width: 480px;
  margin-top: 20px;
}
.profile-form :deep(.el-form-item__label) {
  color: var(--tech-text-regular);
}
.profile-form :deep(.el-input__wrapper) {
  background-color: var(--tech-bg-page);
  border-color: var(--tech-border);
  box-shadow: none;
}
.profile-form :deep(.el-input__wrapper:hover),
.profile-form :deep(.el-input__wrapper.is-focus) {
  border-color: var(--tech-primary);
  box-shadow: 0 0 0 1px var(--tech-primary) inset;
}
.profile-form :deep(.el-input__inner) {
  color: var(--tech-text-primary);
}
.profile-form :deep(.el-input__inner::placeholder) {
  color: var(--tech-text-placeholder);
}
.profile-form :deep(.el-input.is-disabled .el-input__wrapper) {
  background-color: rgba(0, 0, 0, 0.2);
  opacity: 0.6;
}
</style>
