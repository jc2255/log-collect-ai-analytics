<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>日志库管理</span>
        <el-button type="primary">新建日志库</el-button>
      </div>
    </template>
    <el-table :data="[]" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="日志库名称" />
      <el-table-column prop="code" label="编码" />
      <el-table-column prop="retention_days" label="保留天数" />
      <el-table-column prop="api_key" label="API Key" />
      <el-table-column prop="status" label="状态" />
      <el-table-column label="操作" width="250">
        <template #default>
          <el-button size="small">字段管理</el-button>
          <el-button size="small">编辑</el-button>
          <el-button size="small" type="danger">删除</el-button>
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
