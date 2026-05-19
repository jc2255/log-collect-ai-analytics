<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <el-input v-model="query.username" placeholder="用户名" style="width: 160px" clearable prefix-icon="User" />
      <el-input v-model="query.ip" placeholder="IP地址" style="width: 160px" clearable prefix-icon="Location" />
      <el-date-picker v-model="query.date_range" type="daterange" range-separator="-" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width: 260px" />
      <el-button type="primary" @click="fetchData"><el-icon><Search /></el-icon>查询</el-button>
      <el-button type="danger" @click="handleClean"><el-icon><Delete /></el-icon>清空</el-button>
    </div>
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="username" label="用户名" width="120" />
      <el-table-column prop="ip" label="登录IP" width="140" />
      <el-table-column prop="location" label="登录地点" width="150" />
      <el-table-column prop="browser" label="浏览器" width="120" />
      <el-table-column prop="os" label="操作系统" width="120" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '成功' : '失败' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="msg" label="消息" />
      <el-table-column prop="created_at" label="登录时间" width="180" />
    </el-table>
    <el-pagination
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      layout="total, prev, pager, next"
      style="margin-top: 16px"
      @current-change="fetchData"
    />
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { loginLogApi } from '../../api'

const loading = ref(false)
const tableData = ref<any[]>([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 20, username: '', ip: '', date_range: null as string[] | null })

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: query.page, page_size: query.page_size, username: query.username, ip: query.ip }
    if (query.date_range) {
      params.start_time = query.date_range[0]
      params.end_time = query.date_range[1]
    }
    const res: any = await loginLogApi.list(params)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally { loading.value = false }
}

async function handleClean() {
  await ElMessageBox.confirm('确认清空所有登录日志?', '提示', { type: 'warning' })
  await loginLogApi.clean()
  ElMessage.success('已清空')
  fetchData()
}

onMounted(fetchData)
</script>


