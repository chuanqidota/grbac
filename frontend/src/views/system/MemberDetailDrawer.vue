<template>
  <el-drawer
    v-model="visible"
    :title="`授权 — ${member?.chinese_name ? `${member.chinese_name}(${member.username})` : (member?.username ?? '')}`"
    size="70%"
    @close="handleClose"
  >
    <el-tabs v-model="activeTab">
      <!-- Tab 1: Assign Roles (editable) -->
      <el-tab-pane label="分配角色" name="role">
        <div class="assign-container">
          <div class="assign-section">
            <h4>已分配角色</h4>
            <el-skeleton :loading="loadingRoles" animated :count="3">
              <template #template>
                <div v-for="i in 3" :key="i" style="display: flex; gap: 16px; margin-bottom: 12px;">
                  <el-skeleton-item variant="text" style="width: 25%;" />
                  <el-skeleton-item variant="text" style="width: 25%;" />
                  <el-skeleton-item variant="text" style="width: 30%;" />
                  <el-skeleton-item variant="text" style="width: 15%;" />
                </div>
              </template>
              <template #default>
                <el-table v-if="memberRoles.length > 0" :data="memberRoles" border size="small">
                  <el-table-column prop="name" label="角色名称" min-width="120">
                    <template #default="{ row }">
                      <el-tag type="primary">{{ row.name }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="code" label="角色编码" min-width="120">
                    <template #default="{ row }">
                      <el-tag>{{ row.code }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="description" label="描述" min-width="180" />
                  <el-table-column label="操作" width="100">
                    <template #default="{ row }">
                      <el-button type="danger" link @click="handleRemoveRole(row)">
                        移除
                      </el-button>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-else description="暂无角色分配" />
              </template>
            </el-skeleton>
          </div>
          <div class="assign-section">
            <h4>添加角色</h4>
            <el-checkbox-group v-model="selectedRoleIds">
              <el-checkbox
                v-for="role in availableRoles"
                :key="role.id"
                :label="role.id"
              >
                {{ role.name }} ({{ role.code }})
              </el-checkbox>
            </el-checkbox-group>
            <el-empty v-if="availableRoles.length === 0" description="该系统暂无可分配角色" :image-size="60" />
            <el-button
              type="primary"
              :disabled="selectedRoleIds.length === 0"
              :loading="assigning"
              @click="handleAssignRoles"
              style="margin-top: 12px"
            >
              添加选中角色
            </el-button>
          </div>
        </div>
      </el-tab-pane>

      <!-- Tab 2: Menus (read-only) -->
      <el-tab-pane label="分配菜单" name="menu">
        <div class="assign-container">
          <div class="assign-section">
            <h4>可访问菜单</h4>
            <el-tree
              :data="memberMenus"
              :props="{ label: 'name', children: 'children' }"
              node-key="id"
              default-expand-all
              v-loading="loadingMenus"
            />
            <el-empty v-if="!loadingMenus && memberMenus.length === 0" description="暂无可访问菜单" />
          </div>
        </div>
      </el-tab-pane>

      <!-- Tab 3: Permissions (read-only) -->
      <el-tab-pane label="分配接口权限" name="permission">
        <div class="assign-container">
          <div class="assign-section">
            <h4>接口权限</h4>
            <el-input
              v-model="permSearch"
              placeholder="搜索权限编码或名称"
              clearable
              style="width: 300px; margin-bottom: 12px;"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-table :data="filteredPermissions" v-loading="loadingPerms" border size="small">
              <el-table-column prop="code" label="权限编码" min-width="150" />
              <el-table-column prop="name" label="权限名称" min-width="150" />
              <el-table-column prop="method" label="请求方法" width="100">
                <template #default="{ row }">
                  <el-tag :type="methodTagType(row.method)">{{ row.method }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="path" label="路径" min-width="200" />
            </el-table>
            <el-empty v-if="!loadingPerms && memberPermissions.length === 0" description="暂无接口权限" />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getMemberRoles, getMemberMenus, getMemberPermissions } from '@/api/system'
import { getRoles, assignUsers, removeRoleUser } from '@/api/role'

interface MemberInfo {
  id: number
  username: string
  chinese_name?: string
  email?: string
  role: string
}

interface Role {
  id: number
  name: string
  code: string
  description?: string
}

interface MenuTree {
  id: number
  name: string
  children?: MenuTree[]
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
  member: MemberInfo | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const activeTab = ref('role')

// Roles
const memberRoles = ref<Role[]>([])
const loadingRoles = ref(false)
const availableRoles = ref<Role[]>([])
const selectedRoleIds = ref<number[]>([])
const assigning = ref(false)

// Menus
const memberMenus = ref<MenuTree[]>([])
const loadingMenus = ref(false)

// Permissions
const memberPermissions = ref<Permission[]>([])
const loadingPerms = ref(false)
const permSearch = ref('')

const filteredPermissions = computed(() => {
  if (!permSearch.value) return memberPermissions.value
  const q = permSearch.value.toLowerCase()
  return memberPermissions.value.filter(p =>
    p.code.toLowerCase().includes(q) || p.name.toLowerCase().includes(q)
  )
})

function methodTagType(method: string) {
  const map: Record<string, string> = { GET: 'success', POST: 'primary', PUT: 'warning', DELETE: 'danger' }
  return map[method] || 'info'
}

async function fetchMemberRoles() {
  if (!props.member) return
  loadingRoles.value = true
  try {
    const data: any = await getMemberRoles(props.systemId, props.member.id)
    memberRoles.value = Array.isArray(data) ? data : (data.list || [])
  } catch (error: any) {
    ElMessage.error(error.message || '获取成员角色失败')
  } finally {
    loadingRoles.value = false
  }
}

async function fetchMemberMenus() {
  if (!props.member) return
  loadingMenus.value = true
  try {
    const data: any = await getMemberMenus(props.systemId, props.member.id)
    memberMenus.value = Array.isArray(data) ? data : (data.list || [])
  } catch (error: any) {
    ElMessage.error(error.message || '获取成员菜单失败')
  } finally {
    loadingMenus.value = false
  }
}

async function fetchMemberPermissions() {
  if (!props.member) return
  loadingPerms.value = true
  try {
    const data: any = await getMemberPermissions(props.systemId, props.member.id)
    memberPermissions.value = Array.isArray(data) ? data : (data.list || [])
  } catch (error: any) {
    ElMessage.error(error.message || '获取成员权限失败')
  } finally {
    loadingPerms.value = false
  }
}

async function fetchAvailableRoles() {
  try {
    const data: any = await getRoles(props.systemId)
    const allRoles: Role[] = data.list || []
    const assignedIds = new Set(memberRoles.value.map(r => r.id))
    availableRoles.value = allRoles.filter(r => !assignedIds.has(r.id))
  } catch {
    availableRoles.value = []
  }
}

async function handleAssignRoles() {
  if (!props.member || selectedRoleIds.value.length === 0) return
  assigning.value = true
  try {
    await Promise.all(
      selectedRoleIds.value.map(roleId =>
        assignUsers(props.systemId, roleId, [props.member!.id])
      )
    )
    ElMessage.success('角色分配成功')
    selectedRoleIds.value = []
    fetchMemberRoles()
    fetchAvailableRoles()
    fetchMemberMenus()
    fetchMemberPermissions()
    emit('success')
  } catch (error: any) {
    ElMessage.error(error.message || '角色分配失败')
  } finally {
    assigning.value = false
  }
}

async function handleRemoveRole(role: Role) {
  if (!props.member) return
  try {
    await ElMessageBox.confirm(
      `确定要移除成员 "${props.member.chinese_name ? `${props.member.chinese_name}(${props.member.username})` : props.member.username}" 的 "${role.name}" 角色吗？`,
      '确认移除',
      { type: 'warning' }
    )
    await removeRoleUser(props.systemId, role.id, props.member.id)
    ElMessage.success('角色移除成功')
    fetchMemberRoles()
    fetchAvailableRoles()
    fetchMemberMenus()
    fetchMemberPermissions()
    emit('success')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '角色移除失败')
    }
  }
}

function clearData() {
  activeTab.value = 'role'
  memberRoles.value = []
  memberMenus.value = []
  memberPermissions.value = []
  permSearch.value = ''
  selectedRoleIds.value = []
  availableRoles.value = []
}

function loadAllData() {
  if (!props.member) return
  clearData()
  fetchMemberRoles()
  fetchAvailableRoles()
  fetchMemberMenus()
  fetchMemberPermissions()
}

function handleClose() {
  clearData()
}

watch(() => props.modelValue, (val) => {
  if (val && props.member) {
    loadAllData()
  }
})

watch(() => props.member, (newMember, oldMember) => {
  if (props.modelValue && newMember && newMember.id !== oldMember?.id) {
    loadAllData()
  }
})

onMounted(() => {
  if (props.modelValue && props.member) {
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

.assign-section h4 {
  margin: 0 0 var(--space-sm) 0;
  font-size: var(--font-size-base);
  font-weight: 600;
}
</style>
