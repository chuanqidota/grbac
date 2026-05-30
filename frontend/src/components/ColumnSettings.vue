<template>
  <el-popover placement="bottom-end" :width="200" trigger="click">
    <template #reference>
      <el-button :icon="Setting" circle />
    </template>
    <div>
      <div style="display: flex; justify-content: space-between; margin-bottom: 8px; font-weight: 600;">
        <span>列设置</span>
        <el-button type="primary" link @click="resetColumns">重置</el-button>
      </div>
      <el-checkbox-group v-model="visibleColumns" @change="handleChange">
        <div v-for="col in columns" :key="col.key" style="padding: 4px 0;">
          <el-checkbox :label="col.key">{{ col.label }}</el-checkbox>
        </div>
      </el-checkbox-group>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Setting } from '@element-plus/icons-vue'

interface Column { key: string; label: string }

const props = defineProps<{ columns: Column[]; storageKey: string }>()
const emit = defineEmits<{ (e: 'change', keys: string[]): void }>()

const defaultKeys = props.columns.map(c => c.key)

function load(): string[] {
  try {
    const s = localStorage.getItem(`grbac_columns_${props.storageKey}`)
    if (s) return JSON.parse(s)
  } catch {}
  return defaultKeys
}

const visibleColumns = ref(load())

function handleChange(keys: string[]) {
  localStorage.setItem(`grbac_columns_${props.storageKey}`, JSON.stringify(keys))
  emit('change', keys)
}

function resetColumns() {
  visibleColumns.value = defaultKeys
  localStorage.removeItem(`grbac_columns_${props.storageKey}`)
  emit('change', defaultKeys)
}

emit('change', visibleColumns.value)
</script>
