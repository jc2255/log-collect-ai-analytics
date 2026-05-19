<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <el-button @click="fetchData"><el-icon><Refresh /></el-icon>刷新</el-button>
    </div>
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="hostname" label="主机名" width="180" />
      <el-table-column prop="ip" label="IP地址" width="150" />
      <el-table-column prop="os_type" label="系统" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ row.os_type || '-' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="版本" width="100" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="isOnline(row) ? 'success' : 'danger'" size="small" effect="dark">
            <el-icon style="vertical-align:middle;margin-right:4px"><component :is="isOnline(row) ? 'CircleCheck' : 'CircleClose'" /></el-icon>
            {{ isOnline(row) ? '在线' : '离线' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="最后心跳" width="180">
        <template #default="{ row }">
          <span v-if="row.last_heartbeat">{{ formatTime(row.last_heartbeat) }}</span>
          <span v-else style="color:#999">从未上线</span>
        </template>
      </el-table-column>
      <el-table-column label="离线时长" width="120">
        <template #default="{ row }">
          <span v-if="!isOnline(row) && row.last_heartbeat" style="color:#f56c6c">{{ offlineDuration(row) }}</span>
          <span v-else-if="isOnline(row)" style="color:#67c23a">-</span>
        </template>
      </el-table-column>
      <el-table-column label="采集任务" width="100">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="showTasks(row)">{{ getTaskCount(row.id) }}</el-button>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-popconfirm title="确认删除该Agent?" @confirm="handleDelete(row.id)">
            <template #reference><el-button size="small" type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- Agent关联的采集任务 -->
    <el-dialog v-model="taskDialogVisible" :title="`Agent [${currentAgent?.hostname}] 的采集任务`" width="700px">
      <el-table :data="agentTasks" stripe size="small">
        <el-table-column prop="store_name" label="日志库" width="130" />
        <el-table-column prop="log_path_pattern" label="日志路径" min-width="250" />
        <el-table-column prop="parse_mode" label="解析模式" width="90" />
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { collectApi } from '../../api'

const loading = ref(false)
const tableData = ref<any[]>([])
const allTasks = ref<any[]>([])
const taskDialogVisible = ref(false)
const currentAgent = ref<any>(null)
const agentTasks = ref<any[]>([])
let timer: ReturnType<typeof setInterval> | null = null

function isOnline(row: any) {
  if (!row.last_heartbeat) return false
  // 90秒内有心跳视为在线
  return (Date.now() / 1000 - Number(row.last_heartbeat)) < 90
}

function formatTime(ts: number) {
  const d = new Date(ts * 1000)
  return d.toLocaleString('zh-CN', { hour12: false })
}

function offlineDuration(row: any) {
  const sec = Math.floor(Date.now() / 1000 - Number(row.last_heartbeat))
  if (sec < 60) return `${sec}秒`
  if (sec < 3600) return `${Math.floor(sec / 60)}分钟`
  if (sec < 86400) return `${Math.floor(sec / 3600)}小时`
  return `${Math.floor(sec / 86400)}天`
}

function getTaskCount(agentId: number) {
  return allTasks.value.filter((t: any) => t.agent_id === agentId || t.agent_id === 0).length
}

function showTasks(row: any) {
  currentAgent.value = row
  agentTasks.value = allTasks.value.filter((t: any) => t.agent_id === row.id || t.agent_id === 0)
  taskDialogVisible.value = true
}

async function fetchData() {
  loading.value = true
  try {
    const [agentRes, taskRes]: any[] = await Promise.all([
      collectApi.listAgents(),
      collectApi.listTasks(),
    ])
    tableData.value = agentRes.data?.list || []
    allTasks.value = taskRes.data?.list || []
  } finally { loading.value = false }
}

async function handleDelete(id: number) {
  await collectApi.deleteAgent(id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(() => {
  fetchData()
  // 每30秒自动刷新
  timer = setInterval(fetchData, 30000)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.page-toolbar { margin-bottom: 16px; }
</style>
