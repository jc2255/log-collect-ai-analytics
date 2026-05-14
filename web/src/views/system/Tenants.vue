<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>租户管理</span>
        <el-button type="primary">新增租户</el-button>
      </div>
    </template>
    <el-table :data="[]" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="租户名称" />
      <el-table-column prop="code" label="租户编码" />
      <el-table-column prop="status" label="状态" />
      <el-table-column prop="description" label="描述" />
      <el-table-column label="操作" width="200">
        <template #default>
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
