<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <el-button type="primary" @click="handleAdd"><el-icon><Plus /></el-icon>新增采集任务</el-button>
    </div>
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column label="执行Agent" width="120">
        <template #default="{ row }">
          <el-tag v-if="row.agent_id === 0" type="info" size="small">所有Agent</el-tag>
          <span v-else>{{ getAgentName(row.agent_id) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="store_name" label="目标日志库" width="150" />
      <el-table-column prop="log_path_pattern" label="日志路径" min-width="250" />
      <el-table-column prop="parse_mode" label="解析模式" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ parseModeMap[row.parse_mode] || row.parse_mode }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" :type="row.status === 1 ? 'warning' : 'success'" @click="toggleStatus(row)">
            {{ row.status === 1 ? '禁用' : '启用' }}
          </el-button>
          <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
            <template #reference><el-button size="small" type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑采集任务' : '新增采集任务'" width="550px" :close-on-click-modal="false">
      <el-form :model="form" label-width="120px">
        <el-form-item label="执行Agent">
          <el-select v-model="form.agent_id" placeholder="选择Agent" style="width:100%">
            <el-option :value="0" label="所有Agent（每台机器都执行）" />
            <el-option v-for="a in agentList" :key="a.id" :label="`${a.hostname} (${a.ip})`" :value="a.id">
              <span>{{ a.hostname }}</span>
              <span style="float:right;color:#999;font-size:12px">{{ a.ip }}</span>
            </el-option>
          </el-select>
          <div class="form-hint">选"所有Agent"则每台注册的Agent都会执行此采集任务</div>
        </el-form-item>
        <el-form-item label="目标日志库" required>
          <el-select v-model="form.store_id" placeholder="选择日志库" style="width:100%" @change="onStoreChange">
            <el-option v-for="s in storeList" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="日志路径" required>
          <el-input v-model="form.log_path_pattern" placeholder="如: /var/log/nginx/*.log 或 /app/logs/**/*.log" />
          <div class="form-hint">支持 glob 通配符，如 /var/log/*.log</div>
        </el-form-item>
        <el-form-item label="解析模式">
          <el-select v-model="form.parse_mode" style="width:100%">
            <el-option label="原始文本(raw)" value="raw" />
            <el-option label="JSON解析" value="json" />
            <el-option label="正则解析" value="regex" />
            <el-option label="分隔符解析" value="delimiter" />
          </el-select>
        </el-form-item>
        <el-form-item label="多行匹配正则" v-if="form.parse_mode === 'raw' || form.parse_mode === 'regex'">
          <el-input v-model="form.multiline_pattern" placeholder="如: ^\d{4}-\d{2}-\d{2} 匹配以日期开头的行" />
          <div class="form-hint">不匹配的行将追加到上一条日志（用于多行堆栈日志）</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { collectApi, logStoreApi } from '../../api'

const parseModeMap: Record<string, string> = { raw: '原始文本', json: 'JSON', regex: '正则', delimiter: '分隔符' }

const loading = ref(false)
const tableData = ref<any[]>([])
const storeList = ref<any[]>([])
const agentList = ref<any[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({
  agent_id: 0 as number,
  store_id: undefined as number | undefined,
  store_name: '',
  log_path_pattern: '',
  multiline_pattern: '',
  parse_mode: 'raw',
})

function getAgentName(agentId: number) {
  const agent = agentList.value.find((a: any) => a.id === agentId)
  return agent ? agent.hostname : `Agent#${agentId}`
}

async function fetchData() {
  loading.value = true
  try {
    const res: any = await collectApi.listTasks()
    tableData.value = res.data?.list || []
  } finally { loading.value = false }
}

async function fetchStores() {
  try {
    const res: any = await logStoreApi.list()
    storeList.value = res.data?.list || res.data || []
  } catch { /* */ }
}

async function fetchAgents() {
  try {
    const res: any = await collectApi.listAgents()
    agentList.value = res.data?.list || []
  } catch { /* */ }
}

function onStoreChange(id: number) {
  const store = storeList.value.find((s: any) => s.id === id)
  if (store) form.store_name = store.name
}

function handleAdd() {
  editingId.value = null
  Object.assign(form, { agent_id: 0, store_id: undefined, store_name: '', log_path_pattern: '', multiline_pattern: '', parse_mode: 'raw' })
  dialogVisible.value = true
}

function handleEdit(row: any) {
  editingId.value = row.id
  Object.assign(form, {
    agent_id: row.agent_id || 0,
    store_id: row.store_id,
    store_name: row.store_name,
    log_path_pattern: row.log_path_pattern,
    multiline_pattern: row.multiline_pattern || '',
    parse_mode: row.parse_mode || 'raw',
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!form.store_id || !form.log_path_pattern) {
    ElMessage.warning('请选择日志库和填写日志路径')
    return
  }
  try {
    if (editingId.value) {
      await collectApi.updateTask(editingId.value, form)
    } else {
      await collectApi.createTask(form)
    }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchData()
  } catch { /* */ }
}

async function handleDelete(id: number) {
  await collectApi.deleteTask(id)
  ElMessage.success('删除成功')
  fetchData()
}

async function toggleStatus(row: any) {
  const newStatus = row.status === 1 ? 0 : 1
  await collectApi.updateTask(row.id, { status: newStatus })
  fetchData()
}

onMounted(() => { fetchData(); fetchStores(); fetchAgents() })
</script>

<style scoped>
.page-toolbar { margin-bottom: 16px; }
.form-hint { margin-top: 4px; color: #909399; font-size: 12px; }
</style>
