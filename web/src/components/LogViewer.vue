<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const logs = ref<string[]>([])
let timer: number | undefined

onMounted(() => { timer = window.setInterval(poll, 3000) })
onUnmounted(() => { if (timer) clearInterval(timer) })

async function poll() { /* 日志通过 Go stdout 输出 */ }

function clearLogs() { logs.value = [] }
</script>

<template>
  <div class="log-container">
    <div class="log-toolbar">
      <button class="btn small" @click="clearLogs">清空日志</button>
    </div>
    <pre class="log-viewer">
      <div v-for="(line, i) in logs" :key="i">{{ line }}</div>
      <div v-if="!logs.length" class="placeholder">日志通过 Go 命令行输出查看</div>
    </pre>
  </div>
</template>

<style scoped>
.log-container { background: #1e1e2e; border-radius: 6px; overflow: hidden; }
.log-toolbar { padding: 8px 12px; background: #2d2d3f; display: flex; gap: 8px; }
.log-viewer { padding: 12px; font-family: 'Cascadia Code', 'Fira Code', monospace; font-size: 12px; color: #a0a0c0; min-height: 300px; white-space: pre-wrap; }
.placeholder { color: #555; }
.btn { padding: 4px 8px; border: 1px solid #555; border-radius: 4px; background: transparent; color: #ccc; cursor: pointer; font-size: 12px; }
</style>
