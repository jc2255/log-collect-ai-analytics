<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>角色管理</span>
        <el-button type="primary">新增角色</el-button>
      </div>
    </template>
    <el-table :data="[]" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="角色名称" />
      <el-table-column prop="code" label="角色编码" />
      <el-table-column prop="description" label="描述" />
      <el-table-column label="操作" width="250">
        <template #default>
          <el-button size="small">编辑</el-button>
          <el-button size="small" type="warning">分配权限</el-button>
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
