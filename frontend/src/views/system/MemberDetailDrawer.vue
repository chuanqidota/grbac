<template>
  <el-drawer
    v-model="visible"
    :title="`成员详情 — ${member?.username ?? ''}`"
    size="60%"
    @close="handleClose"
  >
    <template v-if="member">
      <!-- Section 1: Basic Info -->
      <el-descriptions :column="2" border style="margin-bottom: 24px;">
        <el-descriptions-item label="ID">{{ member.id }}</el-descriptions-item>
        <el-descriptions-item label="用户名">{{ member.username }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ member.email || '-' }}</el-descriptions-item>
        <el-descriptions-item label="系统角色">
          <el-tag :type="member.role === 'admin' ? 'danger' : 'primary'">
            {{ member.role === 'admin' ? '管理员' : '成员' }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <!-- Section 2: RBAC Roles -->
      <div class="section-header">
        <h4>RBAC 角色</h4>
        <el-button type="primary" size="small" @click="showAssignRoleDialog">
          分配角色
        </el-button>
      </div>

      <el-table :data="memberRoles" v-loading="loadingRoles" border stripe style="margin-bottom: 24px;">
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
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" link @click="handleRemoveRole(row)">
              移除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loadingRoles && memberRoles.length === 0" description="暂无 RBAC 角色分配" />

      <!-- Section 3: Effective Menus -->
      <div class="section-header">
        <h4>可访问菜单</h4>
      </div>

      <el-tree
        :data="memberMenus"
        :props="{ label: 'name', children: 'children' }"
        node-key="id"
        show-checkbox
        default-expand-all
        v-loading="loadingMenus"
        style="margin-bottom: 24px;"
      />

      <el-empty v-if="!loadingMenus && memberMenus.length === 0" description="暂无可访问菜单" />

      <!-- Section 4: Effective Permissions -->
      <div class="section-header">
        <h4>接口权限</h4>
      </div>

      <div style="margin-bottom: 12px;">
        <el-input
          v-model="permSearch"
          placeholder="搜索权限编码或名称"
          clearable
          style="width: 300px;"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>

      <el-table :data="filteredPermissions" v-loading="loadingPerms" border stripe>
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
    </template>

    <!-- Assign Role Sub-Dialog -->
    <el-dialog
      v-model="assignRoleVisible"
      title="分配角色"
      width="400px"
      append-to-body
    >
      <el-checkbox-group v-model="selectedRoleIds">
        <el-checkbox
          v-for="role in availableRoles"
          :key="role.id"
          :label="role.id"
        >
          {{ role.name }} ({{ role.code }})
        </el-checkbox>
      </el-checkbox-group>
      <el-empty v-if="availableRoles.length === 0" description="该系统暂无角色" :image-size="60" />
      <template #footer>
        <el-button @click="assignRoleVisible = false">取消</el-button>
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
import { Search } from '@element-plus/icons-vue'
import { getMemberRoles, getMemberMenus, getMemberPermissions } from '@/api/system'
import { getRoles, assignUsers, removeRoleUser } from '@/api/role'

interface MemberInfo {
  id: number
  username: string
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

// Roles
const memberRoles = ref<Role[]>([])
const loadingRoles = ref(false)
const assignRoleVisible = ref(false)
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
    // Exclude roles the member already has
    const assignedIds = new Set(memberRoles.value.map(r => r.id))
    availableRoles.value = allRoles.filter(r => !assignedIds.has(r.id))
  } catch {
    availableRoles.value = []
  }
}

function showAssignRoleDialog() {
  selectedRoleIds.value = []
  fetchAvailableRoles()
  assignRoleVisible.value = true
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
    assignRoleVisible.value = false
    fetchMemberRoles()
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
      `确定要移除成员 "${props.member.username}" 的 "${role.name}" 角色吗？`,
      '确认移除',
      { type: 'warning' }
    )
    await removeRoleUser(props.systemId, role.id, props.member.id)
    ElMessage.success('角色移除成功')
    fetchMemberRoles()
    emit('success')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '角色移除失败')
    }
  }
}

function clearData() {
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
