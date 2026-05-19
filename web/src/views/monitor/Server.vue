<template>
  <div class="server-page">
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card shadow="never" class="monitor-card">
          <template #header>
            <div class="card-header">
              <div class="card-header-icon cpu-icon"><el-icon><Cpu /></el-icon></div>
              <div>
                <div class="card-header-title">CPU信息</div>
                <div class="card-header-sub">Central Processing Unit</div>
              </div>
            </div>
          </template>
          <div class="monitor-items">
            <div class="monitor-item">
              <span class="monitor-label">核心数</span>
              <span class="monitor-value">{{ info.cpu?.cores || '-' }}</span>
            </div>
            <div class="monitor-item">
              <span class="monitor-label">使用率</span>
              <div class="monitor-bar-wrapper">
                <div class="monitor-bar">
                  <div class="monitor-bar-fill" :style="{ width: (info.cpu?.used_percent || 0) + '%', background: getBarColor(info.cpu?.used_percent || 0) }"></div>
                </div>
                <span class="monitor-value">{{ info.cpu?.used_percent?.toFixed(1) || 0 }}%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="monitor-card">
          <template #header>
            <div class="card-header">
              <div class="card-header-icon mem-icon"><el-icon><Coin /></el-icon></div>
              <div>
                <div class="card-header-title">内存信息</div>
                <div class="card-header-sub">Memory Usage</div>
              </div>
            </div>
          </template>
          <div class="monitor-items">
            <div class="monitor-item">
              <span class="monitor-label">总内存</span>
              <span class="monitor-value">{{ formatBytes(info.memory?.total) }}</span>
            </div>
            <div class="monitor-item">
              <span class="monitor-label">已使用</span>
              <span class="monitor-value">{{ formatBytes(info.memory?.used) }}</span>
            </div>
            <div class="monitor-item">
              <span class="monitor-label">使用率</span>
              <div class="monitor-bar-wrapper">
                <div class="monitor-bar">
                  <div class="monitor-bar-fill" :style="{ width: (info.memory?.used_percent || 0) + '%', background: getBarColor(info.memory?.used_percent || 0) }"></div>
                </div>
                <span class="monitor-value">{{ info.memory?.used_percent?.toFixed(1) || 0 }}%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card shadow="never" class="monitor-card">
          <template #header>
            <div class="card-header">
              <div class="card-header-icon disk-icon"><el-icon><Files /></el-icon></div>
              <div>
                <div class="card-header-title">磁盘信息</div>
                <div class="card-header-sub">Disk Storage</div>
              </div>
            </div>
          </template>
          <div class="monitor-items">
            <div class="monitor-item">
              <span class="monitor-label">总空间</span>
              <span class="monitor-value">{{ formatBytes(info.disk?.total) }}</span>
            </div>
            <div class="monitor-item">
              <span class="monitor-label">已使用</span>
              <span class="monitor-value">{{ formatBytes(info.disk?.used) }}</span>
            </div>
            <div class="monitor-item">
              <span class="monitor-label">使用率</span>
              <div class="monitor-bar-wrapper">
                <div class="monitor-bar">
                  <div class="monitor-bar-fill" :style="{ width: (info.disk?.used_percent || 0) + '%', background: getBarColor(info.disk?.used_percent || 0) }"></div>
                </div>
                <span class="monitor-value">{{ info.disk?.used_percent?.toFixed(1) || 0 }}%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="monitor-card">
          <template #header>
            <div class="card-header">
              <div class="card-header-icon runtime-icon"><el-icon><Monitor /></el-icon></div>
              <div>
                <div class="card-header-title">Go运行时</div>
                <div class="card-header-sub">Runtime Information</div>
              </div>
            </div>
          </template>
          <div class="monitor-items">
            <div class="monitor-item">
              <span class="monitor-label">Go版本</span>
              <span class="monitor-value code-text">{{ info.runtime?.go_version || '-' }}</span>
            </div>
            <div class="monitor-item">
              <span class="monitor-label">Goroutines</span>
              <span class="monitor-value code-text">{{ info.runtime?.goroutines || 0 }}</span>
            </div>
            <div class="monitor-item">
              <span class="monitor-label">启动时间</span>
              <span class="monitor-value code-text">{{ info.runtime?.start_time || '-' }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { monitorApi } from '../../api'

const info = ref<any>({})

function formatBytes(bytes: number) {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

function getBarColor(percent: number) {
  if (percent < 60) return 'var(--tech-success)'
  if (percent < 80) return 'var(--tech-warning)'
  return 'var(--tech-danger)'
}

async function fetchData() {
  try {
    const res: any = await monitorApi.getServerInfo()
    info.value = res.data || {}
  } catch { /* */ }
}

onMounted(fetchData)
</script>

<style scoped>
.server-page {
  min-height: 100%;
}
.monitor-card {
  height: 100%;
}
.card-header {
  display: flex;
  align-items: center;
  gap: 14px;
}
.card-header-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--tech-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}
.cpu-icon {
  background: rgba(0, 212, 255, 0.12);
  color: var(--tech-primary);
}
.mem-icon {
  background: rgba(16, 185, 129, 0.12);
  color: var(--tech-success);
}
.disk-icon {
  background: rgba(245, 158, 11, 0.12);
  color: var(--tech-warning);
}
.runtime-icon {
  background: rgba(99, 102, 241, 0.12);
  color: var(--tech-info);
}
.card-header-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--tech-text-primary);
}
.card-header-sub {
  font-size: 11px;
  color: var(--tech-text-placeholder);
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', monospace;
  margin-top: 2px;
}

.monitor-items {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.monitor-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--tech-bg-elevated);
  border-radius: var(--tech-radius-sm);
  border: 1px solid var(--tech-border);
}
.monitor-label {
  font-size: 13px;
  color: var(--tech-text-secondary);
}
.monitor-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--tech-text-primary);
}
.code-text {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', monospace;
  font-size: 13px;
}
.monitor-bar-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
}
.monitor-bar {
  width: 120px;
  height: 8px;
  background: var(--tech-bg-base);
  border-radius: 4px;
  overflow: hidden;
}
.monitor-bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.6s ease;
}
</style>
