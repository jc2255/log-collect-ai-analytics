<template>
  <div>
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>日志总量</template>
          <div class="stat-value">{{ stats.totalLogs }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>今日日志</template>
          <div class="stat-value">{{ stats.todayLogs }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>在线Agent</template>
          <div class="stat-value">{{ stats.onlineAgents }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>活跃告警</template>
          <div class="stat-value text-danger">{{ stats.activeAlerts }}</div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'

const stats = reactive({
  totalLogs: '0',
  todayLogs: '0',
  onlineAgents: '0',
  activeAlerts: '0',
})
</script>

<style scoped>
.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
}
.text-danger {
  color: #f56c6c !important;
}
</style>
