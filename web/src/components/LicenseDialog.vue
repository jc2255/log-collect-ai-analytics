<template>
  <el-dialog
    v-model="visible"
    title="授权码绑定"
    width="520px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
    class="license-dialog"
  >
    <div class="license-info">
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
    <el-form @submit.prevent="handleActivate">
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
      <el-button @click="handleLogout">退出登录</el-button>
      <el-button type="primary" :loading="activating" @click="handleActivate">
        激活授权码
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { licenseApi } from '../api'
import { useUserStore } from '../store/user'

const props = defineProps<{
  machineId?: string
}>()

const emit = defineEmits<{
  (e: 'activated'): void
}>()

const visible = ref(true)
const licenseKey = ref('')
const fetchedMachineId = ref('')
const activating = ref(false)
const router = useRouter()
const userStore = useUserStore()

// 优先用 prop 传入的 machineId，否则用自己获取的
const displayMachineId = computed(() => props.machineId || fetchedMachineId.value)

// 获取机器ID（如果 prop 没传的话）
onMounted(async () => {
  if (props.machineId) return // prop 已有则不再请求
  try {
    const res: any = await licenseApi.getStatus()
    fetchedMachineId.value = res.data.machine_id || ''
  } catch (err) {
    console.error('获取机器ID失败:', err)
  }
})

function copyMachineId() {
  if (!displayMachineId.value) return
  navigator.clipboard.writeText(displayMachineId.value).then(() => {
    ElMessage.success('机器ID已复制到剪贴板')
  })
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
</style>
