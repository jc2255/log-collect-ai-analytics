<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <div class="toolbar-filters">
        <el-select v-model="filters.repository" placeholder="仓库" clearable style="width:160px" @change="handleSearch">
          <el-option v-for="r in repoOptions" :key="r" :label="r" :value="r" />
        </el-select>
        <el-select v-model="filters.state" placeholder="状态" clearable style="width:130px" @change="handleSearch">
          <el-option label="SUCCESS" value="SUCCESS" />
          <el-option label="IN_PROGRESS" value="IN_PROGRESS" />
          <el-option label="FAILED" value="FAILED" />
        </el-select>
      </div>
      <el-button type="primary" @click="fetchData"><el-icon><Refresh /></el-icon>刷新</el-button>
    </div>
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="snapshot_name" label="快照名称" min-width="280" show-overflow-tooltip />
      <el-table-column prop="state" label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.state === 'SUCCESS' ? 'success' : row.state === 'IN_PROGRESS' ? 'warning' : 'danger'">{{ row.state }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="repository" label="仓库" width="160" show-overflow-tooltip />
      <el-table-column label="索引" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.indices_list?.join(', ') || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="start_time" label="开始时间" width="200" />
      <el-table-column prop="end_time" label="结束时间" width="200" />
      <el-table-column label="耗时" width="100">
        <template #default="{ row }">
          {{ row.duration_ms ? (row.duration_ms >= 1000 ? (row.duration_ms / 1000).toFixed(1) + 's' : row.duration_ms + 'ms') : '-' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140">
        <template #default="{ row }">
          <el-popconfirm title="确认恢复该快照?" @confirm="handleRestore(row.snapshot_name, row.repository)">
            <template #reference><el-button type="warning" link style="padding:0">恢复</el-button></template>
          </el-popconfirm>
          <el-divider direction="vertical" />
          <el-popconfirm title="确认删除该快照?" @confirm="handleDelete(row.snapshot_name, row.repository)">
            <template #reference><el-button type="danger" link style="padding:0">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSearch"
        @current-change="fetchData"
      />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { backupApi } from '../../api'

const loading = ref(false)
const tableData = ref<any[]>([])
const repoOptions = ref<string[]>([])

const filters = reactive({ repository: '', state: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await backupApi.listSnapshots({
      page: pagination.page,
      page_size: pagination.pageSize,
      repository: filters.repository || undefined,
      state: filters.state || undefined,
    })
    const data = res.data || {}
    tableData.value = data.list || []
    pagination.total = data.total || 0
  } catch {
    tableData.value = []
  } finally { loading.value = false }
}

function handleSearch() {
  pagination.page = 1
  fetchData()
}

async function handleRestore(name: string, repo: string) {
  await backupApi.restoreSnapshot(name, repo)
  ElMessage.success('恢复任务已提交')
}

async function handleDelete(name: string, repo: string) {
  await backupApi.deleteSnapshot(name, repo)
  ElMessage.success('删除成功')
  fetchData()
}

// 获取仓库列表用于筛选（从当前页数据中提取）
function extractRepos() {
  const repos = new Set<string>()
  tableData.value.forEach(r => { if (r.repository) repos.add(r.repository) })
  // 保留已有选项
  repoOptions.value.forEach(r => repos.add(r))
  repoOptions.value = Array.from(repos)
}

onMounted(async () => {
  await fetchData()
  extractRepos()
})
</script>

<style scoped>
.page-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.toolbar-filters {
  display: flex;
  gap: 12px;
}
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
