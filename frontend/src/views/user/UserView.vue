<template>
  <div class="user-view">
    <div class="page-header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        创建用户
      </el-button>
    </div>

    <div class="search-bar">
      <el-input
        v-model="searchUsername"
        placeholder="搜索用户名"
        clearable
        style="width: 300px"
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select
        v-model="filterSuperAdmin"
        placeholder="超管筛选"
        clearable
        style="width: 150px"
        @change="handleSearch"
      >
        <el-option label="仅超管" :value="1" />
        <el-option label="非超管" :value="0" />
      </el-select>
    </div>

    <el-table :data="users" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" min-width="120" />
      <el-table-column prop="chinese_name" label="中文名" min-width="100" />
      <el-table-column prop="email" label="邮箱" min-width="180" />
      <el-table-column prop="phone" label="手机号" min-width="120" />
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
      <el-table-column label="超管" width="100">
        <template #default="{ row }">
          <el-switch
            v-model="row.is_super_admin"
            :active-value="1"
            :inactive-value="0"
            active-text="是"
            inactive-text="否"
            @change="handleSuperAdminChange(row)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" min-width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="300" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="showDetail(row)">
            详情
          </el-button>
          <el-button type="primary" link @click="showEditDialog(row)">
            编辑
          </el-button>
          <el-button type="warning" link @click="showResetPasswordDialog(row)">
            重置密码
          </el-button>
          <el-button type="danger" link @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑用户' : '创建用户'"
      width="500px"
      :before-close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            :disabled="isEditing"
            placeholder="请输入用户名（英文名）"
            autofocus
          />
        </el-form-item>
        <el-form-item label="中文名" prop="chinese_name">
          <el-input
            v-model="form.chinese_name"
            placeholder="请输入中文名"
          />
        </el-form-item>
        <el-form-item v-if="!isEditing" label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder="请输入密码"
          />
          <div class="form-tip">默认密码: grbac@2024</div>
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- Reset Password Dialog -->
    <el-dialog
      v-model="resetPasswordVisible"
      title="重置密码"
      width="450px"
      :before-close="handleResetDialogClose"
    >
      <el-form
        ref="resetFormRef"
        :model="resetForm"
        :rules="resetRules"
        label-width="100px"
      >
        <el-form-item label="用户名">
          <el-input :model-value="resetPasswordUser?.username" disabled />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="resetForm.newPassword"
            type="password"
            show-password
            placeholder="请输入新密码"
            autofocus
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPasswordVisible = false">取消</el-button>
        <el-button type="primary" @click="handleResetPassword" :loading="resetSubmitting">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- User Detail Drawer -->
    <UserDetailDrawer
      v-model="detailVisible"
      :user="selectedUser"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { getUsers, createUser, updateUser, deleteUser, updateUserStatus, updateUserSuperAdmin, resetUserPassword } from '@/api/user'
import { formatDate } from '@/utils/format'
import UserDetailDrawer from './UserDetailDrawer.vue'

interface User {
  id: number
  username: string
  chinese_name?: string
  email?: string
  phone?: string
  status: number
  is_super_admin: number
  created_at: string
}

const users = ref<User[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchUsername = ref('')
const filterSuperAdmin = ref<number | null>(null)

const dialogVisible = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref<number | null>(null)

const detailVisible = ref(false)
const selectedUser = ref<User | null>(null)

const resetPasswordVisible = ref(false)
const resetSubmitting = ref(false)
const resetPasswordUser = ref<User | null>(null)
const resetFormRef = ref<FormInstance>()
const resetForm = ref({ newPassword: '' })
const originalResetForm = ref<string>('')
const resetFormDirty = computed(() => JSON.stringify(resetForm.value) !== originalResetForm.value)
const resetRules: FormRules = {
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码长度不能少于 8 位', trigger: 'blur' }
  ]
}

const formRef = ref<FormInstance>()
const form = ref({
  username: '',
  chinese_name: '',
  password: '',
  email: '',
  phone: ''
})
const originalForm = ref<string>('')
const formDirty = computed(() => JSON.stringify(form.value) !== originalForm.value)

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3 到 50 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码长度不能少于 8 位', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ]
}

async function fetchUsers() {
  loading.value = true
  try {
    const params: Record<string, number> = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (filterSuperAdmin.value !== null) {
      params.is_super_admin = filterSuperAdmin.value
    }
    const data: any = await getUsers(params)
    users.value = data.list || []
    total.value = data.total || 0
  } catch (error: any) {
    ElMessage.error(error.message || '获取用户列表失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  fetchUsers()
}

function handleSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
  fetchUsers()
}

function handleCurrentChange(page: number) {
  currentPage.value = page
  fetchUsers()
}

function showCreateDialog() {
  isEditing.value = false
  editingId.value = null
  form.value = { username: '', chinese_name: '', password: '', email: '', phone: '' }
  originalForm.value = JSON.stringify(form.value)
  dialogVisible.value = true
}

function showDetail(user: User) {
  selectedUser.value = user
  detailVisible.value = true
}

function showEditDialog(user: User) {
  isEditing.value = true
  editingId.value = user.id
  form.value = {
    username: user.username,
    chinese_name: user.chinese_name || '',
    password: '',
    email: user.email || '',
    phone: user.phone || ''
  }
  originalForm.value = JSON.stringify(form.value)
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (isEditing.value && editingId.value) {
        await updateUser(editingId.value, {
          chinese_name: form.value.chinese_name,
          email: form.value.email,
          phone: form.value.phone
        })
        ElMessage.success('更新成功')
      } else {
        await createUser({
          username: form.value.username,
          chinese_name: form.value.chinese_name,
          password: form.value.password,
          email: form.value.email,
          phone: form.value.phone
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchUsers()
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function handleDelete(user: User) {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.username}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )

    await deleteUser(user.id)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function handleStatusChange(user: User) {
  try {
    await updateUserStatus(user.id, user.status)
    ElMessage.success('状态更新成功')
  } catch (error: any) {
    user.status = user.status === 1 ? 0 : 1
    ElMessage.error(error.message || '状态更新失败')
  }
}

async function handleSuperAdminChange(user: User) {
  try {
    await ElMessageBox.confirm(
      user.is_super_admin === 1
        ? `确定要将 "${user.username}" 设为超管吗？超管可以管理所有系统。`
        : `确定要取消 "${user.username}" 的超管权限吗？`,
      '确认操作',
      { type: 'warning' }
    )
    await updateUserSuperAdmin(user.id, user.is_super_admin)
    ElMessage.success(user.is_super_admin === 1 ? '已设为超管' : '已取消超管')
  } catch (error: any) {
    if (error !== 'cancel') {
      user.is_super_admin = user.is_super_admin === 1 ? 0 : 1
      ElMessage.error(error.message || '操作失败')
    } else {
      user.is_super_admin = user.is_super_admin === 1 ? 0 : 1
    }
  }
}

function showResetPasswordDialog(user: User) {
  resetPasswordUser.value = user
  resetForm.value = { newPassword: '' }
  originalResetForm.value = JSON.stringify(resetForm.value)
  resetPasswordVisible.value = true
}

async function handleResetPassword() {
  if (!resetFormRef.value || !resetPasswordUser.value) return

  await resetFormRef.value.validate(async (valid) => {
    if (!valid) return

    resetSubmitting.value = true
    try {
      await resetUserPassword(resetPasswordUser.value!.id, resetForm.value.newPassword)
      ElMessage.success('密码重置成功')
      resetPasswordVisible.value = false
    } catch (error: any) {
      ElMessage.error(error.message || '密码重置失败')
    } finally {
      resetSubmitting.value = false
    }
  })
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

async function handleResetDialogClose(done: () => void) {
  if (resetFormDirty.value) {
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
  fetchUsers()
})
</script>

<style scoped>
.user-view {
  padding: 0;
}

.search-bar {
  display: flex;
  gap: 12px;
  margin-bottom: var(--space-md);
}

.pagination {
  margin-top: var(--space-md);
  display: flex;
  justify-content: flex-end;
}

.form-tip {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  margin-top: 4px;
}
</style>
