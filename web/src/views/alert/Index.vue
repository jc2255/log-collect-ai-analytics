<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>告警管理</span>
        <el-button type="primary">新建告警规则</el-button>
      </div>
    </template>
    <el-alert type="info" :closable="false" description="告警引擎功能将在Phase 4实现" style="margin-bottom: 16px" />
    <el-table :data="[]" border stripe>
      <el-table-column prop="name" label="规则名称" />
      <el-table-column prop="severity" label="级别" />
      <el-table-column prop="cron_expr" label="执行周期" />
      <el-table-column prop="status" label="状态" />
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
