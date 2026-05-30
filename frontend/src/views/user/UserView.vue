<template>
  <div class="user-view">
    <div class="page-header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="userDialog.open(defaultForm)">
        <el-icon><Plus /></el-icon>
        创建用户
      </el-button>
    </div>

    <div class="search-bar">
      <el-input
        v-model="urlState.keyword"
        placeholder="搜索用户名"
        clearable
        style="width: 300px"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select
        v-model="urlState.superAdmin"
        placeholder="超管筛选"
        clearable
        style="width: 150px"
        @change="handleFilterChange"
      >
        <el-option label="仅超管" :value="1" />
        <el-option label="非超管" :value="0" />
      </el-select>
    </div>

    <el-skeleton :loading="userTable.skeleton.value" animated :count="5">
      <template #template>
        <div v-for="i in 5" :key="i" style="display: flex; gap: 16px; margin-bottom: 12px;">
          <el-skeleton-item variant="text" style="width: 5%;" />
          <el-skeleton-item variant="text" style="width: 10%;" />
          <el-skeleton-item variant="text" style="width: 10%;" />
          <el-skeleton-item variant="text" style="width: 15%;" />
          <el-skeleton-item variant="text" style="width: 10%;" />
          <el-skeleton-item variant="text" style="width: 8%;" />
          <el-skeleton-item variant="text" style="width: 8%;" />
          <el-skeleton-item variant="text" style="width: 15%;" />
          <el-skeleton-item variant="text" style="width: 15%;" />
        </div>
      </template>
      <template #default>
        <el-table v-if="userTable.data.value.length > 0" :data="userTable.data.value" border stripe @sort-change="handleSortChange">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" min-width="120" sortable="custom" />
          <el-table-column prop="chinese_name" label="中文名" min-width="100" sortable="custom" />
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
          <el-table-column prop="created_at" label="创建时间" min-width="180" sortable="custom">
            <template #default="{ row }">
              {{ formatDate(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="300" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="showDetail(row)">
                详情
              </el-button>
              <el-button type="primary" link @click="userDialog.open(row)">
                编辑
              </el-button>
              <el-button type="warning" link @click="showResetPasswordDialog(row)">
                重置密码
              </el-button>
              <el-button type="danger" link @click="confirmDelete.confirmDelete({ id: row.id, name: row.username })">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无用户">
          <el-button type="primary" @click="userDialog.open(defaultForm)">创建用户</el-button>
        </el-empty>
      </template>
    </el-skeleton>

    <div class="pagination">
      <el-pagination
        v-model:current-page="userTable.page.value"
        v-model:page-size="userTable.pageSize.value"
        :page-sizes="[10, 20, 50, 100]"
        :total="userTable.total.value"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="userDialog.visible.value"
      :title="userDialog.isEditing.value ? '编辑用户' : '创建用户'"
      width="500px"
      :before-close="() => userDialog.close()"
    >
      <el-form
        ref="formRef"
        :model="userDialog.formData"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="userDialog.formData.username"
            :disabled="userDialog.isEditing.value"
            placeholder="请输入用户名（英文名）"
            :autofocus="!userDialog.isEditing.value"
          />
        </el-form-item>
        <el-form-item label="中文名" prop="chinese_name">
          <el-input
            v-model="userDialog.formData.chinese_name"
            placeholder="请输入中文名"
          />
        </el-form-item>
        <el-form-item v-if="!userDialog.isEditing.value" label="密码" prop="password">
          <el-input
            v-model="userDialog.formData.password"
            type="password"
            show-password
            placeholder="请输入密码"
          />
          <div class="form-tip">默认密码: grbac@2024</div>
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userDialog.formData.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="userDialog.formData.phone" placeholder="请输入手机号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="userDialog.close()">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="userDialog.submitting.value">
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
        <el-button @click="handleResetDialogClose(() => { resetPasswordVisible = false })">取消</el-button>
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
import { ref, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { getUsers, createUser, updateUser, deleteUser, updateUserStatus, updateUserSuperAdmin, resetUserPassword } from '@/api/user'
import { formatDate } from '@/utils/format'
import { useUrlState, useTable, useFormDialog, useConfirmDelete } from '@/composables'
import UserDetailDrawer from './UserDetailDrawer.vue'

defineOptions({ name: 'UserView' })

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

// ── URL state (keyword, page, pageSize, superAdmin) ──
const { state: urlState } = useUrlState(
  { keyword: '', page: 1, pageSize: 20, superAdmin: null as number | null },
  { debounce: 0 }
)

// ── Table ──
const userTable = useTable<User>(
  async (params) => {
    const reqParams: Record<string, any> = { ...params }
    if (urlState.superAdmin !== null) {
      reqParams.is_super_admin = urlState.superAdmin
    }
    return getUsers(reqParams) as any
  },
  { defaultPageSize: 20 }
)

// Sync URL page/pageSize when table pagination changes
watch(() => userTable.page.value, (val) => { urlState.page = val })
watch(() => userTable.pageSize.value, (val) => { urlState.pageSize = val })

// Restore page from URL on mount
nextTick(() => {
  if (urlState.page > 1) userTable.page.value = urlState.page
  if (urlState.pageSize !== 20) userTable.pageSize.value = urlState.pageSize
  userTable.refresh()
})

// Debounced keyword search
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(() => urlState.keyword, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    userTable.page.value = 1
    userTable.refresh()
  }, 300)
})

function handleSizeChange(size: number) {
  userTable.pageSize.value = size
  userTable.page.value = 1
  userTable.refresh()
}

function handleCurrentChange(page: number) {
  userTable.page.value = page
  userTable.refresh()
}

function handleFilterChange() {
  userTable.page.value = 1
  userTable.refresh()
}

function handleSortChange({ prop, order }: { prop: string; order: string | null }) {
  if (!prop || !order) return
  const list = [...userTable.data.value]
  list.sort((a: any, b: any) => {
    const va = a[prop]
    const vb = b[prop]
    if (va == null) return 1
    if (vb == null) return -1
    const cmp = typeof va === 'string' ? va.localeCompare(vb) : va - vb
    return order === 'ascending' ? cmp : -cmp
  })
  userTable.data.value = list
}

// ── Create / Edit dialog ──
const defaultForm = { username: '', chinese_name: '', password: '', email: '', phone: '' }

const formRef = ref<FormInstance>()
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

const userDialog = useFormDialog<typeof defaultForm>(
  async (data) => {
    if (userDialog.isEditing.value && userDialog.editingId.value) {
      await updateUser(userDialog.editingId.value, {
        chinese_name: data.chinese_name,
        email: data.email,
        phone: data.phone
      })
      ElMessage.success('更新成功')
    } else {
      await createUser({
        username: data.username,
        chinese_name: data.chinese_name,
        password: data.password,
        email: data.email,
        phone: data.phone
      })
      ElMessage.success('创建成功')
    }
    userTable.refresh()
  },
  {
    onError: (error) => ElMessage.error(error.message || '操作失败')
  }
)

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    await userDialog.submit()
  })
}

// ── Delete ──
const confirmDelete = useConfirmDelete(
  (id) => deleteUser(id),
  {
    entityName: '用户',
    onSuccess: () => userTable.refresh()
  }
)

// ── Detail drawer ──
const detailVisible = ref(false)
const selectedUser = ref<User | null>(null)

function showDetail(user: User) {
  selectedUser.value = user
  detailVisible.value = true
}

// ── Reset password ──
const resetPasswordVisible = ref(false)
const resetSubmitting = ref(false)
const resetPasswordUser = ref<User | null>(null)
const resetFormRef = ref<FormInstance>()
const resetForm = ref({ newPassword: '' })
const resetRules: FormRules = {
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码长度不能少于 8 位', trigger: 'blur' }
  ]
}

function showResetPasswordDialog(user: User) {
  resetPasswordUser.value = user
  resetForm.value = { newPassword: '' }
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

async function handleResetDialogClose(done: () => void) {
  if (resetForm.value.newPassword) {
    try {
      await ElMessageBox.confirm('表单已修改，确认放弃更改？', '提示', { type: 'warning' })
      done()
    } catch {
      // user cancelled
    }
  } else {
    done()
  }
}

// ── Status change ──
async function handleStatusChange(user: User) {
  try {
    await updateUserStatus(user.id, user.status)
    ElMessage.success('状态更新成功')
  } catch (error: any) {
    user.status = user.status === 1 ? 0 : 1
    ElMessage.error(error.message || '状态更新失败')
  }
}

// ── Super admin change ──
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
