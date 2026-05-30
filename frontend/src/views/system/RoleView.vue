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
        v-model="urlState.keyword"
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
      :data="displayRoles"
      border
      stripe
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
    >
      <el-table-column type="selection" width="50" />
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="角色名称" min-width="120" sortable="custom" />
      <el-table-column prop="code" label="角色编码" min-width="120" sortable="custom">
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
          <el-button type="danger" link @click="confirmDelete(row)" :disabled="row.is_default === 1">
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
      :before-close="handleBeforeClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入角色名称" autofocus />
        </el-form-item>
        <el-form-item label="角色编码" prop="code">
          <el-input
            v-model="formData.code"
            :disabled="isEditing"
            placeholder="请输入角色编码"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入角色描述"
          />
        </el-form-item>
        <el-form-item v-if="!isEditing" label="默认角色">
          <el-switch v-model="formData.is_default" />
          <span class="form-tip">开启后，该角色的菜单和接口权限自动对系统内所有用户生效</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeDialog">取消</el-button>
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
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { getRoles, createRole, updateRole, deleteRole } from '@/api/role'
import { useUrlState, useFormDialog, useConfirmDelete } from '@/composables'
import RoleAssignDrawer from './RoleAssignDrawer.vue'

defineOptions({ name: 'RoleView' })

interface Role {
  id: number
  name: string
  code: string
  description?: string
  is_default: number
  created_at: string
}

interface RoleForm {
  name: string
  code: string
  description: string
  is_default: boolean
}

const route = useRoute()
const systemId = computed(() => Number(route.params.id))

const roles = ref<Role[]>([])
const loading = ref(false)
const selectedIds = ref<number[]>([])

// --- URL state (search keyword) ---
const { state: urlState } = useUrlState({ keyword: '' })

// --- Fetch roles ---
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

// --- Form dialog ---
const formRef = ref<FormInstance>()
const rules: FormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_-]*$/, message: '编码只能包含字母、数字、下划线和连字符，且以字母开头', trigger: 'blur' }
  ]
}

async function submitRole(data: RoleForm & { id?: number }) {
  if (data.id) {
    await updateRole(systemId.value, data.id, {
      name: data.name,
      description: data.description
    })
    ElMessage.success('更新成功')
  } else {
    await createRole(systemId.value, {
      name: data.name,
      code: data.code,
      description: data.description,
      is_default: data.is_default ? 1 : 0
    })
    ElMessage.success('创建成功')
  }
}

const {
  visible: dialogVisible,
  isEditing,
  formData,
  formChanged,
  submitting,
  open: openDialog,
  close: closeDialog,
  submit: submitForm
} = useFormDialog<RoleForm>(submitRole, { onSuccess: fetchRoles })

function showCreateDialog() {
  openDialog({ name: '', code: '', description: '', is_default: false })
}

function showEditDialog(role: Role) {
  openDialog({
    id: role.id,
    name: role.name,
    code: role.code,
    description: role.description || '',
    is_default: role.is_default === 1
  })
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  await submitForm()
}

async function handleBeforeClose(done: () => void) {
  if (formChanged.value) {
    try {
      await ElMessageBox.confirm('表单已修改，确认放弃更改？', '提示', { type: 'warning' })
      done()
    } catch {
      // user cancelled close
    }
  } else {
    done()
  }
}

// --- Filtered roles (mutable ref for sorting) ---
function applyFilter(): Role[] {
  const kw = urlState.keyword.toLowerCase().trim()
  if (!kw) return [...roles.value]
  return roles.value.filter(r =>
    r.name.toLowerCase().includes(kw) ||
    r.code.toLowerCase().includes(kw) ||
    (r.description && r.description.toLowerCase().includes(kw))
  )
}

const displayRoles = ref<Role[]>([])
watch([roles, () => urlState.keyword], () => {
  displayRoles.value = applyFilter()
}, { immediate: true })

function handleSortChange({ prop, order }: { prop: string; order: string | null }) {
  if (!prop || !order) return
  const list = [...displayRoles.value]
  list.sort((a: any, b: any) => {
    const va = a[prop]
    const vb = b[prop]
    if (va == null) return 1
    if (vb == null) return -1
    const cmp = typeof va === 'string' ? va.localeCompare(vb) : va - vb
    return order === 'ascending' ? cmp : -cmp
  })
  displayRoles.value = list
}

// --- Delete ---
const { confirmDelete, batchDelete } = useConfirmDelete(
  (id) => deleteRole(systemId.value, id),
  { onSuccess: fetchRoles, entityName: '角色' }
)

function handleBatchDelete() {
  batchDelete(selectedIds.value, {
    onComplete: () => { selectedIds.value = [] }
  })
}

// --- Assign drawer ---
const drawerVisible = ref(false)
const selectedRole = ref<Role | null>(null)

function showAssignDrawer(role: Role) {
  selectedRole.value = role
  drawerVisible.value = true
}

function handleSelectionChange(selection: Role[]) {
  selectedIds.value = selection.map(r => r.id)
}

// --- Lifecycle ---
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
