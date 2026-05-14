<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>Agent管理</span>
      </div>
    </template>
    <el-table :data="[]" border stripe>
      <el-table-column prop="hostname" label="主机名" />
      <el-table-column prop="ip" label="IP" />
      <el-table-column prop="os_type" label="系统" />
      <el-table-column prop="version" label="版本" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="row.status === 'online' ? 'success' : 'danger'">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="last_heartbeat" label="最后心跳" />
      <el-table-column label="操作" width="150">
        <template #default>
          <el-button size="small">采集任务</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
