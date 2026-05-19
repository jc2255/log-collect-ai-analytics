<template>
  <div class="dashboard">
    <!-- 统计卡片行 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :span="6" v-for="(item, index) in statCards" :key="index">
        <div class="stat-card" :style="{ '--card-gradient': item.gradient }">
          <div class="stat-card-bg"></div>
          <div class="stat-card-content">
            <div class="stat-icon-wrapper">
              <el-icon :size="28"><component :is="item.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ item.value }}</div>
              <div class="stat-label">{{ item.label }}</div>
            </div>
          </div>
          <div class="stat-card-deco"></div>
        </div>
      </el-col>
    </el-row>

    <!-- 图表行 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
        <el-card shadow="never" class="chart-card">
          <template #header>
            <div class="card-header">
              <span class="card-header-title">各日志库文档数量</span>
              <span class="card-header-sub">Log Store Document Distribution</span>
            </div>
          </template>
          <div ref="chartRef" style="height: 380px"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never" class="chart-card">
          <template #header>
            <div class="card-header">
              <span class="card-header-title">日志库占比</span>
              <span class="card-header-sub">Storage Distribution</span>
            </div>
          </template>
          <div ref="pieChartRef" style="height: 380px"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 底部信息行 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card shadow="never" class="info-card">
          <template #header>
            <div class="card-header">
              <span class="card-header-title">系统概览</span>
            </div>
          </template>
          <div class="system-info">
            <div class="info-item" v-for="info in systemInfo" :key="info.label">
              <span class="info-label">{{ info.label }}</span>
              <span class="info-value" :style="{ color: info.color || 'var(--tech-text-primary)' }">{{ info.value }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="info-card">
          <template #header>
            <div class="card-header">
              <span class="card-header-title">今日动态</span>
            </div>
          </template>
          <div class="system-info">
            <div class="info-item" v-for="info in todayInfo" :key="info.label">
              <span class="info-label">{{ info.label }}</span>
              <span class="info-value" :style="{ color: info.color || 'var(--tech-text-primary)' }">{{ info.value }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import * as echarts from 'echarts'
import { dashboardApi } from '../../api'

const chartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()
const stats = ref({
  logstore_count: 0,
  total_docs: 0,
  alert_count: 0,
  today_docs: 0,
  logstore_stats: [] as { name: string; doc_count: number }[],
})

const statCards = computed(() => [
  { icon: 'Folder', label: '日志库数量', value: stats.value.logstore_count, gradient: 'var(--tech-gradient-stat-1)' },
  { icon: 'Document', label: '日志总数', value: formatNumber(stats.value.total_docs), gradient: 'var(--tech-gradient-stat-2)' },
  { icon: 'Warning', label: '告警数量', value: stats.value.alert_count, gradient: 'var(--tech-gradient-stat-3)' },
  { icon: 'Connection', label: '今日日志', value: formatNumber(stats.value.today_docs), gradient: 'var(--tech-gradient-stat-4)' },
])

const systemInfo = computed(() => [
  { label: '日志库总数', value: stats.value.logstore_count, color: 'var(--tech-primary)' },
  { label: '文档总量', value: formatNumber(stats.value.total_docs), color: 'var(--tech-success)' },
  { label: '告警数量', value: stats.value.alert_count, color: 'var(--tech-warning)' },
  { label: '系统状态', value: '正常运行', color: 'var(--tech-success)' },
])

const todayInfo = computed(() => [
  { label: '今日采集量', value: formatNumber(stats.value.today_docs), color: 'var(--tech-primary)' },
  { label: '活跃日志库', value: stats.value.logstore_stats.length, color: 'var(--tech-info)' },
  { label: '采集速率', value: `${stats.value.today_docs > 0 ? Math.round(stats.value.today_docs / 24) : 0} 条/时`, color: 'var(--tech-success)' },
  { label: '系统负载', value: '正常', color: 'var(--tech-success)' },
])

function formatNumber(num: number) {
  if (!num) return '0'
  if (num >= 10000) return (num / 10000).toFixed(1) + ' 万'
  return num.toLocaleString()
}

async function fetchData() {
  try {
    const res: any = await dashboardApi.getStats()
    stats.value = res.data
    renderBarChart()
    renderPieChart()
  } catch {
    // ignore
  }
}

function renderBarChart() {
  if (!chartRef.value) return
  const chart = echarts.init(chartRef.value)
  const data = stats.value.logstore_stats || []
  chart.setOption({
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(17, 29, 49, 0.95)',
      borderColor: 'rgba(0, 212, 255, 0.2)',
      textStyle: { color: '#e2e8f0' },
      axisPointer: { type: 'shadow' },
    },
    grid: { left: '3%', right: '4%', bottom: '3%', top: '12%', containLabel: true },
    xAxis: {
      type: 'category',
      data: data.map((i) => i.name),
      axisLine: { lineStyle: { color: 'rgba(0, 212, 255, 0.1)' } },
      axisTick: { show: false },
      axisLabel: { color: '#64748b', fontSize: 12 },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: 'rgba(0, 212, 255, 0.05)' } },
      axisLabel: { color: '#64748b', fontSize: 12 },
    },
    series: [
      {
        name: '文档数',
        type: 'bar',
        data: data.map((i) => i.doc_count),
        barWidth: '50%',
        itemStyle: {
          borderRadius: [4, 4, 0, 0],
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#00d4ff' },
            { offset: 1, color: 'rgba(0, 212, 255, 0.15)' },
          ]),
        },
        emphasis: {
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: '#33ddff' },
              { offset: 1, color: 'rgba(0, 212, 255, 0.3)' },
            ]),
          },
        },
      },
    ],
  })
  window.addEventListener('resize', () => chart.resize())
}

function renderPieChart() {
  if (!pieChartRef.value) return
  const chart = echarts.init(pieChartRef.value)
  const data = stats.value.logstore_stats || []
  chart.setOption({
    tooltip: {
      trigger: 'item',
      backgroundColor: 'rgba(17, 29, 49, 0.95)',
      borderColor: 'rgba(0, 212, 255, 0.2)',
      textStyle: { color: '#e2e8f0' },
    },
    legend: {
      orient: 'horizontal',
      bottom: 0,
      textStyle: { color: '#94a3b8', fontSize: 12 },
    },
    series: [
      {
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '45%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 6,
          borderColor: '#111d31',
          borderWidth: 2,
        },
        label: { show: false },
        emphasis: {
          label: { show: true, fontSize: 14, fontWeight: 'bold', color: '#e2e8f0' },
        },
        data: data.map((i) => ({ name: i.name, value: i.doc_count })),
        color: ['#00d4ff', '#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'],
      },
    ],
  })
  window.addEventListener('resize', () => chart.resize())
}

onMounted(fetchData)
</script>

<style scoped>
.dashboard {
  min-height: 100%;
}

/* ========== 统计卡片 ========== */
.stat-row {
  margin-bottom: 0;
}
.stat-card {
  position: relative;
  border-radius: var(--tech-radius-lg);
  padding: 24px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid var(--tech-border);
  background: var(--tech-bg-card);
}
.stat-card:hover {
  border-color: var(--tech-border-active);
  box-shadow: var(--tech-shadow-glow);
  transform: translateY(-2px);
}
.stat-card-bg {
  position: absolute;
  top: 0;
  right: 0;
  width: 120px;
  height: 100%;
  background: var(--card-gradient);
  opacity: 0.06;
  border-radius: 0 var(--tech-radius-lg) var(--tech-radius-lg) 0;
}
.stat-card-content {
  display: flex;
  align-items: center;
  gap: 16px;
  position: relative;
  z-index: 1;
}
.stat-icon-wrapper {
  width: 52px;
  height: 52px;
  border-radius: var(--tech-radius-md);
  background: var(--card-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0a1628;
  flex-shrink: 0;
}
.stat-info .stat-value {
  font-size: 28px;
  font-weight: 800;
  color: var(--tech-text-primary);
  line-height: 1.2;
  letter-spacing: -0.5px;
}
.stat-info .stat-label {
  font-size: 13px;
  color: var(--tech-text-secondary);
  margin-top: 4px;
}
.stat-card-deco {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--card-gradient);
  opacity: 0.5;
}

/* ========== 图表卡片 ========== */
.chart-card {
  height: 100%;
}
.chart-card :deep(.el-card__header) {
  padding: 16px 20px;
}
.card-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
}
.card-header-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--tech-text-primary);
}
.card-header-sub {
  font-size: 12px;
  color: var(--tech-text-placeholder);
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', monospace;
}

/* ========== 信息卡片 ========== */
.info-card :deep(.el-card__header) {
  padding: 16px 20px;
}
.system-info {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.info-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 16px;
  background: var(--tech-bg-elevated);
  border-radius: var(--tech-radius-md);
  border: 1px solid var(--tech-border);
  transition: all 0.2s ease;
}
.info-item:hover {
  border-color: var(--tech-border-active);
}
.info-label {
  font-size: 12px;
  color: var(--tech-text-secondary);
}
.info-value {
  font-size: 18px;
  font-weight: 700;
}
</style>
