<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getRules, enableRule, disableRule, deleteRule, cleanupHosts } from '../api'
import type { HijackRule } from '../api'
import RuleForm from '../components/RuleForm.vue'

const rules = ref<HijackRule[]>([])
const showForm = ref(false)
const editingRule = ref<HijackRule | undefined>()

onMounted(loadRules)
async function loadRules() { rules.value = await getRules() }

async function toggleRule(r: HijackRule) {
  if (r.enabled) { await disableRule(r.id) } else { await enableRule(r.id) }
  await loadRules()
}

async function doDelete(id: string) { if (!confirm('确定删除此规则？')) return; await deleteRule(id); await loadRules() }
function editRule(r: HijackRule) { editingRule.value = r; showForm.value = true }
function addRule() { editingRule.value = undefined; showForm.value = true }
async function onSaved() { showForm.value = false; await loadRules() }
async function doCleanup() { if (!confirm('清理所有 ComeHere 添加到 hosts 的条目？')) return; await cleanupHosts(); alert('已清理') }
</script>

<template>
  <div class="toolbar">
    <button class="btn primary" @click="addRule">+ 添加劫持</button>
    <button class="btn danger" @click="doCleanup">清理残留</button>
    <router-link to="/logs" class="btn">日志</router-link>
  </div>
  <div class="notice">⚠️ 退出程序后 hosts 会残留，请先暂停所有规则再退出</div>
  <table class="rules-table" v-if="rules.length">
    <thead><tr><th>源域名</th><th>目标</th><th>状态</th><th>操作</th></tr></thead>
    <tbody>
      <tr v-for="r in rules" :key="r.id">
        <td>{{ r.source }}:{{ r.source_port }}</td>
        <td>{{ r.target }}</td>
        <td><span :class="r.enabled ? 'badge-on' : 'badge-off'">{{ r.enabled ? '运行中' : '已暂停' }}</span></td>
        <td class="actions">
          <button class="btn small" @click="toggleRule(r)">{{ r.enabled ? '暂停' : '启用' }}</button>
          <button class="btn small" @click="editRule(r)" :disabled="r.enabled">编辑</button>
          <button class="btn small danger" @click="doDelete(r.id)" :disabled="r.enabled">删除</button>
        </td>
      </tr>
    </tbody>
  </table>
  <div v-else class="empty">暂无劫持规则，点击"+ 添加劫持"创建</div>
  <RuleForm v-if="showForm" :rule="editingRule" @close="showForm=false" @saved="onSaved" />
</template>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.notice { background: #fff3cd; color: #856404; padding: 8px 12px; border-radius: 4px; font-size: 13px; margin-bottom: 12px; }
.rules-table { width: 100%; border-collapse: collapse; background: #fff; border-radius: 6px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,.1); }
.rules-table th { background: #f8f9fa; padding: 10px 12px; text-align: left; font-size: 12px; color: #666; text-transform: uppercase; }
.rules-table td { padding: 10px 12px; border-top: 1px solid #eee; font-size: 14px; }
.actions { display: flex; gap: 4px; }
.badge-on { background: #e8f5e9; color: #2e7d32; padding: 2px 8px; border-radius: 10px; font-size: 12px; }
.badge-off { background: #f5f5f5; color: #999; padding: 2px 8px; border-radius: 10px; font-size: 12px; }
.empty { text-align: center; padding: 40px; color: #999; background: #fff; border-radius: 6px; }
.btn { padding: 6px 12px; border: 1px solid #ddd; border-radius: 4px; background: #fff; cursor: pointer; font-size: 13px; }
.btn.primary { background: #1a1a2e; color: #fff; border-color: #1a1a2e; }
.btn.danger { color: #d32f2f; border-color: #d32f2f; }
.btn.small { padding: 4px 8px; font-size: 12px; }
.btn:disabled { opacity: .4; cursor: not-allowed; }
</style>
