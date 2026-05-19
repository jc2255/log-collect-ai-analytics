<template>
  <div class="login-container">
    <!-- 动态网格背景 -->
    <div class="grid-bg"></div>
    <!-- 光晕装饰 -->
    <div class="glow glow-1"></div>
    <div class="glow glow-2"></div>
    <div class="glow glow-3"></div>

    <div class="login-wrapper">
      <!-- 左侧品牌区 -->
      <div class="login-brand">
        <div class="brand-content">
          <div class="brand-icon">L</div>
          <h1 class="brand-title">LCA</h1>
          <p class="brand-subtitle">日志收集智能分析平台</p>
          <div class="brand-features">
            <div class="feature-item">
              <span class="feature-dot"></span>
              <span>实时日志采集与分析</span>
            </div>
            <div class="feature-item">
              <span class="feature-dot"></span>
              <span>智能告警与异常检测</span>
            </div>
            <div class="feature-item">
              <span class="feature-dot"></span>
              <span>多维度数据可视化</span>
            </div>
          </div>
        </div>
        <!-- 装饰线条 -->
        <div class="brand-deco-line"></div>
      </div>

      <!-- 右侧登录表单 -->
      <div class="login-card">
        <div class="card-inner">
          <div class="card-header">
            <h2 class="card-title">欢迎登录</h2>
            <p class="card-desc">登录以访问日志分析平台</p>
          </div>
          <el-form ref="formRef" :model="form" :rules="rules" @keyup.enter="handleLogin" class="login-form">
            <el-form-item prop="username">
              <el-input v-model="form.username" placeholder="请输入用户名" prefix-icon="User" size="large" />
            </el-form-item>
            <el-form-item prop="password">
              <el-input v-model="form.password" placeholder="请输入密码" prefix-icon="Lock" type="password" size="large" show-password />
            </el-form-item>
            <el-form-item prop="captcha_code">
              <div class="captcha-row">
                <el-input v-model="form.captcha_code" placeholder="验证码" prefix-icon="Key" size="large" style="flex:1" />
                <img :src="captchaImage" class="captcha-img" @click="refreshCaptcha" title="点击刷新" />
              </div>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handleLogin" size="large" class="login-btn">
                登 录
              </el-button>
            </el-form-item>
          </el-form>
          <div class="card-footer">
            <span>管理员: admin / admin123</span>
            <span>测试用户: zhangsan / 123456</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部版权 -->
    <div class="login-footer">
      <span>LCA Log Analytics Platform &copy; 2024</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../../store/user'
import { ElMessage, type FormInstance } from 'element-plus'
import request from '../../api/request'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const captchaImage = ref('')
const captchaId = ref('')

const form = reactive({
  username: '',
  password: '',
  captcha_code: '',
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captcha_code: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
}

async function refreshCaptcha() {
  try {
    const res: any = await request.get('/captcha')
    captchaId.value = res.data.captcha_id
    captchaImage.value = res.data.captcha_image
  } catch {
    // ignore
  }
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res: any = await request.post('/auth/login', {
      username: form.username,
      password: form.password,
      captcha_id: captchaId.value,
      captcha_code: form.captcha_code,
    })
    // Store token and user info
    userStore.token = res.data.token
    userStore.userInfo = {
      id: res.data.user_id,
      username: res.data.username,
      nickname: res.data.nickname,
      dept_id: 0,
      post_id: 0,
      status: 1,
      roles: [],
    }
    localStorage.setItem('lca_token', res.data.token)
    ElMessage.success('登录成功')
    router.push('/')
  } catch {
    refreshCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(refreshCaptcha)
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--tech-gradient-login);
  position: relative;
  overflow: hidden;
}

/* 动态网格背景 */
.grid-bg {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 212, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 212, 255, 0.03) 1px, transparent 1px);
  background-size: 60px 60px;
  animation: gridMove 20s linear infinite;
}
@keyframes gridMove {
  0% { transform: translate(0, 0); }
  100% { transform: translate(60px, 60px); }
}

/* 光晕 */
.glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  pointer-events: none;
}
.glow-1 {
  width: 500px;
  height: 500px;
  background: rgba(0, 212, 255, 0.08);
  top: -10%;
  left: -5%;
  animation: glowFloat1 8s ease-in-out infinite;
}
.glow-2 {
  width: 400px;
  height: 400px;
  background: rgba(99, 102, 241, 0.08);
  bottom: -10%;
  right: -5%;
  animation: glowFloat2 10s ease-in-out infinite;
}
.glow-3 {
  width: 300px;
  height: 300px;
  background: rgba(14, 165, 233, 0.06);
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation: glowFloat3 12s ease-in-out infinite;
}
@keyframes glowFloat1 {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(30px, 20px); }
}
@keyframes glowFloat2 {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(-20px, -30px); }
}
@keyframes glowFloat3 {
  0%, 100% { transform: translate(-50%, -50%) scale(1); }
  50% { transform: translate(-50%, -50%) scale(1.2); }
}

/* 登录包装器 */
.login-wrapper {
  display: flex;
  border-radius: var(--tech-radius-xl);
  overflow: hidden;
  box-shadow: var(--tech-shadow-lg), 0 0 60px rgba(0, 212, 255, 0.1);
  border: 1px solid var(--tech-border-active);
  position: relative;
  z-index: 1;
  max-width: 860px;
  width: 90%;
}

/* 左侧品牌区 */
.login-brand {
  width: 380px;
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.08) 0%, rgba(99, 102, 241, 0.06) 100%);
  backdrop-filter: blur(20px);
  padding: 60px 40px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  position: relative;
  border-right: 1px solid var(--tech-border);
}
.brand-content {
  position: relative;
  z-index: 1;
}
.brand-icon {
  width: 64px;
  height: 64px;
  background: var(--tech-gradient-primary);
  border-radius: var(--tech-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0a1628;
  font-size: 32px;
  font-weight: 800;
  margin-bottom: 24px;
}
.brand-title {
  font-size: 36px;
  font-weight: 800;
  background: var(--tech-gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 8px 0;
  letter-spacing: 2px;
}
.brand-subtitle {
  font-size: 16px;
  color: var(--tech-text-regular);
  margin: 0 0 40px 0;
  letter-spacing: 1px;
}
.brand-features {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
  color: var(--tech-text-secondary);
}
.feature-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--tech-primary);
  box-shadow: 0 0 8px var(--tech-primary-glow);
  flex-shrink: 0;
}
.brand-deco-line {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 1px;
  background: linear-gradient(180deg, transparent 0%, var(--tech-primary) 50%, transparent 100%);
  opacity: 0.3;
}

/* 右侧登录卡片 */
.login-card {
  flex: 1;
  background: rgba(17, 29, 49, 0.95);
  backdrop-filter: blur(20px);
  padding: 0;
}
.card-inner {
  padding: 50px 40px;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.card-header {
  margin-bottom: 36px;
}
.card-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--tech-text-primary);
  margin: 0 0 8px 0;
}
.card-desc {
  font-size: 14px;
  color: var(--tech-text-secondary);
  margin: 0;
}
.login-form {
  width: 100%;
}
.captcha-row {
  display: flex;
  gap: 12px;
  width: 100%;
}
.captcha-img {
  height: 40px;
  cursor: pointer;
  border-radius: var(--tech-radius-sm);
  border: 1px solid var(--tech-border);
  transition: border-color 0.3s;
}
.captcha-img:hover {
  border-color: var(--tech-primary);
}
.login-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  letter-spacing: 4px;
  border-radius: var(--tech-radius-md);
  background: var(--tech-gradient-primary) !important;
  border: none !important;
  color: #0a1628 !important;
  font-weight: 700;
  transition: all 0.3s ease;
}
.login-btn:hover {
  box-shadow: 0 0 24px rgba(0, 212, 255, 0.4);
  transform: translateY(-1px);
}
.card-footer {
  margin-top: 24px;
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--tech-text-placeholder);
}

/* 底部版权 */
.login-footer {
  position: absolute;
  bottom: 24px;
  left: 0;
  right: 0;
  text-align: center;
  font-size: 12px;
  color: var(--tech-text-placeholder);
  z-index: 1;
}

@media (max-width: 768px) {
  .login-brand {
    display: none;
  }
  .login-wrapper {
    max-width: 420px;
  }
}
</style>
