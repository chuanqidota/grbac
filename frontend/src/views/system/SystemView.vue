<template>
  <div class="system-view">
    <div class="page-header">
      <h2>系统管理</h2>
      <el-button type="primary" @click="showCreateDialog" v-if="userStore.isSuperAdmin()">
        <el-icon><Plus /></el-icon>
        创建系统
      </el-button>
    </div>

    <el-skeleton :loading="loading" animated :count="3">
      <template #template>
        <div class="bento-grid">
          <div v-for="i in 3" :key="i" class="bento-card">
            <div class="card-header">
              <el-skeleton-item variant="circle" style="width: 40px; height: 40px;" />
              <div style="flex: 1;">
                <el-skeleton-item variant="text" style="width: 60%; height: 20px;" />
                <el-skeleton-item variant="text" style="width: 40%; height: 16px; margin-top: 8px;" />
              </div>
            </div>
            <div class="card-body">
              <el-skeleton-item variant="text" style="width: 100%; height: 16px;" />
              <el-skeleton-item variant="text" style="width: 80%; height: 16px; margin-top: 8px;" />
            </div>
            <div class="card-footer">
              <el-skeleton-item variant="text" style="width: 30%; height: 32px;" />
            </div>
          </div>
        </div>
      </template>
      <template #default>
        <div class="bento-grid" v-if="systems.length > 0">
          <div
            v-for="system in systems"
            :key="system.id"
            class="bento-card"
          >
            <div class="card-header">
              <div class="card-icon">
                <el-icon><Monitor /></el-icon>
              </div>
              <div class="card-title-group">
                <h3 class="card-title">{{ system.name }}</h3>
                <el-tag size="small" type="info">{{ system.code }}</el-tag>
              </div>
            </div>

            <div class="card-body">
              <p class="card-desc">{{ system.description || '暂无描述' }}</p>
            </div>

            <div class="card-admins" v-if="system._admins && system._admins.length > 0">
              <span class="card-admins-label">管理员</span>
              <div class="card-admins-list">
                <el-tag
                  v-for="admin in system._admins"
                  :key="admin.user_id"
                  type="warning"
                  size="small"
                >
                  {{ admin.chinese_name ? `${admin.chinese_name}` : admin.username }}
                </el-tag>
              </div>
            </div>
            <div class="card-admins" v-else>
              <span class="card-admins-label">管理员</span>
              <span class="card-admins-empty">未设置</span>
            </div>

            <div class="card-footer">
              <el-button
                v-if="userStore.isSuperAdmin() || system.current_user_role === 'admin'"
                type="primary"
                link
                @click="showAdminDialog(system)"
              >
                <el-icon><User /></el-icon>
                管理员
              </el-button>
              <el-button v-if="userStore.isSuperAdmin()" type="primary" link @click="showEditDialog(system)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button v-if="userStore.isSuperAdmin()" type="danger" link @click="handleDelete(system)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </div>
        </div>
        <el-empty v-else description="暂无系统" />
      </template>
    </el-skeleton>

    <!-- Create Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑系统' : '创建系统'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="系统名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入系统名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入系统描述"
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

    <!-- Create Success Dialog (shows code) -->
    <el-dialog
      v-model="successVisible"
      title="系统创建成功"
      width="500px"
    >
      <el-descriptions :column="1" border>
        <el-descriptions-item label="系统名称">{{ createdSystem?.name }}</el-descriptions-item>
        <el-descriptions-item label="系统编码">
          <el-tag>{{ createdSystem?.code }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button type="primary" @click="successVisible = false">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- Admin Management Dialog -->
    <el-dialog
      v-model="adminDialogVisible"
      :title="`管理管理员 - ${adminSystem?.name || ''}`"
      width="600px"
    >
      <div class="admin-section">
        <h4>当前管理员</h4>
        <div v-if="adminList.length > 0" class="admin-list">
          <el-tag
            v-for="admin in adminList"
            :key="admin.user_id"
            type="warning"
            closable
            size="large"
            style="margin: 4px 8px 4px 0;"
            @close="handleRemoveAdmin(admin)"
          >
            {{ admin.chinese_name ? `${admin.chinese_name}(${admin.username})` : admin.username }}
          </el-tag>
        </div>
        <el-empty v-else description="暂无管理员" :image-size="60" />
      </div>

      <el-divider />

      <div class="admin-section">
        <h4>添加管理员</h4>
        <div style="display: flex; gap: 12px;">
          <el-select
            v-model="newAdminUserId"
            filterable
            remote
            :remote-method="searchUsersForAdmin"
            :loading="searchingAdminUsers"
            placeholder="搜索用户名"
            style="flex: 1"
          >
            <el-option
              v-for="user in availableAdminUsers"
              :key="user.id"
              :label="user.chinese_name ? `${user.chinese_name}(${user.username})` : user.username"
              :value="user.id"
            />
          </el-select>
          <el-button
            type="primary"
            @click="handleAddAdmin"
            :loading="addingAdmin"
            :disabled="!newAdminUserId"
          >
            添加
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Monitor, User, Edit, Delete } from '@element-plus/icons-vue'
import { getSystems, createSystem, updateSystem, deleteSystem, getSystemMembers, addSystemMember, removeSystemMember } from '@/api/system'
import { getUsers } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { useSystemStore } from '@/stores/system'

const userStore = useUserStore()
const systemStore = useSystemStore()

interface AdminInfo {
  user_id: number
  username: string
  chinese_name?: string
}

interface System {
  id: number
  name: string
  code: string
  description?: string
  created_at: string
  current_user_role?: string
  _admins?: AdminInfo[]
}

const systems = ref<System[]>([])
const loading = ref(false)

const dialogVisible = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref<number | null>(null)

const successVisible = ref(false)
const createdSystem = ref<{ name: string; code: string } | null>(null)

// Admin management
const adminDialogVisible = ref(false)
const adminSystem = ref<System | null>(null)
const adminList = ref<AdminInfo[]>([])
const newAdminUserId = ref<number | null>(null)
const availableAdminUsers = ref<{ id: number; username: string; chinese_name?: string }[]>([])
const searchingAdminUsers = ref(false)
const addingAdmin = ref(false)

const formRef = ref<FormInstance>()
const form = ref({
  name: '',
  description: ''
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入系统名称', trigger: 'blur' }
  ]
}

async function fetchSystems() {
  loading.value = true
  try {
    const data: any = await getSystems()
    const list = data.list || []
    // Fetch admins for all systems in parallel
    const memberPromises = list.map((sys: System) =>
      getSystemMembers(sys.id)
        .then((membersData: any) => {
          const members = membersData.list || membersData || []
          sys._admins = members.filter((m: any) => m.role === 'admin').map((m: any) => ({
            user_id: m.user_id,
            username: m.username,
            chinese_name: m.chinese_name
          }))
        })
        .catch(() => {
          sys._admins = []
        })
    )
    await Promise.all(memberPromises)
    systems.value = list
  } catch (error: any) {
    ElMessage.error(error.message || '获取系统列表失败')
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  isEditing.value = false
  editingId.value = null
  form.value = { name: '', description: '' }
  dialogVisible.value = true
}

function showEditDialog(system: System) {
  isEditing.value = true
  editingId.value = system.id
  form.value = {
    name: system.name,
    description: system.description || ''
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (isEditing.value && editingId.value) {
        await updateSystem(editingId.value, {
          name: form.value.name,
          description: form.value.description
        })
        ElMessage.success('更新成功')
      } else {
        const data: any = await createSystem({
          name: form.value.name,
          description: form.value.description
        })
        createdSystem.value = {
          name: data.name,
          code: data.code
        }
        successVisible.value = true
      }
      dialogVisible.value = false
      fetchSystems()
      systemStore.fetchSystems()
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function handleDelete(system: System) {
  try {
    await ElMessageBox.confirm(
      `确定要删除系统 "${system.name}" 吗？此操作将删除该系统下的所有角色、菜单和权限。`,
      '确认删除',
      { type: 'warning' }
    )

    await deleteSystem(system.id)
    ElMessage.success('删除成功')
    fetchSystems()
    systemStore.fetchSystems()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function showAdminDialog(system: System) {
  adminSystem.value = system
  adminList.value = system._admins || []
  newAdminUserId.value = null
  availableAdminUsers.value = []
  adminDialogVisible.value = true
}

async function searchUsersForAdmin(query: string) {
  if (!query) { availableAdminUsers.value = []; return }
  searchingAdminUsers.value = true
  try {
    const data: any = await getUsers({ page: 1, page_size: 50 })
    const users = data.list || []
    availableAdminUsers.value = users.filter((u: any) =>
      u.username.toLowerCase().includes(query.toLowerCase()) ||
      (u.chinese_name && u.chinese_name.toLowerCase().includes(query.toLowerCase()))
    )
  } catch {
    // ignore
  } finally {
    searchingAdminUsers.value = false
  }
}

async function handleAddAdmin() {
  if (!adminSystem.value || !newAdminUserId.value) return

  addingAdmin.value = true
  try {
    await addSystemMember(adminSystem.value.id, {
      user_id: newAdminUserId.value,
      role: 'admin'
    })
    ElMessage.success('管理员添加成功')
    newAdminUserId.value = null
    availableAdminUsers.value = []
    // Refresh admin list
    const membersData: any = await getSystemMembers(adminSystem.value.id)
    const members = membersData.list || membersData || []
    adminList.value = members.filter((m: any) => m.role === 'admin').map((m: any) => ({
      user_id: m.user_id,
      username: m.username,
      chinese_name: m.chinese_name
    }))
    fetchSystems()
    systemStore.fetchSystems()
  } catch (error: any) {
    ElMessage.error(error.message || '添加管理员失败')
  } finally {
    addingAdmin.value = false
  }
}

async function handleRemoveAdmin(admin: AdminInfo) {
  if (!adminSystem.value) return

  try {
    await ElMessageBox.confirm(
      `确定要移除 "${admin.username}" 的管理员权限吗？`,
      '确认移除',
      { type: 'warning' }
    )
    await removeSystemMember(adminSystem.value.id, admin.user_id)
    ElMessage.success('管理员移除成功')
    adminList.value = adminList.value.filter(a => a.user_id !== admin.user_id)
    fetchSystems()
    systemStore.fetchSystems()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '移除管理员失败')
    }
  }
}

onMounted(() => {
  fetchSystems()
})
</script>

<style scoped>
.system-view {
  padding: 0;
}

/* ========== Bento Grid Layout ========== */
.bento-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

/* ========== Bento Card — Apple Style ========== */
.bento-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-xl);
  padding: var(--space-lg);
  box-shadow: var(--shadow-md);
  border: 1px solid rgba(228, 236, 252, 0.8);
  transition: box-shadow var(--transition-base), transform var(--transition-base);
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.bento-card:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

/* ========== Card Header ========== */
.card-header {
  display: flex;
  align-items: flex-start;
  gap: var(--space-md);
}

.card-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-lg);
  background: var(--color-primary-light);
  color: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  flex-shrink: 0;
  box-shadow: var(--shadow-sm);
}

.card-title-group {
  flex: 1;
  min-width: 0;
}

.card-title {
  margin: 0 0 var(--space-xs) 0;
  font-size: var(--font-size-lg);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ========== Card Body ========== */
.card-body {
  flex: 1;
}

.card-desc {
  margin: 0;
  font-size: var(--font-size-base);
  color: var(--color-text-secondary);
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* ========== Card Admins ========== */
.card-admins {
  padding-top: var(--space-sm);
  border-top: 1px solid var(--color-border);
}

.card-admins-label {
  display: block;
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  font-weight: 500;
  margin-bottom: var(--space-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.card-admins-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.card-admins-empty {
  font-size: var(--font-size-sm);
  color: var(--color-text-placeholder);
}

/* ========== Card Footer ========== */
.card-footer {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding-top: var(--space-sm);
  border-top: 1px solid var(--color-border);
}

.card-footer .el-button {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

/* ========== Admin Dialog ========== */
.admin-section {
  margin-bottom: 16px;
}

.admin-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.admin-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

/* ========== Responsive ========== */
@media (max-width: 768px) {
  .bento-grid {
    grid-template-columns: 1fr;
  }
}

@media (min-width: 769px) and (max-width: 1024px) {
  .bento-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
