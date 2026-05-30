<template>
  <div class="role-view">
    <div class="page-header">
      <h2>角色管理</h2>
      <div class="header-actions">
        <el-button v-if="selectedIds.length > 0" type="danger" @click="handleBatchDelete">
          批量删除 ({{ selectedIds.length }})
        </el-button>
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          创建角色
        </el-button>
      </div>
    </div>

    <div class="search-bar">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索角色名称或编码"
        clearable
        style="width: 300px"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
    </div>

    <el-skeleton :loading="loading" animated :count="5">
      <template #template>
        <el-skeleton-item variant="text" style="width: 40%; height: 32px; margin-bottom: 16px;" />
        <div v-for="i in 5" :key="i" style="display: flex; gap: 16px; margin-bottom: 12px;">
          <el-skeleton-item variant="text" style="width: 3%;" />
          <el-skeleton-item variant="text" style="width: 5%;" />
          <el-skeleton-item variant="text" style="width: 15%;" />
          <el-skeleton-item variant="text" style="width: 15%;" />
          <el-skeleton-item variant="text" style="width: 25%;" />
          <el-skeleton-item variant="text" style="width: 10%;" />
          <el-skeleton-item variant="text" style="width: 12%;" />
        </div>
      </template>
      <template #default>
    <el-table
      :data="filteredRoles"
      border
      stripe
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="50" />
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="角色名称" min-width="120" />
      <el-table-column prop="code" label="角色编码" min-width="120">
        <template #default="{ row }">
          <el-tag>{{ row.code }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.is_default === 1" type="warning">默认</el-tag>
          <el-tag v-else type="info">普通</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button type="success" link @click="showAssignDrawer(row)">
            授权
          </el-button>
          <el-button type="primary" link @click="showEditDialog(row)">
            编辑
          </el-button>
          <el-button type="danger" link @click="handleDelete(row)" :disabled="row.is_default === 1">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
      </template>
    </el-skeleton>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑角色' : '创建角色'"
      width="500px"
      :before-close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" autofocus />
        </el-form-item>
        <el-form-item label="角色编码" prop="code">
          <el-input
            v-model="form.code"
            :disabled="isEditing"
            placeholder="请输入角色编码"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入角色描述"
          />
        </el-form-item>
        <el-form-item v-if="!isEditing" label="默认角色">
          <el-switch v-model="form.is_default" />
          <span class="form-tip">开启后，该角色的菜单和接口权限自动对系统内所有用户生效</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleDialogClose(() => { dialogVisible = false })">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- Assign Drawer -->
    <RoleAssignDrawer
      v-if="selectedRole"
      v-model="drawerVisible"
      :system-id="systemId"
      :role="selectedRole"
      @success="fetchRoles"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { getRoles, createRole, updateRole, deleteRole } from '@/api/role'
import RoleAssignDrawer from './RoleAssignDrawer.vue'

interface Role {
  id: number
  name: string
  code: string
  description?: string
  is_default: number
  created_at: string
}

const route = useRoute()
const systemId = computed(() => Number(route.params.id))

const roles = ref<Role[]>([])
const loading = ref(false)
const searchKeyword = ref('')
const selectedIds = ref<number[]>([])

const dialogVisible = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref<number | null>(null)
const drawerVisible = ref(false)
const selectedRole = ref<Role | null>(null)

const formRef = ref<FormInstance>()
const form = ref({
  name: '',
  code: '',
  description: '',
  is_default: false
})
const originalForm = ref<string>('')
const formDirty = computed(() => JSON.stringify(form.value) !== originalForm.value)

const rules: FormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_-]*$/, message: '编码只能包含字母、数字、下划线和连字符，且以字母开头', trigger: 'blur' }
  ]
}

const filteredRoles = computed(() => {
  const kw = searchKeyword.value.toLowerCase().trim()
  if (!kw) return roles.value
  return roles.value.filter(r =>
    r.name.toLowerCase().includes(kw) ||
    r.code.toLowerCase().includes(kw) ||
    (r.description && r.description.toLowerCase().includes(kw))
  )
})

async function fetchRoles() {
  loading.value = true
  try {
    const data: any = await getRoles(systemId.value)
    roles.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取角色列表失败')
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(selection: Role[]) {
  selectedIds.value = selection.map(r => r.id)
}

function showCreateDialog() {
  isEditing.value = false
  editingId.value = null
  form.value = { name: '', code: '', description: '', is_default: false }
  originalForm.value = JSON.stringify(form.value)
  dialogVisible.value = true
}

function showEditDialog(role: Role) {
  isEditing.value = true
  editingId.value = role.id
  form.value = {
    name: role.name,
    code: role.code,
    description: role.description || '',
    is_default: role.is_default === 1
  }
  originalForm.value = JSON.stringify(form.value)
  dialogVisible.value = true
}

function showAssignDrawer(role: Role) {
  selectedRole.value = role
  drawerVisible.value = true
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
    if (isEditing.value && editingId.value) {
      await updateRole(systemId.value, editingId.value, {
        name: form.value.name,
        description: form.value.description
      })
      ElMessage.success('更新成功')
    } else {
      await createRole(systemId.value, {
        name: form.value.name,
        code: form.value.code,
        description: form.value.description,
        is_default: form.value.is_default ? 1 : 0
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchRoles()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(role: Role) {
  try {
    await ElMessageBox.confirm(`确定要删除角色 "${role.name}" 吗？`, '确认删除', { type: 'warning' })
    await deleteRole(systemId.value, role.id)
    ElMessage.success('删除成功')
    fetchRoles()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '删除失败')
  }
}

async function handleBatchDelete() {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.value.length} 个角色吗？`, '批量删除', { type: 'warning' })
    await Promise.all(selectedIds.value.map(id => deleteRole(systemId.value, id)))
    ElMessage.success('批量删除成功')
    selectedIds.value = []
    fetchRoles()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '批量删除失败')
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

onMounted(() => {
  fetchRoles()
})
</script>

<style scoped>
.role-view {
  padding: 0;
}
.form-tip {
  margin-left: 12px;
  font-size: 12px;
  color: var(--color-text-regular);
}
</style>
