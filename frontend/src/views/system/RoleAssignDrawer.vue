<template>
  <el-drawer
    v-model="visible"
    :title="`授权 - ${role?.name}`"
    size="70%"
    @close="handleClose"
  >
    <el-tabs v-model="activeTab">
      <!-- Assign Users -->
      <el-tab-pane label="分配用户" name="user">
        <div class="assign-container">
          <div class="assign-section">
            <h4>已分配用户</h4>
            <el-table :data="assignedUsers" v-loading="loadingUsers" border size="small">
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column prop="username" label="用户名" />
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
                :label="user.username"
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
              :default-checked-keys="assignedMenuIds"
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
            <h4>接口权限列表</h4>
            <el-table
              :data="allPermissions"
              v-loading="loadingPermissions"
              border
              size="small"
              @selection-change="handlePermissionSelectionChange"
            >
              <el-table-column type="selection" width="55" />
              <el-table-column prop="code" label="权限编码" />
              <el-table-column prop="name" label="权限名称" />
              <el-table-column prop="method" label="请求方法" width="100" />
              <el-table-column prop="path" label="请求路径" />
            </el-table>
          </div>
          <div class="assign-actions">
            <el-button type="primary" @click="assignPermissions" :loading="savingPermissions">
              保存权限分配
            </el-button>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getRoleUsers,
  assignUsers as assignUsersApi,
  removeRoleUser,
  getRoleMenus,
  assignMenus as assignMenusApi,
  getRolePermissions,
  assignPermissions as assignPermissionsApi
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

// Permissions
const allPermissions = ref<Permission[]>([])
const assignedPermissionIds = ref<number[]>([])
const loadingPermissions = ref(false)
const savingPermissions = ref(false)
const selectedPermissionIds = ref<number[]>([])

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
      u.username.toLowerCase().includes(query.toLowerCase())
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
    await assignUsersApi(props.systemId, props.role.id, selectedUserIds.value)
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
    const data: any = await getPermissions(props.systemId)
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
    assignedPermissionIds.value = (data.list || []).map((p: any) => p.id || p)
  } catch (error: any) {
    ElMessage.error(error.message || '获取已分配权限失败')
  }
}

function handlePermissionSelectionChange(selection: Permission[]) {
  selectedPermissionIds.value = selection.map(p => p.id)
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

function handleClose() {
  activeTab.value = 'user'
}

watch(() => props.modelValue, (val) => {
  if (val && props.role) {
    fetchAssignedUsers()
    fetchMenuTree()
    fetchAssignedMenus()
    fetchPermissions()
    fetchAssignedPermissions()
  }
})
</script>

<style scoped>
.assign-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.assign-section {
  flex: 1;
}

.assign-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
}

.assign-actions {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
