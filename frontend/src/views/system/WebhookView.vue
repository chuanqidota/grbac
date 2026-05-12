<template>
  <div class="webhook-view">
    <div class="page-header">
      <h2>Webhook 管理</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        添加 Webhook
      </el-button>
    </div>

    <el-table :data="webhooks" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="url" label="Webhook URL" min-width="250" show-overflow-tooltip />
      <el-table-column prop="events" label="事件" min-width="200">
        <template #default="{ row }">
          <el-tag v-for="event in parseEvents(row.events)" :key="event" size="small" style="margin-right: 4px; margin-bottom: 2px;">
            {{ event }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-switch
            v-model="row.status"
            :active-value="1"
            :inactive-value="0"
            @change="handleStatusChange(row)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" min-width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="showEditDialog(row)">编辑</el-button>
          <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑 Webhook' : '创建 Webhook'"
      width="500px"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="Webhook URL" prop="url">
          <el-input v-model="form.url" placeholder="请输入 Webhook URL" />
        </el-form-item>
        <el-form-item label="触发事件" prop="events">
          <el-checkbox-group v-model="form.events">
            <el-checkbox label="role.created">角色创建</el-checkbox>
            <el-checkbox label="role.updated">角色更新</el-checkbox>
            <el-checkbox label="role.deleted">角色删除</el-checkbox>
            <el-checkbox label="menu.created">菜单创建</el-checkbox>
            <el-checkbox label="menu.updated">菜单更新</el-checkbox>
            <el-checkbox label="menu.deleted">菜单删除</el-checkbox>
            <el-checkbox label="permission.created">权限创建</el-checkbox>
            <el-checkbox label="permission.updated">权限更新</el-checkbox>
            <el-checkbox label="permission.deleted">权限删除</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getWebhooks, createWebhook, updateWebhook, deleteWebhook } from '@/api/webhook'

interface Webhook {
  id: number
  url: string
  events: string
  status: number
  created_at: string
}

const route = useRoute()
const systemId = computed(() => Number(route.params.id))

const webhooks = ref<Webhook[]>([])
const loading = ref(false)

const dialogVisible = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref<number | null>(null)

const formRef = ref<FormInstance>()
const form = ref({ url: '', events: [] as string[] })

const rules: FormRules = {
  url: [
    { required: true, message: '请输入 Webhook URL', trigger: 'blur' },
    { type: 'url', message: '请输入正确的 URL', trigger: 'blur' }
  ],
  events: [{ type: 'array', required: true, message: '请选择至少一个事件', trigger: 'change' }]
}

function parseEvents(eventsStr: string): string[] {
  if (!eventsStr) return []
  try { return JSON.parse(eventsStr) } catch { return eventsStr.split(',').map(e => e.trim()) }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

async function fetchWebhooks() {
  loading.value = true
  try {
    const data: any = await getWebhooks(systemId.value)
    webhooks.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取 Webhook 列表失败')
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  isEditing.value = false
  editingId.value = null
  form.value = { url: '', events: [] }
  dialogVisible.value = true
}

function showEditDialog(webhook: Webhook) {
  isEditing.value = true
  editingId.value = webhook.id
  form.value = { url: webhook.url, events: parseEvents(webhook.events) }
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  submitting.value = true
  try {
    const data = { url: form.value.url, events: JSON.stringify(form.value.events) }
    if (isEditing.value && editingId.value) {
      await updateWebhook(systemId.value, editingId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createWebhook(systemId.value, data)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchWebhooks()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(webhook: Webhook) {
  try {
    await ElMessageBox.confirm('确定要删除此 Webhook 吗？', '确认删除', { type: 'warning' })
    await deleteWebhook(systemId.value, webhook.id)
    ElMessage.success('删除成功')
    fetchWebhooks()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '删除失败')
  }
}

async function handleStatusChange(webhook: Webhook) {
  try {
    await updateWebhook(systemId.value, webhook.id, { status: webhook.status })
    ElMessage.success('状态更新成功')
  } catch (error: any) {
    webhook.status = webhook.status === 1 ? 0 : 1
    ElMessage.error(error.message || '状态更新失败')
  }
}

onMounted(() => { fetchWebhooks() })
</script>

<style scoped>
.webhook-view {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}
</style>
