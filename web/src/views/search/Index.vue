<template>
  <div class="search-container">
    <!-- 查询栏 -->
    <el-card class="search-bar">
      <el-form :inline="true">
        <el-form-item label="日志库">
          <el-select v-model="selectedStore" placeholder="选择日志库" style="width: 200px">
            <el-option label="全部" value="" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker v-model="timeRange" type="datetimerange" range-separator="至" start-placeholder="开始时间" end-placeholder="结束时间" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="queryStr" placeholder="输入查询条件 (KQL语法)" style="width: 400px" clearable @keyup.enter="doSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="doSearch">查询</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 查询结果 -->
    <el-card class="search-result" style="margin-top: 16px">
      <template #header>查询结果</template>
      <el-empty v-if="!logs.length" description="暂无数据，请输入查询条件" />
      <div v-else class="log-list">
        <div v-for="(log, idx) in logs" :key="idx" class="log-item">
          <pre>{{ JSON.stringify(log, null, 2) }}</pre>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const selectedStore = ref('')
const timeRange = ref<[Date, Date] | null>(null)
const queryStr = ref('')
const logs = ref<any[]>([])

function doSearch() {
  // TODO: 调用ES查询API
  console.log('Search:', queryStr.value, timeRange.value)
}
</script>

<style scoped>
.log-item {
  border-bottom: 1px solid #eee;
  padding: 8px 0;
  font-size: 12px;
}
.log-item pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
