<template>
  <div>
    <router-link to="/" class="back">← 返回</router-link>
    <h2 style="margin: 12px 0">代理日志</h2>
    <LogViewer />
  </div>
</template>

<script setup lang="ts">
import LogViewer from '../components/LogViewer.vue'
</script>

<style scoped>
.back { color: #1a1a2e; text-decoration: none; font-size: 14px; }
</style>
