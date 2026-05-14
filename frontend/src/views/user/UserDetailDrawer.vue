<template>
  <el-drawer
    v-model="visible"
    :title="`用户详情 — ${user?.chinese_name ? `${user.chinese_name}(${user.username})` : (user?.username ?? '')}`"
    size="60%"
    @close="handleClose"
  >
    <template v-if="user">
      <!-- User Info -->
      <el-descriptions :column="2" border style="margin-bottom: 24px;">
        <el-descriptions-item label="ID">{{ user.id }}</el-descriptions-item>
        <el-descriptions-item label="用户名">{{ user.username }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ user.email || '-' }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ user.phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="user.status === 1 ? 'success' : 'danger'">
            {{ user.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(user.created_at) }}</el-descriptions-item>
      </el-descriptions>

      <!-- Roles Section -->
      <div class="section-header">
        <h4>系统角色</h4>
        <el-button type="primary" size="small" @click="showAddRoleDialog">
          添加角色
        </el-button>
      </div>

      <el-table :data="userRoles" v-loading="loadingRoles" border stripe>
        <el-table-column prop="system_name" label="系统" min-width="120" />
        <el-table-column prop="role_name" label="角色" min-width="120">
          <template #default="{ row }">
            <el-tag type="primary">{{ row.role_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="role_code" label="角色编码" min-width="120">
          <template #default="{ row }">
            <el-tag>{{ row.role_code }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" link @click="handleRemoveRole(row)">
              移除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loadingRoles && userRoles.length === 0" description="暂无角色分配" />
    </template>

    <!-- Add Role Dialog -->
    <el-dialog
      v-model="addRoleVisible"
      title="添加角色"
      width="500px"
      append-to-body
    >
      <el-form label-width="80px">
        <el-form-item label="系统">
          <el-select
            v-model="selectedSystemId"
            placeholder="选择系统"
            style="width: 100%"
            @change="fetchSystemRoles"
          >
            <el-option
              v-for="sys in systems"
              :key="sys.id"
              :label="sys.name"
              :value="sys.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="角色" v-if="selectedSystemId">
          <el-checkbox-group v-model="selectedRoleIds">
            <el-checkbox
              v-for="role in systemRoles"
              :key="role.id"
              :label="role.id"
            >
              {{ role.name }} ({{ role.code }})
            </el-checkbox>
          </el-checkbox-group>
          <el-empty v-if="systemRoles.length === 0" description="该系统暂无角色" :image-size="60" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addRoleVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleAssignRoles"
          :loading="assigning"
          :disabled="selectedRoleIds.length === 0"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserRoles } from '@/api/user'
import { getSystems } from '@/api/system'
import { getRoles, assignUsers, removeRoleUser } from '@/api/role'

interface UserInfo {
  id: number
  username: string
  chinese_name?: string
  email?: string
  phone?: string
  status: number
  created_at: string
}

interface UserRoleInfo {
  system_id: number
  system_name: string
  role_id: number
  role_name: string
  role_code: string
}

interface System {
  id: number
  name: string
}

interface Role {
  id: number
  name: string
  code: string
}

const props = defineProps<{
  modelValue: boolean
  user: UserInfo | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const userRoles = ref<UserRoleInfo[]>([])
const loadingRoles = ref(false)

const addRoleVisible = ref(false)
const systems = ref<System[]>([])
const systemRoles = ref<Role[]>([])
const selectedSystemId = ref<number | null>(null)
const selectedRoleIds = ref<number[]>([])
const assigning = ref(false)

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

async function fetchUserRoles() {
  if (!props.user) return
  loadingRoles.value = true
  try {
    const data: any = await getUserRoles(props.user.id)
    userRoles.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取用户角色失败')
  } finally {
    loadingRoles.value = false
  }
}

async function fetchSystems() {
  try {
    const data: any = await getSystems({ page: 1, page_size: 100 })
    systems.value = data.list || []
  } catch {
    // ignore
  }
}

async function fetchSystemRoles(systemId: number) {
  try {
    const data: any = await getRoles(systemId)
    systemRoles.value = data.list || []
    selectedRoleIds.value = []
  } catch {
    systemRoles.value = []
  }
}

function showAddRoleDialog() {
  selectedSystemId.value = null
  selectedRoleIds.value = []
  systemRoles.value = []
  addRoleVisible.value = true
  fetchSystems()
}

async function handleAssignRoles() {
  if (!props.user || !selectedSystemId.value || selectedRoleIds.value.length === 0) return
  assigning.value = true
  try {
    await Promise.all(
      selectedRoleIds.value.map(roleId =>
        assignUsers(selectedSystemId.value!, roleId, [props.user!.id])
      )
    )
    ElMessage.success('角色分配成功')
    addRoleVisible.value = false
    fetchUserRoles()
  } catch (error: any) {
    ElMessage.error(error.message || '角色分配失败')
  } finally {
    assigning.value = false
  }
}

async function handleRemoveRole(role: UserRoleInfo) {
  if (!props.user) return
  try {
    await ElMessageBox.confirm(
      `确定要移除用户 "${props.user.chinese_name ? `${props.user.chinese_name}(${props.user.username})` : props.user.username}" 在 "${role.system_name}" 系统中的 "${role.role_name}" 角色吗？`,
      '确认移除',
      { type: 'warning' }
    )
    await removeRoleUser(role.system_id, role.role_id, props.user.id)
    ElMessage.success('角色移除成功')
    fetchUserRoles()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '角色移除失败')
    }
  }
}

function handleClose() {
  userRoles.value = []
}

watch(() => props.modelValue, (val) => {
  if (val && props.user) {
    fetchUserRoles()
  }
})
</script>

<style scoped>
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-md);
}

.section-header h4 {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: 600;
}
</style>
