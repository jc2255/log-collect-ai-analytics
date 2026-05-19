<template>
  <el-card shadow="never">
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="username" label="用户名" width="150" />
      <el-table-column prop="ip" label="登录IP" width="150" />
      <el-table-column prop="browser" label="浏览器" width="120" />
      <el-table-column prop="os" label="操作系统" width="120" />
      <el-table-column prop="created_at" label="登录时间" width="180" />
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-popconfirm title="确认强退该用户?" @confirm="handleForceLogout(row.id)">
            <template #reference><el-button size="small" type="danger">强退</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { onlineApi } from '../../api'

const loading = ref(false)
const tableData = ref<any[]>([])

async function fetchData() {
  loading.value = true
  try {
    const res: any = await onlineApi.list()
    tableData.value = res.data?.list || []
  } finally { loading.value = false }
}

async function handleForceLogout(id: string) {
  await onlineApi.forceLogout(id)
  ElMessage.success('强退成功')
  fetchData()
}

onMounted(fetchData)
</script>
