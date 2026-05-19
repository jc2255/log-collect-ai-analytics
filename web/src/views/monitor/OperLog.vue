<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <el-input v-model="query.username" placeholder="操作人" style="width: 160px" clearable prefix-icon="User" />
      <el-input v-model="query.resource" placeholder="资源路径" style="width: 220px" clearable prefix-icon="Link" />
      <el-button type="primary" @click="fetchData"><el-icon><Search /></el-icon>查询</el-button>
    </div>
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="username" label="操作人" width="120" />
      <el-table-column prop="action" label="操作" width="200" />
      <el-table-column prop="resource" label="资源" />
      <el-table-column prop="ip" label="IP" width="140" />
      <el-table-column prop="created_at" label="操作时间" width="180" />
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
import { operLogApi } from '../../api'

const loading = ref(false)
const tableData = ref<any[]>([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 20, username: '', resource: '' })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await operLogApi.list(query)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally { loading.value = false }
}

onMounted(fetchData)
</script>


