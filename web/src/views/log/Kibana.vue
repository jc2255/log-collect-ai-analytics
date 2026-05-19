<template>
  <el-card shadow="never">
    <div class="kibana-container">
      <el-alert v-if="!kibanaUrl" title="请在系统配置中设置 Kibana 地址" type="warning" :closable="false" />
      <iframe v-else :src="kibanaUrl" frameborder="0" class="kibana-frame"></iframe>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

// Kibana URL from env or default
const kibanaUrl = ref(import.meta.env.VITE_KIBANA_URL || 'http://localhost:5601')
</script>

<style scoped>
.kibana-container {
  width: 100%;
  height: calc(100vh - 180px);
  background: var(--tech-bg-elevated);
  border-radius: var(--tech-radius-md);
  overflow: hidden;
}
.kibana-frame {
  width: 100%;
  height: 100%;
  border: none;
}
</style>
