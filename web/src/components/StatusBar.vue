<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { getStatus } from '../api'

const status = ref({ rules_total: 0, rules_enabled: 0, orphaned_hosts: [] as string[] })
let timer: number | undefined

onMounted(async () => { await refresh(); timer = window.setInterval(refresh, 5000) })
onUnmounted(() => { if (timer) clearInterval(timer) })

async function refresh() { try { status.value = await getStatus() } catch {} }
</script>

<template>
  <header class="header">
    <div class="header-inner">
      <h1>ComeHere ⚡</h1>
      <div class="status-info">
        <span class="dot running"></span>
        <span>运行中</span>
        <span class="sep">|</span>
        <span>已启用 {{ status.rules_enabled }}/{{ status.rules_total }}</span>
        <span v-if="status.orphaned_hosts?.length" class="orphaned">⚠️ {{ status.orphaned_hosts.length }} 个残留</span>
      </div>
    </div>
  </header>
</template>

<style scoped>
.header { background: #1a1a2e; color: #fff; padding: 12px 20px; }
.header-inner { max-width: 960px; margin: 0 auto; display: flex; justify-content: space-between; align-items: center; }
h1 { font-size: 20px; }
.status-info { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; }
.dot.running { background: #4caf50; }
.sep { color: #555; }
.orphaned { color: #ff9800; }
</style>
