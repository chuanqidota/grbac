<template>
  <el-drawer
    v-model="visible"
    :title="wizardMode ? '快速创建角色' : `授权 - ${role?.name}`"
    size="70%"
    @close="handleClose"
  >
    <!-- Wizard mode -->
    <template v-if="wizardMode">
      <el-steps :active="wizardStep" finish-status="success" style="margin-bottom: 24px;">
        <el-step title="基本信息" />
        <el-step title="分配成员" />
        <el-step title="分配菜单" />
        <el-step title="分配权限" />
      </el-steps>

      <!-- Step 0: basic info form -->
      <div v-show="wizardStep === 0">
        <el-form label-width="100px">
          <el-form-item label="角色名称" required>
            <el-input v-model="wizardForm.name" placeholder="请输入角色名称" autofocus />
          </el-form-item>
          <el-form-item label="角色编码" required>
            <el-input v-model="wizardForm.code" placeholder="请输入角色编码" />
          </el-form-item>
          <el-form-item label="描述">
            <el-input v-model="wizardForm.description" type="textarea" :rows="3" />
          </el-form-item>
        </el-form>
      </div>

      <!-- Step 1: assign users -->
      <div v-show="wizardStep === 1">
        <el-select
          v-model="wizardForm.userIds"
          multiple
          filterable
          remote
          :remote-method="searchUsers"
          :loading="searchingUsers"
          placeholder="搜索用户名"
          style="width: 100%"
        >
          <el-option
            v-for="u in availableUsers"
            :key="u.id"
            :label="u.chinese_name ? `${u.chinese_name}(${u.username})` : u.username"
            :value="u.id"
          />
        </el-select>
      </div>

      <!-- Step 2: assign menus -->
      <div v-show="wizardStep === 2">
        <el-tree
          ref="wizardMenuTreeRef"
          :data="menuTree"
          :props="{ label: 'name', children: 'children' }"
          show-checkbox
          node-key="id"
          v-loading="loadingMenus"
        />
      </div>

      <!-- Step 3: assign permissions -->
      <div v-show="wizardStep === 3">
        <el-transfer
          v-model="wizardForm.permissionIds"
          :data="transferPermissions"
          :titles="['未分配', '已分配']"
          filterable
          :props="{ key: 'id', label: 'label' }"
          v-loading="loadingPermissions"
        />
      </div>

      <!-- Navigation buttons -->
      <div style="display: flex; justify-content: space-between; margin-top: 24px;">
        <el-button v-if="wizardStep > 0" @click="prevWizardStep">上一步</el-button>
        <div v-else></div>
        <div>
          <el-button @click="wizardMode = false">取消</el-button>
          <el-button
            v-if="wizardStep < 3"
            type="primary"
            @click="nextWizardStep"
            :disabled="wizardStep === 0 && (!wizardForm.name || !wizardForm.code)"
          >
            下一步
          </el-button>
          <el-button
            v-else
            type="primary"
            @click="submitWizard"
            :loading="wizardSubmitting"
          >
            完成创建
          </el-button>
        </div>
      </div>
    </template>

    <!-- Normal mode (existing tabs) -->
    <template v-else>
    <el-tabs v-model="activeTab">
      <!-- Assign Users -->
      <el-tab-pane label="分配用户" name="user">
        <div class="assign-container">
          <div class="assign-section">
            <h4>已分配用户</h4>
            <el-table :data="assignedUsers" v-loading="loadingUsers" border size="small">
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column label="用户">
                <template #default="{ row }">
                  {{ row.chinese_name ? `${row.chinese_name}(${row.username})` : row.username }}
                </template>
              </el-table-column>
              <el-table-column label="操作" width="100">
                <template #default="{ row }">
                  <el-button type="danger" link @click="removeUser(row)">
                    移除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
          <div class="assign-section">
            <h4>添加用户</h4>
            <el-select
              v-model="selectedUserIds"
              multiple
              filterable
              remote
              :remote-method="searchUsers"
              :loading="searchingUsers"
              placeholder="搜索用户名"
              style="width: 100%"
            >
              <el-option
                v-for="user in availableUsers"
                :key="user.id"
                :label="user.chinese_name ? `${user.chinese_name}(${user.username})` : user.username"
                :value="user.id"
              />
            </el-select>
            <el-button
              type="primary"
              :disabled="selectedUserIds.length === 0"
              @click="assignUsers"
              style="margin-top: 12px"
            >
              添加选中用户
            </el-button>
          </div>
        </div>
      </el-tab-pane>

      <!-- Assign Menus -->
      <el-tab-pane label="分配菜单" name="menu">
        <div class="assign-container">
          <div class="assign-section">
            <h4>菜单树</h4>
            <el-tree
              ref="menuTreeRef"
              :data="menuTree"
              :props="{ label: 'name', children: 'children' }"
              show-checkbox
              node-key="id"
              v-loading="loadingMenus"
            />
          </div>
          <div class="assign-actions">
            <el-button type="primary" @click="assignMenus" :loading="savingMenus">
              保存菜单分配
            </el-button>
          </div>
        </div>
      </el-tab-pane>

      <!-- Assign Permissions -->
      <el-tab-pane label="分配接口权限" name="permission">
        <div class="assign-container">
          <div class="assign-section">
            <el-transfer
              v-model="selectedPermissionIds"
              :data="transferPermissions"
              :titles="['未分配权限', '已分配权限']"
              filterable
              filter-placeholder="搜索权限编码或名称"
              :props="{ key: 'id', label: 'label' }"
              v-loading="loadingPermissions"
            />
          </div>
          <div class="assign-actions">
            <el-button type="primary" @click="assignPermissions" :loading="savingPermissions">
              保存权限分配
            </el-button>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick, onMounted } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import {
  getRoleUsers,
  assignUsers as assignUsersApi,
  removeRoleUser,
  getRoleMenus,
  assignMenus as assignMenusApi,
  getRolePermissions,
  assignPermissions as assignPermissionsApi,
  createRole
} from '@/api/role'
import { getMenus } from '@/api/menu'
import { getPermissions } from '@/api/permission'
import { getUsers } from '@/api/user'

interface Role {
  id: number
  name: string
  code: string
}

interface User {
  id: number
  username: string
  chinese_name?: string
}

interface Menu {
  id: number
  name: string
  children?: Menu[]
}

interface Permission {
  id: number
  code: string
  name: string
  method: string
  path: string
}

const props = defineProps<{
  modelValue: boolean
  systemId: number
  role: Role | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const activeTab = ref('user')

// Users
const assignedUsers = ref<User[]>([])
const loadingUsers = ref(false)
const availableUsers = ref<User[]>([])
const searchingUsers = ref(false)
const selectedUserIds = ref<number[]>([])

// Menus
const menuTree = ref<Menu[]>([])
const assignedMenuIds = ref<number[]>([])
const loadingMenus = ref(false)
const savingMenus = ref(false)
const menuTreeRef = ref()

function getLeafNodeIds(tree: Menu[], ids: Set<number>): number[] {
  const leafIds: number[] = []
  function traverse(nodes: Menu[]) {
    for (const node of nodes) {
      if (node.children && node.children.length > 0) {
        traverse(node.children)
      } else {
        if (ids.has(node.id)) {
          leafIds.push(node.id)
        }
      }
    }
  }
  traverse(tree)
  return leafIds
}

// Permissions
const allPermissions = ref<Permission[]>([])
const loadingPermissions = ref(false)
const savingPermissions = ref(false)
const selectedPermissionIds = ref<number[]>([])

const transferPermissions = computed(() =>
  allPermissions.value.map(p => ({
    id: p.id,
    label: `${p.code} (${p.method} ${p.path})`
  }))
)

// Wizard mode
const wizardMode = ref(false)
const wizardStep = ref(0)
const wizardSubmitting = ref(false)
const wizardMenuTreeRef = ref()

const wizardForm = ref({
  name: '',
  code: '',
  description: '',
  is_default: false,
  userIds: [] as number[],
  menuIds: [] as number[],
  permissionIds: [] as number[],
})

function startWizard() {
  wizardMode.value = true
  wizardStep.value = 0
  wizardForm.value = { name: '', code: '', description: '', is_default: false, userIds: [], menuIds: [], permissionIds: [] }
  fetchMenuTree()
  fetchPermissions()
}

function nextWizardStep() {
  if (wizardStep.value < 3) wizardStep.value++
}

function prevWizardStep() {
  if (wizardStep.value > 0) wizardStep.value--
}

async function submitWizard() {
  wizardSubmitting.value = true
  try {
    const roleData: any = await createRole(props.systemId, {
      name: wizardForm.value.name,
      code: wizardForm.value.code,
      description: wizardForm.value.description,
      is_default: wizardForm.value.is_default ? 1 : 0,
    })
    const newRoleId = roleData.id
    if (wizardForm.value.userIds.length > 0) {
      await assignUsersApi(props.systemId, newRoleId, wizardForm.value.userIds)
    }
    // Get checked menu IDs from tree
    if (wizardMenuTreeRef.value) {
      const checked = wizardMenuTreeRef.value.getCheckedKeys()
      const half = wizardMenuTreeRef.value.getHalfCheckedKeys()
      wizardForm.value.menuIds = [...checked, ...half]
    }
    if (wizardForm.value.menuIds.length > 0) {
      await assignMenusApi(props.systemId, newRoleId, wizardForm.value.menuIds)
    }
    if (wizardForm.value.permissionIds.length > 0) {
      await assignPermissionsApi(props.systemId, newRoleId, wizardForm.value.permissionIds)
    }
    ElNotification.success({ title: '创建成功', message: `角色 "${wizardForm.value.name}" 已创建并完成配置` })
    wizardMode.value = false
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElNotification.error({ title: '创建失败', message: error.message || '部分步骤失败' })
  } finally {
    wizardSubmitting.value = false
  }
}

async function fetchAssignedUsers() {
  if (!props.role) return
  loadingUsers.value = true
  try {
    const data: any = await getRoleUsers(props.systemId, props.role.id)
    assignedUsers.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取已分配用户失败')
  } finally {
    loadingUsers.value = false
  }
}

async function searchUsers(query: string) {
  if (!query) {
    availableUsers.value = []
    return
  }
  searchingUsers.value = true
  try {
    const data: any = await getUsers({ page: 1, page_size: 50 })
    const users = data.list || []
    availableUsers.value = users.filter((u: User) =>
      u.username.toLowerCase().includes(query.toLowerCase()) ||
      (u.chinese_name && u.chinese_name.toLowerCase().includes(query.toLowerCase()))
    )
  } catch (error: any) {
    ElMessage.error(error.message || '搜索用户失败')
  } finally {
    searchingUsers.value = false
  }
}

async function assignUsers() {
  if (!props.role || selectedUserIds.value.length === 0) return
  try {
    const existingIds = assignedUsers.value.map(u => u.id)
    const allIds = [...new Set([...existingIds, ...selectedUserIds.value])]
    await assignUsersApi(props.systemId, props.role.id, allIds)
    ElMessage.success('用户分配成功')
    selectedUserIds.value = []
    fetchAssignedUsers()
    emit('success')
  } catch (error: any) {
    ElMessage.error(error.message || '用户分配失败')
  }
}

async function removeUser(user: User) {
  if (!props.role) return
  try {
    await removeRoleUser(props.systemId, props.role.id, user.id)
    ElMessage.success('用户移除成功')
    fetchAssignedUsers()
    emit('success')
  } catch (error: any) {
    ElMessage.error(error.message || '用户移除失败')
  }
}

async function fetchMenuTree() {
  loadingMenus.value = true
  try {
    const data: any = await getMenus(props.systemId)
    menuTree.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取菜单树失败')
  } finally {
    loadingMenus.value = false
  }
}

async function fetchAssignedMenus() {
  if (!props.role) return
  try {
    const data: any = await getRoleMenus(props.systemId, props.role.id)
    assignedMenuIds.value = (data.list || []).map((m: any) => m.id || m)
  } catch (error: any) {
    ElMessage.error(error.message || '获取已分配菜单失败')
  }
}

async function assignMenus() {
  if (!props.role || !menuTreeRef.value) return
  savingMenus.value = true
  try {
    const checkedKeys = menuTreeRef.value.getCheckedKeys()
    const halfCheckedKeys = menuTreeRef.value.getHalfCheckedKeys()
    const allKeys = [...checkedKeys, ...halfCheckedKeys]
    await assignMenusApi(props.systemId, props.role.id, allKeys)
    ElMessage.success('菜单分配成功')
    emit('success')
  } catch (error: any) {
    ElMessage.error(error.message || '菜单分配失败')
  } finally {
    savingMenus.value = false
  }
}

async function fetchPermissions() {
  loadingPermissions.value = true
  try {
    const data: any = await getPermissions(props.systemId, { page: 1, page_size: 500 })
    allPermissions.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取权限列表失败')
  } finally {
    loadingPermissions.value = false
  }
}

async function fetchAssignedPermissions() {
  if (!props.role) return
  try {
    const data: any = await getRolePermissions(props.systemId, props.role.id)
    selectedPermissionIds.value = (data.list || []).map((p: any) => p.id || p)
  } catch (error: any) {
    ElMessage.error(error.message || '获取已分配权限失败')
  }
}

async function assignPermissions() {
  if (!props.role) return
  savingPermissions.value = true
  try {
    await assignPermissionsApi(props.systemId, props.role.id, selectedPermissionIds.value)
    ElMessage.success('权限分配成功')
    emit('success')
  } catch (error: any) {
    ElMessage.error(error.message || '权限分配失败')
  } finally {
    savingPermissions.value = false
  }
}

function clearData() {
  assignedUsers.value = []
  menuTree.value = []
  assignedMenuIds.value = []
  allPermissions.value = []
  selectedPermissionIds.value = []
  selectedUserIds.value = []
  availableUsers.value = []
}

function handleClose() {
  activeTab.value = 'user'
  wizardMode.value = false
  wizardStep.value = 0
  clearData()
}

function loadAllData() {
  if (!props.role) return
  clearData()
  fetchAssignedUsers()
  fetchMenuTree()
  fetchAssignedMenus()
  fetchPermissions()
  fetchAssignedPermissions()
}

watch([menuTree, assignedMenuIds], () => {
  if (menuTree.value.length > 0 && assignedMenuIds.value.length > 0 && menuTreeRef.value) {
    const idSet = new Set(assignedMenuIds.value)
    const leafIds = getLeafNodeIds(menuTree.value, idSet)
    nextTick(() => {
      menuTreeRef.value?.setCheckedKeys(leafIds)
    })
  }
})

watch(() => props.modelValue, (val) => {
  if (val && props.role) {
    if (props.role.id === 0) {
      startWizard()
    } else {
      loadAllData()
    }
  }
})

watch(() => props.role, (newRole, oldRole) => {
  if (props.modelValue && newRole && newRole.id !== oldRole?.id) {
    loadAllData()
  }
})

onMounted(() => {
  if (props.modelValue && props.role) {
    loadAllData()
  }
})
</script>

<style scoped>
.assign-container {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.assign-section {
  flex: 1;
}

.assign-section h4 {
  margin: 0 0 var(--space-sm) 0;
  font-size: var(--font-size-base);
  font-weight: 600;
}

.assign-actions {
  margin-top: var(--space-md);
  display: flex;
  justify-content: flex-end;
}

.assign-section :deep(.el-transfer) {
  display: flex;
  align-items: center;
}

.assign-section :deep(.el-transfer-panel) {
  width: 320px;
  height: 450px;
}

.assign-section :deep(.el-transfer-panel__body) {
  height: calc(100% - 40px);
}

.assign-section :deep(.el-transfer-panel__list) {
  height: 100%;
}
</style>
