<template>
  <div class="role-tab">
    <div class="tab-header">
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        创建角色
      </el-button>
    </div>

    <el-table :data="roles" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="角色名称" min-width="120" />
      <el-table-column prop="code" label="角色编码" min-width="120" />
      <el-table-column prop="description" label="描述" min-width="180" />
      <el-table-column prop="created_at" label="创建时间" min-width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="showEditDialog(row)">
            编辑
          </el-button>
          <el-button type="success" link @click="showAssignDrawer(row)">
            授权
          </el-button>
          <el-button type="danger" link @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑角色' : '创建角色'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" />
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
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
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
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getRoles, createRole, updateRole, deleteRole } from '@/api/role'
import RoleAssignDrawer from './RoleAssignDrawer.vue'

interface Role {
  id: number
  name: string
  code: string
  description?: string
  created_at: string
}

const props = defineProps<{
  systemId: number
}>()

const roles = ref<Role[]>([])
const loading = ref(false)
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
  description: ''
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_-]*$/, message: '编码只能包含字母、数字、下划线和连字符，且以字母开头', trigger: 'blur' }
  ]
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

async function fetchRoles() {
  loading.value = true
  try {
    const data: any = await getRoles(props.systemId)
    roles.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取角色列表失败')
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  isEditing.value = false
  editingId.value = null
  form.value = { name: '', code: '', description: '' }
  dialogVisible.value = true
}

function showEditDialog(role: Role) {
  isEditing.value = true
  editingId.value = role.id
  form.value = {
    name: role.name,
    code: role.code,
    description: role.description || ''
  }
  dialogVisible.value = true
}

function showAssignDrawer(role: Role) {
  selectedRole.value = role
  drawerVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (isEditing.value && editingId.value) {
        await updateRole(props.systemId, editingId.value, {
          name: form.value.name,
          description: form.value.description
        })
        ElMessage.success('更新成功')
      } else {
        await createRole(props.systemId, {
          name: form.value.name,
          code: form.value.code,
          description: form.value.description
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
  })
}

async function handleDelete(role: Role) {
  try {
    await ElMessageBox.confirm(
      `确定要删除角色 "${role.name}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )

    await deleteRole(props.systemId, role.id)
    ElMessage.success('删除成功')
    fetchRoles()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

watch(() => props.systemId, () => {
  fetchRoles()
})

onMounted(() => {
  fetchRoles()
})
</script>

<style scoped>
.role-tab {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.tab-header {
  margin-bottom: 16px;
}
</style>
