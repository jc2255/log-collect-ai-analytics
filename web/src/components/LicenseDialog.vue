<template>
  <el-dialog
    v-model="visible"
    title="授权码管理"
    width="520px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
    class="license-dialog"
  >
    <!-- 已激活状态 -->
    <div v-if="activated" class="license-info">
      <el-alert
        type="success"
        :closable="false"
        show-icon
        title="系统已激活授权码"
      />
      <div class="license-detail">
        <div class="detail-row">
          <span class="label">授权类型：</span>
          <span class="value">{{ licenseTypeText }}</span>
        </div>
        <div class="detail-row" v-if="expiresAt">
          <span class="label">到期时间：</span>
          <span class="value">{{ expiresAt }}</span>
        </div>
        <div class="detail-row" v-else>
          <span class="label">到期时间：</span>
          <span class="value permanent">永久有效</span>
        </div>
        <div class="detail-row">
          <span class="label">绑定时间：</span>
          <span class="value">{{ boundAt }}</span>
        </div>
        <div class="machine-id-row">
          <span class="label">当前机器ID：</span>
          <code class="machine-id" v-if="displayMachineId">{{ displayMachineId }}</code>
          <code class="machine-id loading" v-else>获取中...</code>
          <el-button size="small" type="primary" link @click="copyMachineId" :disabled="!displayMachineId">复制</el-button>
        </div>
      </div>
      <el-divider />
      <p class="unbind-hint">如需更换授权码或迁移到其他机器，请先解绑当前授权码。</p>
    </div>

    <!-- 未激活状态 -->
    <div v-else class="license-info">
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        title="系统尚未绑定授权码，请输入授权码后继续使用"
      />
      <div class="machine-id-row">
        <span class="label">当前机器ID：</span>
        <code class="machine-id" v-if="displayMachineId">{{ displayMachineId }}</code>
        <code class="machine-id loading" v-else>获取中...</code>
        <el-button size="small" type="primary" link @click="copyMachineId" :disabled="!displayMachineId">复制</el-button>
      </div>
      <p class="hint">请将机器ID复制到 LCA官网 生成对应授权码</p>
    </div>
    <el-form v-if="!activated" @submit.prevent="handleActivate">
      <el-form-item>
        <el-input
          v-model="licenseKey"
          placeholder="请输入授权码"
          type="textarea"
          :rows="3"
          size="large"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <template v-if="activated">
        <el-button @click="handleClose">关闭</el-button>
        <el-button type="danger" :loading="deactivating" @click="handleDeactivate">
          解绑授权码
        </el-button>
      </template>
      <template v-else>
        <el-button @click="handleLogout">退出登录</el-button>
        <el-button type="primary" :loading="activating" @click="handleActivate">
          激活授权码
        </el-button>
      </template>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { licenseApi } from '../api'
import { useUserStore } from '../store/user'

const props = defineProps<{
  machineId?: string
  licenseStatus?: {
    activated: boolean
    license_type?: string
    expires_at?: string
    bound_at?: string
  }
}>()

const emit = defineEmits<{
  (e: 'activated'): void
  (e: 'deactivated'): void
}>()

const visible = ref(true)
const licenseKey = ref('')
const fetchedMachineId = ref('')
const activating = ref(false)
const deactivating = ref(false)
const router = useRouter()
const userStore = useUserStore()

// 已激活信息（优先从 prop 取，否则从接口获取）
const activated = ref(false)
const licenseType = ref('')
const expiresAt = ref('')
const boundAt = ref('')

const licenseTypeText = computed(() => {
  const map: Record<string, string> = { monthly: '月付授权', yearly: '年付授权', permanent: '永久授权' }
  return map[licenseType.value] || licenseType.value || '—'
})

// 优先用 prop 传入的 machineId，否则用自己获取的
const displayMachineId = computed(() => props.machineId || fetchedMachineId.value)

// 加载授权状态
onMounted(async () => {
  // 从 prop 初始化
  if (props.licenseStatus?.activated) {
    activated.value = true
    licenseType.value = props.licenseStatus.license_type || ''
    expiresAt.value = props.licenseStatus.expires_at || ''
    boundAt.value = props.licenseStatus.bound_at || ''
  }
  // 获取机器ID（如果 prop 没传的话）
  if (!props.machineId) {
    try {
      const res: any = await licenseApi.getStatus()
      fetchedMachineId.value = res.data.machine_id || ''
      // 如果 prop 没传状态，也从接口取
      if (!props.licenseStatus) {
        activated.value = res.data.activated === true
        licenseType.value = res.data.license_type || ''
        expiresAt.value = res.data.expires_at || ''
        boundAt.value = res.data.bound_at || ''
      }
    } catch (err) {
      console.error('获取机器ID失败:', err)
    }
  }
})

function copyMachineId() {
  if (!displayMachineId.value) return
  const text = displayMachineId.value
  // 优先使用 Clipboard API（仅 HTTPS 或 localhost 下可用）
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(text)
      .then(() => ElMessage.success('机器ID已复制到剪贴板'))
      .catch(() => fallbackCopy(text))
  } else {
    // HTTP 环境降级使用 execCommand
    fallbackCopy(text)
  }
}

function fallbackCopy(text: string) {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.top = '0'
  textarea.style.left = '0'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.focus()
  textarea.select()
  try {
    const ok = document.execCommand('copy')
    if (ok) {
      ElMessage.success('机器ID已复制到剪贴板')
    } else {
      ElMessage.warning('复制失败，请手动选中复制')
    }
  } catch (e) {
    ElMessage.warning('复制失败，请手动选中复制')
  } finally {
    document.body.removeChild(textarea)
  }
}

async function handleActivate() {
  if (!licenseKey.value.trim()) {
    ElMessage.warning('请输入授权码')
    return
  }
  activating.value = true
  try {
    await licenseApi.activate(licenseKey.value.trim())
    ElMessage.success('授权码激活成功')
    visible.value = false
    emit('activated')
  } catch (err: any) {
    const msg = err?.response?.data?.message || '激活失败'
    ElMessage.error(msg)
  } finally {
    activating.value = false
  }
}

async function handleDeactivate() {
  try {
    await ElMessageBox.confirm(
      '解绑后系统将无法使用，需重新激活授权码。确定解绑？',
      '解绑确认',
      { confirmButtonText: '确定解绑', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return // 取消
  }
  deactivating.value = true
  try {
    await licenseApi.deactivate()
    ElMessage.success('授权码已解绑')
    visible.value = false
    emit('deactivated')
  } catch (err: any) {
    const msg = err?.response?.data?.message || '解绑失败'
    ElMessage.error(msg)
  } finally {
    deactivating.value = false
  }
}

function handleClose() {
  visible.value = false
}

function handleLogout() {
  userStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.license-info {
  margin-bottom: 20px;
}
.machine-id-row {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.machine-id-row .label {
  font-size: 14px;
  color: var(--tech-text-regular);
  white-space: nowrap;
}
.machine-id {
  flex: 1;
  font-size: 12px;
  background: rgba(0, 212, 255, 0.06);
  border: 1px solid var(--tech-border);
  border-radius: 4px;
  padding: 4px 8px;
  color: var(--tech-primary);
  word-break: break-all;
  max-height: 60px;
  overflow-y: auto;
}
.machine-id.loading {
  color: var(--tech-text-placeholder);
  font-style: italic;
}
.hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--tech-text-placeholder);
}

/* 已激活详情 */
.license-detail {
  margin-top: 16px;
}
.detail-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.detail-row .label {
  font-size: 14px;
  color: var(--tech-text-regular);
  white-space: nowrap;
}
.detail-row .value {
  font-size: 14px;
  color: var(--tech-text-primary);
  font-weight: 500;
}
.detail-row .value.permanent {
  color: var(--tech-success);
}
.unbind-hint {
  font-size: 12px;
  color: var(--tech-text-placeholder);
  margin: 0;
}
</style>
