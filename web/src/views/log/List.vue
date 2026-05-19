<template>
  <div class="kibana-discover">
    <!-- 顶部工具栏 -->
    <div class="discover-toolbar">
      <el-select v-model="query.logstore" placeholder="选择日志库" style="width: 180px" @change="onStoreChange">
        <el-option v-for="s in logstores" :key="s.id" :label="s.name" :value="s.name" />
      </el-select>
      <div class="search-bar">
        <el-input
          v-model="query.keyword"
          placeholder="KQL / Lucene 搜索，如: level:ERROR AND message:timeout"
          clearable
          @keyup.enter="doSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
      <el-date-picker
        v-model="query.dateRange"
        type="datetimerange"
        range-separator="-"
        start-placeholder="开始"
        end-placeholder="结束"
        value-format="YYYY-MM-DDTHH:mm:ss"
        style="width: 360px"
        @change="doSearch"
      />
      <div class="quick-time">
        <el-button v-for="qt in quickTimes" :key="qt.label" :type="activeQuickTime === qt.label ? 'primary' : 'default'" size="small" @click="setQuickTime(qt)">
          {{ qt.label }}
        </el-button>
      </div>
      <el-button type="primary" @click="doSearch"><el-icon><Search /></el-icon>查询</el-button>
      <el-dropdown trigger="click" @command="setRefreshInterval" style="margin-left: 4px">
        <el-button :type="refreshInterval > 0 ? 'primary' : 'default'" size="small">
          <el-icon><Refresh /></el-icon>{{ refreshInterval > 0 ? `${refreshInterval}s` : '刷新' }}
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item :class="{ 'is-active': refreshInterval === 0 }" command="0">关闭</el-dropdown-item>
            <el-dropdown-item :class="{ 'is-active': refreshInterval === 5 }" command="5">5秒</el-dropdown-item>
            <el-dropdown-item :class="{ 'is-active': refreshInterval === 10 }" command="10">10秒</el-dropdown-item>
            <el-dropdown-item :class="{ 'is-active': refreshInterval === 30 }" command="30">30秒</el-dropdown-item>
            <el-dropdown-item :class="{ 'is-active': refreshInterval === 60 }" command="60">60秒</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <!-- 活动过滤条件标签 -->
    <div v-if="activeFilters.length" class="active-filters">
      <el-tag
        v-for="(f, idx) in activeFilters"
        :key="idx"
        closable
        :type="f.negate ? 'danger' : 'primary'"
        size="small"
        @close="removeFilter(idx)"
      >
        {{ f.negate ? 'NOT ' : '' }}{{ f.field }}: {{ f.value }}
      </el-tag>
    </div>

    <!-- 时间直方图 -->
    <div class="histogram-wrap">
      <div ref="histogramChart" class="histogram-chart"></div>
    </div>

    <!-- 主体区域：左侧字段面板 + 右侧日志表格 -->
    <div class="discover-body">
      <!-- 字段面板 -->
      <div class="field-panel" :class="{ collapsed: fieldPanelCollapsed }">
        <div class="panel-header" @click="fieldPanelCollapsed = !fieldPanelCollapsed">
          <span>可用字段</span>
          <el-icon><ArrowLeft v-if="!fieldPanelCollapsed" /><ArrowRight v-else /></el-icon>
        </div>
        <div v-if="!fieldPanelCollapsed" class="field-list">
          <div class="field-search">
            <el-input v-model="fieldFilter" placeholder="筛选字段" clearable size="small" />
          </div>
          <div
            v-for="f in filteredFields"
            :key="f.name"
            class="field-item"
            :class="{ selected: selectedFields.includes(f.name) }"
            @click="toggleField(f.name)"
          >
            <el-icon v-if="selectedFields.includes(f.name)" style="color: var(--tech-primary); margin-right: 4px"><Check /></el-icon>
            <span class="field-name">{{ f.name }}</span>
            <span class="field-type">{{ f.type }}</span>
          </div>
        </div>
      </div>

      <!-- 日志表格 -->
      <div class="log-panel">
        <!-- 已选字段标签栏 -->
        <div v-if="selectedFields.length" class="selected-fields-bar">
          <span class="label">已选字段:</span>
          <el-tag
            v-for="sf in selectedFields"
            :key="sf"
            size="small"
            closable
            @close="toggleField(sf)"
            style="margin-right: 4px"
          >
            {{ sf }}
          </el-tag>
        </div>

        <el-table
          :data="tableData"
          v-loading="loading"
          stripe
          size="small"
          :row-class-name="getRowClassName"
          @row-click="expandRow"
          style="width: 100%"
        >
          <el-table-column type="expand">
            <template #default="{ row }">
              <div class="log-detail">
                <div v-for="(val, key) in row" :key="key" class="log-detail-row">
                  <span class="log-detail-key" @click="addFilter(key, String(val))">{{ key }}</span>
                  <span class="log-detail-val">{{ formatVal(val) }}</span>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="时间" width="200" prop="@timestamp">
            <template #default="{ row }">
              <span class="log-time">{{ row['@timestamp'] || '' }}</span>
            </template>
          </el-table-column>
          <template v-if="selectedFields.length === 0">
            <el-table-column label="级别" width="80">
              <template #default="{ row }">
                <el-tag :type="levelTagType(row.level)" size="small" effect="dark">{{ row.level || '-' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="消息" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="log-msg">{{ row.message || row.msg || '-' }}</span>
              </template>
            </el-table-column>
          </template>
          <template v-else>
            <el-table-column
              v-for="sf in selectedFields"
              :key="sf"
              :label="sf"
              :prop="sf"
              show-overflow-tooltip
              min-width="120"
            >
              <template #default="{ row }">
                <span v-if="sf === 'level'" :class="'level-text level-' + (row.level || '').toLowerCase()">{{ row.level || '-' }}</span>
                <span v-else>{{ formatVal(row[sf]) }}</span>
              </template>
            </el-table-column>
          </template>
        </el-table>

        <div class="log-pagination">
          <span class="total-text">共 {{ total }} 条</span>
          <el-pagination
            v-model:current-page="query.page"
            v-model:page-size="query.pageSize"
            :total="total"
            :page-sizes="[50, 100, 200, 500]"
            layout="sizes, prev, pager, next"
            @current-change="fetchData"
            @size-change="fetchData"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { esLogApi, logStoreApi } from '../../api'
import * as echarts from 'echarts'

// ===== 状态 =====
const loading = ref(false)
const tableData = ref<any[]>([])
const total = ref(0)
const logstores = ref<any[]>([])
const fields = ref<{ name: string; type: string }[]>([])
const selectedFields = ref<string[]>([])
const fieldFilter = ref('')
const fieldPanelCollapsed = ref(false)
const refreshInterval = ref(0)
let refreshTimer: ReturnType<typeof setInterval> | null = null
const activeQuickTime = ref('')
const histogramChart = ref<HTMLElement>()
let chart: echarts.ECharts | null = null

const query = reactive({
  logstore: '',
  keyword: '',
  dateRange: null as string[] | null,
  page: 1,
  pageSize: 50,
})

interface ActiveFilter { field: string; value: string; negate: boolean }
const activeFilters = ref<ActiveFilter[]>([])

const quickTimes = [
  { label: '15m', minutes: 15 },
  { label: '1h', minutes: 60 },
  { label: '4h', minutes: 240 },
  { label: '24h', minutes: 1440 },
  { label: '7d', minutes: 10080 },
  { label: '30d', minutes: 43200 },
]

const filteredFields = computed(() => {
  if (!fieldFilter.value) return fields.value
  const kw = fieldFilter.value.toLowerCase()
  return fields.value.filter(f => f.name.toLowerCase().includes(kw))
})

// ===== 生命周期 =====
onMounted(async () => {
  await fetchLogstores()
  initChart()
  if (query.logstore) {
    fetchFields()
    doSearch()
  }
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  clearRefreshTimer()
  chart?.dispose()
  window.removeEventListener('resize', handleResize)
})

// ===== 方法 =====
async function fetchLogstores() {
  const res: any = await logStoreApi.list()
  const list = res.data.list || res.data || []
  logstores.value = list
  if (!query.logstore && list.length > 0) {
    query.logstore = list[0].name
  }
}

function onStoreChange() {
  selectedFields.value = []
  activeFilters.value = []
  fetchFields()
  doSearch()
}

async function fetchFields() {
  if (!query.logstore) return
  try {
    const res: any = await esLogApi.fields({ store: query.logstore })
    fields.value = res.data.fields || []
    // 按 name 排序
    fields.value.sort((a: any, b: any) => a.name.localeCompare(b.name))
  } catch {
    fields.value = []
  }
}

async function doSearch() {
  query.page = 1
  await fetchData()
  fetchHistogram()
}

async function fetchData() {
  if (!query.logstore) return
  loading.value = true
  try {
    const params: any = {
      store: query.logstore,
      keyword: buildKeyword(),
      page: query.page,
      page_size: query.pageSize,
    }
    if (query.dateRange) {
      params.start_time = query.dateRange[0]
      params.end_time = query.dateRange[1]
    }
    const res: any = await esLogApi.search(params)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

async function fetchHistogram() {
  if (!query.logstore) return
  try {
    const params: any = {
      store: query.logstore,
      keyword: buildKeyword(),
    }
    if (query.dateRange) {
      params.start_time = query.dateRange[0]
      params.end_time = query.dateRange[1]
    }
    const res: any = await esLogApi.histogram(params)
    renderHistogram(res.data.buckets || [])
  } catch {
    renderHistogram([])
  }
}

function buildKeyword(): string {
  let kw = query.keyword || ''
  for (const f of activeFilters.value) {
    const prefix = f.negate ? 'NOT ' : ''
    const cond = `${prefix}${f.field}:"${f.value}"`
    kw = kw ? `${kw} AND ${cond}` : cond
  }
  return kw
}

// 快速时间选择
function setQuickTime(qt: { label: string; minutes: number }) {
  activeQuickTime.value = qt.label
  const end = new Date()
  const start = new Date(end.getTime() - qt.minutes * 60 * 1000)
  query.dateRange = [formatDate(start), formatDate(end)]
  doSearch()
}

function formatDate(d: Date): string {
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// 自动刷新
function setRefreshInterval(val: string) {
  refreshInterval.value = parseInt(val)
  clearRefreshTimer()
  if (refreshInterval.value > 0) {
    refreshTimer = setInterval(() => {
      fetchData()
      fetchHistogram()
    }, refreshInterval.value * 1000)
  }
}

function clearRefreshTimer() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 字段操作
function toggleField(name: string) {
  const idx = selectedFields.value.indexOf(name)
  if (idx >= 0) {
    selectedFields.value.splice(idx, 1)
  } else {
    selectedFields.value.push(name)
  }
}

// 过滤器操作
function addFilter(field: string, value: string, negate = false) {
  if (field === '_id') return
  activeFilters.value.push({ field, value, negate })
  doSearch()
}

function removeFilter(idx: number) {
  activeFilters.value.splice(idx, 1)
  doSearch()
}

// 展开/折叠行
function expandRow(_row: any) {
  // el-table 内部处理 expand
}

// 直方图
function initChart() {
  nextTick(() => {
    if (histogramChart.value) {
      chart = echarts.init(histogramChart.value)
      chart.setOption({
        tooltip: { trigger: 'axis', backgroundColor: '#152238', borderColor: 'rgba(0,212,255,0.3)', textStyle: { color: '#e2e8f0' } },
        grid: { top: 10, right: 16, bottom: 28, left: 50 },
        xAxis: { type: 'category', data: [], axisLine: { lineStyle: { color: 'rgba(0,212,255,0.15)' } }, axisLabel: { color: '#94a3b8', fontSize: 11 } },
        yAxis: { type: 'value', axisLine: { show: false }, splitLine: { lineStyle: { color: 'rgba(0,212,255,0.06)' } }, axisLabel: { color: '#64748b', fontSize: 11 } },
        series: [{ type: 'bar', data: [], itemStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: '#00d4ff' }, { offset: 1, color: 'rgba(0,212,255,0.15)' }]) }, barMinHeight: 2 }],
      })
    }
  })
}

function renderHistogram(buckets: any[]) {
  if (!chart) return
  const categories = buckets.map((b: any) => {
    const s = b.key_string || ''
    return s.length > 16 ? s.substring(5, 16) : s
  })
  const values = buckets.map((b: any) => b.doc_count || 0)
  chart.setOption({
    xAxis: { data: categories },
    series: [{ data: values }],
  })
}

function handleResize() {
  chart?.resize()
}

// 格式化
function formatVal(val: any): string {
  if (val === null || val === undefined) return '-'
  if (typeof val === 'object') return JSON.stringify(val)
  return String(val)
}

function levelTagType(level?: string): string {
  const l = (level || '').toUpperCase()
  if (l === 'ERROR' || l === 'FATAL') return 'danger'
  if (l === 'WARN' || l === 'WARNING') return 'warning'
  if (l === 'INFO') return 'primary'
  if (l === 'DEBUG' || l === 'TRACE') return 'info'
  return 'info'
}

function getRowClassName({ row }: { row: any }): string {
  const l = (row.level || '').toLowerCase()
  if (l === 'error' || l === 'fatal') return 'row-error'
  if (l === 'warn' || l === 'warning') return 'row-warn'
  return ''
}
</script>

<style scoped>
.kibana-discover {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 140px);
  background: var(--tech-bg-card);
  border-radius: var(--tech-radius-lg);
  border: 1px solid var(--tech-border);
  overflow: hidden;
}

/* 顶部工具栏 */
.discover-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--tech-border);
  background: var(--tech-bg-elevated);
  flex-wrap: wrap;
}
.search-bar {
  flex: 1;
  min-width: 200px;
}
.quick-time {
  display: flex;
  gap: 2px;
}
.quick-time .el-button {
  padding: 4px 8px;
  font-size: 12px;
}

/* 活动过滤条件 */
.active-filters {
  display: flex;
  gap: 6px;
  padding: 6px 16px;
  background: rgba(0, 212, 255, 0.03);
  border-bottom: 1px solid var(--tech-border);
  flex-wrap: wrap;
}

/* 直方图 */
.histogram-wrap {
  height: 120px;
  padding: 0 16px;
  border-bottom: 1px solid var(--tech-border);
}
.histogram-chart {
  width: 100%;
  height: 100%;
}

/* 主体 */
.discover-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* 字段面板 */
.field-panel {
  width: 220px;
  border-right: 1px solid var(--tech-border);
  background: var(--tech-bg-elevated);
  display: flex;
  flex-direction: column;
  transition: width 0.2s;
  overflow: hidden;
}
.field-panel.collapsed {
  width: 36px;
}
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  cursor: pointer;
  color: var(--tech-text-regular);
  font-size: 13px;
  font-weight: 600;
  border-bottom: 1px solid var(--tech-border);
  user-select: none;
}
.panel-header:hover {
  color: var(--tech-primary);
}
.field-search {
  padding: 8px;
}
.field-list {
  flex: 1;
  overflow-y: auto;
}
.field-item {
  display: flex;
  align-items: center;
  padding: 4px 12px;
  cursor: pointer;
  font-size: 12px;
  color: var(--tech-text-regular);
  transition: all 0.15s;
}
.field-item:hover {
  background: var(--tech-bg-hover);
  color: var(--tech-primary);
}
.field-item.selected {
  color: var(--tech-primary);
  background: var(--tech-primary-bg);
}
.field-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.field-type {
  font-size: 10px;
  color: var(--tech-text-secondary);
  margin-left: 4px;
  padding: 1px 4px;
  background: var(--tech-bg-active);
  border-radius: 3px;
}

/* 日志面板 */
.log-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 已选字段标签栏 */
.selected-fields-bar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 16px;
  border-bottom: 1px solid var(--tech-border);
  background: rgba(0, 212, 255, 0.03);
}
.selected-fields-bar .label {
  font-size: 12px;
  color: var(--tech-text-secondary);
  margin-right: 4px;
}

/* 日志详情展开 */
.log-detail {
  padding: 12px 20px;
  background: var(--tech-bg-elevated);
}
.log-detail-row {
  display: flex;
  padding: 3px 0;
  font-size: 12px;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
}
.log-detail-key {
  color: var(--tech-primary);
  min-width: 200px;
  cursor: pointer;
  flex-shrink: 0;
  margin-right: 12px;
}
.log-detail-key:hover {
  text-decoration: underline;
}
.log-detail-val {
  color: var(--tech-text-primary);
  word-break: break-all;
}

/* 日志级别 */
.log-time {
  color: var(--tech-text-secondary);
  font-size: 12px;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
}
.log-msg {
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
}
.level-text {
  font-weight: 700;
  font-size: 11px;
  text-transform: uppercase;
}
.level-error, .level-fatal { color: #ef4444; }
.level-warn, .level-warning { color: #f59e0b; }
.level-info { color: #00d4ff; }
.level-debug, .level-trace { color: #64748b; }

/* 行样式 */
:deep(.row-error) td { background: rgba(239, 68, 68, 0.06) !important; }
:deep(.row-warn) td { background: rgba(245, 158, 11, 0.06) !important; }

/* 分页 */
.log-pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  border-top: 1px solid var(--tech-border);
  background: var(--tech-bg-elevated);
}
.total-text {
  font-size: 13px;
  color: var(--tech-text-secondary);
}

/* el-table 深度覆盖 */
:deep(.el-table) {
  --el-table-border-color: var(--tech-border);
  --el-table-row-hover-bg-color: var(--tech-bg-hover);
}
:deep(.el-table__expanded-cell) {
  padding: 0 !important;
}
:deep(.el-table__expand-icon) {
  color: var(--tech-text-secondary);
}
</style>
