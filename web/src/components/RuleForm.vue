<script setup lang="ts">
import { ref } from 'vue'
import { createRule, updateRule } from '../api'
import type { HijackRule } from '../api'

const props = defineProps<{ rule?: HijackRule }>()
const emit = defineEmits<{ close: []; saved: [] }>()

const form = ref({
  source: props.rule?.source || '',
  source_port: props.rule?.source_port || 443,
  target: props.rule?.target || '',
  target_tls: props.rule?.target_tls || false,
  description: props.rule?.description || '',
})

async function onSubmit() {
  try {
    if (props.rule) { await updateRule(props.rule.id, form.value) }
    else { await createRule(form.value) }
    emit('saved')
  } catch (e: any) { alert(e.message) }
}
</script>

<template>
  <div class="overlay" @click.self="emit('close')">
    <div class="modal">
      <h2>{{ rule ? '编辑劫持规则' : '添加劫持规则' }}</h2>
      <form @submit.prevent="onSubmit">
        <label>源域名 *<input v-model="form.source" required placeholder="api.openai.com" /></label>
        <label>源端口<input v-model.number="form.source_port" type="number" /></label>
        <label>目标地址 *<input v-model="form.target" required placeholder="localhost:3000" /></label>
        <label class="checkbox"><input v-model="form.target_tls" type="checkbox" /> 目标需要 HTTPS</label>
        <label>备注<input v-model="form.description" placeholder="例如: Codex → DeepSeek" /></label>
        <div class="actions">
          <button type="button" @click="emit('close')" class="btn">取消</button>
          <button type="submit" class="btn primary">确认</button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.overlay { position: fixed; inset: 0; background: rgba(0,0,0,.4); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal { background: #fff; border-radius: 8px; padding: 24px; width: 420px; max-width: 90vw; }
h2 { margin-bottom: 16px; font-size: 16px; }
form { display: flex; flex-direction: column; gap: 12px; }
label { display: flex; flex-direction: column; gap: 4px; font-size: 13px; color: #555; }
label.checkbox { flex-direction: row; align-items: center; }
input { padding: 8px 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
input[type=checkbox] { margin-right: 6px; }
.actions { display: flex; gap: 8px; justify-content: flex-end; margin-top: 8px; }
</style>
