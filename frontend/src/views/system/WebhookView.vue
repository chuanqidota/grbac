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

    <!-- Payload Demo Section -->
    <div class="demo-section">
      <el-collapse v-model="activeDemo">
        <el-collapse-item title="Payload Demo - 对接参考" name="demo">
          <p class="demo-hint">
            Webhook 触发时会向注册的 URL 发送 POST 请求，携带以下 JSON payload。接收方通过 <code>system_code</code> 识别所属系统。
          </p>
          <div v-for="group in eventDemoGroups" :key="group.category" class="demo-group">
            <h4>{{ group.label }}</h4>
            <div v-for="demo in group.events" :key="demo.event" class="demo-item">
              <div class="demo-header">
                <el-tag size="small" type="primary">{{ demo.event }}</el-tag>
                <span class="demo-desc">{{ demo.description }}</span>
              </div>
              <pre class="demo-json"><code>{{ demo.payload }}</code></pre>
            </div>
          </div>
        </el-collapse-item>
      </el-collapse>
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑 Webhook' : '创建 Webhook'"
      width="500px"
      :before-close="handleDialogClose"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="Webhook URL" prop="url">
          <el-input v-model="form.url" placeholder="请输入 Webhook URL" autofocus />
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
        <el-button @click="handleDialogClose(() => { dialogVisible = false })">取消</el-button>
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
import { formatDate } from '@/utils/format'

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
const activeDemo = ref<string[]>([])

const dialogVisible = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref<number | null>(null)

const formRef = ref<FormInstance>()
const form = ref({ url: '', events: [] as string[] })
const originalForm = ref<string>('')
const formDirty = computed(() => JSON.stringify(form.value) !== originalForm.value)

const rules: FormRules = {
  url: [
    { required: true, message: '请输入 Webhook URL', trigger: 'blur' },
    { type: 'url', message: '请输入正确的 URL', trigger: 'blur' }
  ],
  events: [{ type: 'array', required: true, message: '请选择至少一个事件', trigger: 'change' }]
}

function jsonStr(obj: any): string {
  return JSON.stringify(obj, null, 2)
}

const eventDemoGroups = computed(() => [
  {
    category: 'role',
    label: '角色事件',
    events: [
      { event: 'role.created', description: '角色创建', payload: jsonStr({ event: 'role.created', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { role_id: 1, role_name: '运维', role_code: 'ops' } }) },
      { event: 'role.updated', description: '角色更新', payload: jsonStr({ event: 'role.updated', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { role_id: 1, role_name: '运维' } }) },
      { event: 'role.deleted', description: '角色删除', payload: jsonStr({ event: 'role.deleted', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { role_id: 1, role_name: '运维' } }) },
      { event: 'role.menus_assigned', description: '分配菜单', payload: jsonStr({ event: 'role.menus_assigned', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { role_id: 1, menu_ids: [1, 2, 3] } }) },
      { event: 'role.permissions_assigned', description: '分配权限', payload: jsonStr({ event: 'role.permissions_assigned', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { role_id: 1, permission_ids: [1, 2] } }) },
      { event: 'role.users_assigned', description: '分配用户', payload: jsonStr({ event: 'role.users_assigned', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { role_id: 1, user_ids: [1, 2, 3] } }) },
      { event: 'role.user_removed', description: '移除用户', payload: jsonStr({ event: 'role.user_removed', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { role_id: 1, user_id: 2 } }) },
    ]
  },
  {
    category: 'menu',
    label: '菜单事件',
    events: [
      { event: 'menu.created', description: '菜单创建', payload: jsonStr({ event: 'menu.created', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { menu_id: 1, menu_name: '用户管理' } }) },
      { event: 'menu.updated', description: '菜单更新', payload: jsonStr({ event: 'menu.updated', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { menu_id: 1, menu_name: '用户管理' } }) },
      { event: 'menu.deleted', description: '菜单删除', payload: jsonStr({ event: 'menu.deleted', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { menu_id: 1, menu_name: '用户管理' } }) },
    ]
  },
  {
    category: 'permission',
    label: '权限事件',
    events: [
      { event: 'permission.created', description: '权限创建', payload: jsonStr({ event: 'permission.created', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { permission_id: 1, permission_code: 'user:list', path: '/api/users', method: 'GET' } }) },
      { event: 'permission.updated', description: '权限更新', payload: jsonStr({ event: 'permission.updated', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { permission_id: 1, permission_code: 'user:list', path: '/api/users', method: 'GET' } }) },
      { event: 'permission.deleted', description: '权限删除', payload: jsonStr({ event: 'permission.deleted', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { permission_id: 1, permission_code: 'user:list', path: '/api/users', method: 'GET' } }) },
    ]
  },
  {
    category: 'system',
    label: '系统事件',
    events: [
      { event: 'system.created', description: '系统创建', payload: jsonStr({ event: 'system.created', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { system_name: '示例系统' } }) },
      { event: 'system.updated', description: '系统更新', payload: jsonStr({ event: 'system.updated', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { system_name: '示例系统' } }) },
      { event: 'system.deleted', description: '系统删除', payload: jsonStr({ event: 'system.deleted', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: {} }) },
      { event: 'system.member_added', description: '添加成员', payload: jsonStr({ event: 'system.member_added', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { user_id: 1, role: 'admin' } }) },
      { event: 'system.member_removed', description: '移除成员', payload: jsonStr({ event: 'system.member_removed', timestamp: '2026-05-14T10:30:00Z', system_code: 'ABC12345', data: { user_id: 1 } }) },
    ]
  }
])

function parseEvents(eventsStr: string): string[] {
  if (!eventsStr) return []
  try { return JSON.parse(eventsStr) } catch { return eventsStr.split(',').map(e => e.trim()) }
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
  originalForm.value = JSON.stringify(form.value)
  dialogVisible.value = true
}

function showEditDialog(webhook: Webhook) {
  isEditing.value = true
  editingId.value = webhook.id
  form.value = { url: webhook.url, events: parseEvents(webhook.events) }
  originalForm.value = JSON.stringify(form.value)
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

async function handleDialogClose(done: () => void) {
  if (formDirty.value) {
    try {
      await ElMessageBox.confirm('表单已修改，确认放弃更改？', '提示', { type: 'warning' })
      done()
    } catch {
      // 用户取消关闭
    }
  } else {
    done()
  }
}

onMounted(() => { fetchWebhooks() })
</script>

<style scoped>
.webhook-view {
  padding: 0;
}
.demo-section {
  margin-top: 24px;
}
.demo-hint {
  color: var(--color-text-regular);
  font-size: 14px;
  margin-bottom: 16px;
  line-height: 1.6;
}
.demo-hint code {
  background: var(--el-fill-color-light);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
}
.demo-group {
  margin-bottom: 20px;
}
.demo-group h4 {
  margin: 0 0 12px;
  font-size: 15px;
  color: var(--color-text-primary);
}
.demo-item {
  margin-bottom: 12px;
}
.demo-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.demo-desc {
  font-size: 13px;
  color: var(--color-text-regular);
}
.demo-json {
  background: var(--el-fill-color-dark);
  border-radius: 6px;
  padding: 12px 16px;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.5;
  margin: 0;
}
.demo-json code {
  color: var(--color-text-primary);
  font-family: 'Courier New', Courier, monospace;
}
</style>
