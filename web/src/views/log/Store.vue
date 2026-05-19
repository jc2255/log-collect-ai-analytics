<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <el-button type="primary" @click="handleAdd"><el-icon><Plus /></el-icon>添加日志库</el-button>
    </div>
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="日志库名称" width="150" />
      <el-table-column prop="api_key" label="API Key" width="200">
        <template #default="{ row }">
          <el-tooltip :content="row.api_key || '-'" placement="top">
            <span class="api-key-text">{{ maskKey(row.api_key) }}</span>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column prop="index_pattern" label="索引模式" width="180" />
      <el-table-column label="压缩" width="80">
        <template #default="{ row }">{{ row.compress ? '压缩' : '不压缩' }}</template>
      </el-table-column>
      <el-table-column label="ILM策略" width="200">
        <template #default="{ row }">
          滚动{{ row.roll_max_days || '-' }}天/{{ row.roll_max_size_gb || '-' }}GB →
          冷{{ row.cold_days || '-' }}天 → 删{{ row.delete_days }}天
        </template>
      </el-table-column>
      <el-table-column label="AI智能告警" width="160">
        <template #default="{ row }">
          <div style="display:flex;align-items:center;gap:8px">
            <el-switch
              v-model="row.ai_alert_enabled"
              @change="(val: boolean) => toggleAIAlert(row, val)"
              active-text=""
              inactive-text=""
            />
            <el-button v-if="row.ai_alert_enabled" size="small" type="primary" link @click="openAIAlertConfig(row)">
              <el-icon><Setting /></el-icon>配置
            </el-button>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
            <template #reference><el-button size="small" type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新增/编辑日志库对话框 -->
    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑日志库' : '添加日志库'" width="600px" :close-on-click-modal="false">
      <el-form :model="form" label-width="120px">
        <el-form-item label="日志库名称" required>
          <el-input v-model="form.name" placeholder="请输入日志库名称" :disabled="!!editingId" />
        </el-form-item>

        <el-form-item label="API Key">
          <div style="display:flex;gap:8px">
            <el-input v-model="form.api_key" placeholder="留空则自动生成" style="flex:1" />
            <el-button @click="genKey" type="primary" plain><el-icon><Refresh /></el-icon>生成</el-button>
          </div>
          <div class="form-hint">Agent 推送日志时需要此 Key 进行身份验证</div>
        </el-form-item>

        <el-form-item label="是否压缩存储">
          <el-radio-group v-model="form.compress">
            <el-radio :value="true">压缩（可节约30%-50%的空间）</el-radio>
            <el-radio :value="false">不压缩（查询速度快）</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-divider content-position="left">索引滚动策略</el-divider>

        <el-form-item label="滚动最大天数">
          <el-input-number v-model="form.roll_max_days" :min="0" placeholder="例如：7（天后滚动）" controls-position="right" />
          <span class="form-hint">0表示不限制</span>
        </el-form-item>
        <el-form-item label="滚动最大容量">
          <el-input-number v-model="form.roll_max_size_gb" :min="0" placeholder="例如：50（GB后滚动）" controls-position="right" />
          <span class="form-hint">0表示不限制</span>
        </el-form-item>

        <el-divider content-position="left">数据归档与清理</el-divider>

        <el-form-item label="冷存储天数">
          <el-input-number v-model="form.cold_days" :min="0" placeholder="例如：30（30天后冷存储）" controls-position="right" />
          <span class="form-hint">0表示不启用</span>
        </el-form-item>
        <el-form-item label="删除数据天数" required>
          <el-input-number v-model="form.delete_days" :min="1" placeholder="例如：90（90天后删除）" controls-position="right" />
        </el-form-item>

        <el-divider content-position="left">OSS 备份配置</el-divider>

        <el-form-item label="仓库名称">
          <el-input v-model="form.oss_repository" placeholder="my_oss_backup" />
        </el-form-item>
        <el-form-item label="Endpoint">
          <el-input v-model="form.oss_endpoint" placeholder="http://oss-cn-beijing.aliyuncs.com" />
        </el-form-item>
        <el-form-item label="Bucket名称">
          <el-input v-model="form.oss_bucket" placeholder="my-bucket" />
        </el-form-item>
        <el-form-item label="AccessKeyID">
          <el-input v-model="form.oss_access_key_id" placeholder="请输入Access Key ID" show-password />
        </el-form-item>
        <el-form-item label="AccessKeySecret">
          <el-input v-model="form.oss_access_key_secret" placeholder="请输入Access Key Secret" show-password />
        </el-form-item>
        <el-form-item label="存储路径">
          <el-input v-model="form.oss_path" placeholder="lca/" />
        </el-form-item>
        <el-form-item label="分块大小">
          <el-input v-model="form.oss_chunk_size" placeholder="500mb" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>

    <!-- AI智能告警配置对话框 -->
    <el-dialog v-model="aiAlertDialogVisible" title="AI 智能告警配置" width="650px" :close-on-click-modal="false">
      <el-form :model="aiAlertForm" label-width="130px">
        <el-divider content-position="left">扫描规则</el-divider>

        <el-form-item label="扫描频率(分钟)">
          <el-input-number v-model="aiAlertForm.scan_interval_minutes" :min="1" :max="60" controls-position="right" />
          <span class="form-hint">每隔多少分钟扫描一次日志</span>
        </el-form-item>

        <el-form-item label="ERROR阈值">
          <el-input-number v-model="aiAlertForm.error_threshold" :min="1" controls-position="right" />
          <span class="form-hint">ERROR日志数量超过此值时触发AI分析</span>
        </el-form-item>

        <el-form-item label="监控关键词">
          <div class="keywords-wrap">
            <el-tag
              v-for="(kw, idx) in aiAlertForm.keywords"
              :key="idx"
              closable
              @close="aiAlertForm.keywords.splice(idx, 1)"
              style="margin-right:4px;margin-bottom:4px"
            >{{ kw }}</el-tag>
            <el-input
              v-if="keywordInputVisible"
              ref="keywordInputRef"
              v-model="keywordInputValue"
              size="small"
              style="width:120px"
              @keyup.enter="addKeyword"
              @blur="addKeyword"
            />
            <el-button v-else size="small" @click="showKeywordInput">+ 添加</el-button>
          </div>
        </el-form-item>

        <el-form-item label="静默时间(分钟)">
          <el-input-number v-model="aiAlertForm.silence_minutes" :min="1" :max="1440" controls-position="right" />
          <span class="form-hint">告警触发后多久内不重复告警</span>
        </el-form-item>

        <el-divider content-position="left">大模型配置</el-divider>

        <el-form-item label="LLM Provider">
          <el-select v-model="aiAlertForm.llm_provider" style="width:200px">
            <el-option label="OpenAI" value="openai" />
            <el-option label="DeepSeek" value="deepseek" />
            <el-option label="通义千问" value="qwen" />
            <el-option label="本地 Ollama" value="ollama" />
          </el-select>
        </el-form-item>

        <el-form-item label="API Base URL">
          <el-input v-model="aiAlertForm.llm_base_url" placeholder="https://api.openai.com/v1" />
        </el-form-item>

        <el-form-item label="Model">
          <el-input v-model="aiAlertForm.llm_model" placeholder="gpt-4o-mini" />
        </el-form-item>

        <el-form-item label="API Key">
          <el-input v-model="aiAlertForm.llm_api_key" placeholder="sk-xxx" show-password />
        </el-form-item>

        <el-divider content-position="left">通知渠道</el-divider>

        <div v-for="(ch, idx) in aiAlertForm.notify_channels" :key="idx" class="notify-channel-item">
          <el-form-item :label="'渠道 ' + (idx + 1)">
            <div style="display:flex;gap:8px;align-items:center;width:100%">
              <el-select v-model="ch.type" style="width:130px">
                <el-option label="企业微信" value="wecom" />
                <el-option label="钉钉" value="dingtalk" />
                <el-option label="邮件" value="email" />
                <el-option label="Webhook" value="webhook" />
              </el-select>
              <template v-if="ch.type === 'email'">
                <el-input v-model="ch.to" placeholder="收件邮箱" style="flex:1" />
              </template>
              <template v-else>
                <el-input v-model="ch.webhook_url" placeholder="Webhook URL" style="flex:1" />
              </template>
              <el-button type="danger" link @click="aiAlertForm.notify_channels.splice(idx, 1)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <!-- 邮件SMTP配置 -->
            <div v-if="ch.type === 'email'" style="margin-top:8px;display:flex;gap:8px;flex-wrap:wrap">
              <el-input v-model="ch.smtp_host" placeholder="SMTP服务器，如 smtp.qq.com" style="width:200px" size="small" />
              <el-input-number v-model="ch.smtp_port" :min="1" :max="65535" placeholder="端口" style="width:100px" size="small" controls-position="right" />
              <el-input v-model="ch.smtp_user" placeholder="SMTP用户名/邮箱" style="width:200px" size="small" />
              <el-input v-model="ch.smtp_pass" placeholder="SMTP密码/授权码" style="width:160px" size="small" show-password />
            </div>
          </el-form-item>
        </div>
        <el-form-item label="">
          <el-button @click="addNotifyChannel">+ 添加通知渠道</el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testAIAlert" :loading="testLoading">测试告警</el-button>
        <el-button @click="aiAlertDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveAIAlertConfig">保存配置</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { logStoreApi, aiAlertApi } from '../../api'

const loading = ref(false)
const tableData = ref<any[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  name: '',
  index_pattern: '',
  api_key: '',
  compress: true,
  roll_max_days: 0,
  roll_max_size_gb: 0,
  cold_days: 0,
  delete_days: 90,
  oss_repository: '',
  oss_endpoint: '',
  oss_bucket: '',
  oss_access_key_id: '',
  oss_access_key_secret: '',
  oss_path: 'lca/',
  oss_chunk_size: '500mb',
  description: '',
})

function resetForm() {
  Object.assign(form, {
    name: '', index_pattern: '', api_key: '', compress: true,
    roll_max_days: 0, roll_max_size_gb: 0, cold_days: 0, delete_days: 90,
    oss_repository: '', oss_endpoint: '', oss_bucket: '',
    oss_access_key_id: '', oss_access_key_secret: '',
    oss_path: 'lca/', oss_chunk_size: '500mb', description: '',
  })
}

function maskKey(key: string) {
  if (!key) return '-'
  if (key.length <= 12) return key
  return key.slice(0, 8) + '****' + key.slice(-4)
}

function genKey() {
  const rand = Math.random().toString(36).substring(2, 10)
  const name = form.name || 'default'
  form.api_key = 'ak_' + name + '_' + rand
}

async function fetchData() {
  loading.value = true
  try {
    const res: any = await logStoreApi.list()
    tableData.value = res.data.list || res.data || []
  } finally { loading.value = false }
}

function handleAdd() {
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

function handleEdit(row: any) {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    index_pattern: row.index_pattern || '',
    api_key: row.api_key || '',
    compress: row.compress ?? true,
    roll_max_days: row.roll_max_days || 0,
    roll_max_size_gb: row.roll_max_size_gb || 0,
    cold_days: row.cold_days || 0,
    delete_days: row.delete_days || 90,
    oss_repository: row.oss_repository || '',
    oss_endpoint: row.oss_endpoint || '',
    oss_bucket: row.oss_bucket || '',
    oss_access_key_id: row.oss_access_key_id || '',
    oss_access_key_secret: row.oss_access_key_secret || '',
    oss_path: row.oss_path || 'lca/',
    oss_chunk_size: row.oss_chunk_size || '500mb',
    description: row.description || '',
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!form.name) {
    ElMessage.warning('请输入日志库名称')
    return
  }
  if (!form.delete_days || form.delete_days < 1) {
    ElMessage.warning('请输入删除数据天数')
    return
  }
  try {
    if (editingId.value) {
      await logStoreApi.update(editingId.value, form)
    } else {
      await logStoreApi.create(form)
    }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchData()
  } catch { /* */ }
}

async function handleDelete(id: number) {
  await logStoreApi.delete(id)
  ElMessage.success('删除成功')
  fetchData()
}

// ===== AI 智能告警 =====
const aiAlertDialogVisible = ref(false)
const aiAlertStoreId = ref<number>(0)
const testLoading = ref(false)
const keywordInputVisible = ref(false)
const keywordInputValue = ref('')
const keywordInputRef = ref<any>(null)

const aiAlertForm = reactive({
  scan_interval_minutes: 5,
  error_threshold: 10,
  keywords: ['timeout', 'OOM', 'panic', 'fatal'] as string[],
  llm_provider: 'openai',
  llm_base_url: 'https://api.openai.com/v1',
  llm_model: 'gpt-4o-mini',
  llm_api_key: '',
  notify_channels: [] as Array<{ type: string; webhook_url?: string; to?: string; smtp_host?: string; smtp_port?: number; smtp_user?: string; smtp_pass?: string }>,
  silence_minutes: 30,
})

async function toggleAIAlert(row: any, val: boolean) {
  try {
    await aiAlertApi.toggle(row.id, val)
    ElMessage.success(val ? 'AI智能告警已启用' : 'AI智能告警已关闭')
    if (val && !row.ai_alert_config) {
      // 首次开启，弹出配置对话框
      openAIAlertConfig(row)
    }
  } catch {
    // 回滚开关状态
    row.ai_alert_enabled = !val
  }
}

async function openAIAlertConfig(row: any) {
  aiAlertStoreId.value = row.id
  // 加载已有配置
  try {
    const res: any = await aiAlertApi.getConfig(row.id)
    const configStr = res.data.ai_alert_config
    if (configStr) {
      const cfg = JSON.parse(configStr)
      Object.assign(aiAlertForm, {
        scan_interval_minutes: cfg.scan_interval_minutes || 5,
        error_threshold: cfg.error_threshold || 10,
        keywords: cfg.keywords || ['timeout', 'OOM', 'panic', 'fatal'],
        llm_provider: cfg.llm_provider || 'openai',
        llm_base_url: cfg.llm_base_url || 'https://api.openai.com/v1',
        llm_model: cfg.llm_model || 'gpt-4o-mini',
        llm_api_key: cfg.llm_api_key || '',
        notify_channels: cfg.notify_channels || [],
        silence_minutes: cfg.silence_minutes || 30,
      })
    } else {
      // 默认配置
      Object.assign(aiAlertForm, {
        scan_interval_minutes: 5,
        error_threshold: 10,
        keywords: ['timeout', 'OOM', 'panic', 'fatal'],
        llm_provider: 'openai',
        llm_base_url: 'https://api.openai.com/v1',
        llm_model: 'gpt-4o-mini',
        llm_api_key: '',
        notify_channels: [],
        silence_minutes: 30,
      })
    }
  } catch {
    // ignore
  }
  aiAlertDialogVisible.value = true
}

async function saveAIAlertConfig() {
  if (!aiAlertForm.llm_api_key && aiAlertForm.llm_provider !== 'ollama') {
    ElMessage.warning('请输入 LLM API Key')
    return
  }
  if (aiAlertForm.notify_channels.length === 0) {
    ElMessage.warning('请至少添加一个通知渠道')
    return
  }
  try {
    const configStr = JSON.stringify(aiAlertForm)
    await aiAlertApi.updateConfig(aiAlertStoreId.value, configStr)
    ElMessage.success('AI告警配置已保存')
    aiAlertDialogVisible.value = false
  } catch { /* */ }
}

async function testAIAlert() {
  testLoading.value = true
  try {
    // 先保存配置再测试
    const configStr = JSON.stringify(aiAlertForm)
    await aiAlertApi.updateConfig(aiAlertStoreId.value, configStr)
    await aiAlertApi.test(aiAlertStoreId.value)
    ElMessage.success('测试告警已触发，请检查通知渠道')
  } catch {
    ElMessage.error('测试失败')
  } finally {
    testLoading.value = false
  }
}

function addNotifyChannel() {
  aiAlertForm.notify_channels.push({ type: 'wecom', webhook_url: '' })
}

function showKeywordInput() {
  keywordInputVisible.value = true
  nextTick(() => {
    keywordInputRef.value?.focus()
  })
}

function addKeyword() {
  const val = keywordInputValue.value.trim()
  if (val && !aiAlertForm.keywords.includes(val)) {
    aiAlertForm.keywords.push(val)
  }
  keywordInputVisible.value = false
  keywordInputValue.value = ''
}

onMounted(fetchData)
</script>

<style scoped>
.page-toolbar { margin-bottom: 16px; }
.form-hint { margin-left: 8px; color: #909399; font-size: 12px; }
:deep(.el-divider__text) { font-weight: 600; color: #303133; }
.keywords-wrap { display: flex; flex-wrap: wrap; align-items: center; gap: 4px; }
.notify-channel-item { margin-bottom: 4px; }
</style>
