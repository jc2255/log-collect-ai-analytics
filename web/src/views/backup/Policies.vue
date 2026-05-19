<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <el-button type="primary" @click="handleAdd"><el-icon><Plus /></el-icon>新增策略</el-button>
    </div>
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="策略名称" width="150" />
      <el-table-column prop="log_store" label="日志库" width="120" />
      <el-table-column label="执行频率" width="100">
        <template #default="{ row }">{{ freqMap[row.frequency] || row.frequency }}</template>
      </el-table-column>
      <el-table-column prop="retention_days" label="快照保留天数" width="120" />
      <el-table-column prop="min_count" label="最少快照数" width="100" />
      <el-table-column prop="max_count" label="最多快照数" width="100" />
      <el-table-column prop="repository" label="OSS仓库" width="130" />
      <el-table-column label="操作" fixed="right" width="260">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="success" @click="handleExecute(row.id)">执行</el-button>
          <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
            <template #reference><el-button size="small" type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑策略' : '新增策略'" width="550px" :close-on-click-modal="false">
      <el-form :model="form" label-width="120px">
        <el-form-item label="策略名称" required><el-input v-model="form.name" placeholder="请输入策略名称" /></el-form-item>
        <el-form-item label="日志库" required>
          <el-select v-model="form.log_store" placeholder="选择日志库" style="width:100%" @change="onLogStoreChange">
            <el-option v-for="s in storeList" :key="s.name" :label="s.name" :value="s.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="执行频率" required>
          <el-select v-model="form.frequency" style="width:100%">
            <el-option label="每天" value="every_day" />
            <el-option label="每周" value="every_week" />
            <el-option label="每月" value="every_month" />
          </el-select>
        </el-form-item>
        <el-form-item label="快照保留天数"><el-input-number v-model="form.retention_days" :min="1" controls-position="right" /></el-form-item>
        <el-form-item label="最少快照数"><el-input-number v-model="form.min_count" :min="1" controls-position="right" /></el-form-item>
        <el-form-item label="最多快照数"><el-input-number v-model="form.max_count" :min="1" controls-position="right" /></el-form-item>
        <el-form-item label="Cron表达式"><el-input v-model="form.cron_expression" placeholder="自定义Cron（可选，覆盖频率）" /></el-form-item>
        <el-form-item label="OSS仓库名称">
          <el-input v-model="form.repository" placeholder="自动从日志库读取" />
          <div class="form-hint" v-if="form.repository">将从日志库的OSS配置自动创建快照仓库</div>
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
import { slmApi, logStoreApi } from '../../api'

const freqMap: Record<string, string> = { every_day: '每天', every_week: '每周', every_month: '每月' }

const loading = ref(false)
const tableData = ref<any[]>([])
const storeList = ref<any[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({
  name: '', log_store: '', frequency: 'every_day',
  retention_days: 30, min_count: 5, max_count: 100,
  cron_expression: '', repository: '',
})

async function fetchData() {
  loading.value = true
  try {
    const res: any = await slmApi.list()
    tableData.value = res.data.list || res.data || []
  } finally { loading.value = false }
}

async function fetchStores() {
  try {
    const res: any = await logStoreApi.list()
    storeList.value = res.data.list || res.data || []
  } catch { /* */ }
}

function onLogStoreChange(name: string) {
  const store = storeList.value.find((s: any) => s.name === name)
  if (store) {
    form.repository = store.oss_repository || ''
  }
}

function handleAdd() {
  editingId.value = null
  Object.assign(form, {
    name: '', log_store: '', frequency: 'every_day',
    retention_days: 30, min_count: 5, max_count: 100,
    cron_expression: '', repository: '',
  })
  dialogVisible.value = true
}

function handleEdit(row: any) {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name, log_store: row.log_store,
    frequency: row.frequency || 'every_day',
    retention_days: row.retention_days || 30,
    min_count: row.min_count || 5,
    max_count: row.max_count || 100,
    cron_expression: row.cron_expression || '',
    repository: row.repository || '',
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!form.name || !form.log_store) {
    ElMessage.warning('请填写策略名称和日志库')
    return
  }
  try {
    if (editingId.value) {
      await slmApi.update(editingId.value, form)
    } else {
      await slmApi.create(form)
    }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchData()
  } catch { /* */ }
}

async function handleDelete(id: number) {
  await slmApi.delete(id)
  ElMessage.success('删除成功')
  fetchData()
}

async function handleExecute(id: number) {
  try {
    await slmApi.execute(id)
    ElMessage.success('执行任务已提交')
  } catch { /* */ }
}

onMounted(() => { fetchData(); fetchStores() })
</script>

<style scoped>
.page-toolbar { margin-bottom: 16px; }
.form-hint { margin-top: 4px; color: #909399; font-size: 12px; }
</style>
