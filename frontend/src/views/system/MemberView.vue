<template>
  <div class="member-view">
    <div class="page-header">
      <h2>成员管理</h2>
      <el-button type="primary" @click="showAssignDialog">
        <el-icon><Plus /></el-icon>
        分配角色
      </el-button>
    </div>

    <el-table :data="members" v-loading="loading" border stripe>
      <el-table-column prop="user_id" label="ID" width="80" />
      <el-table-column label="用户" min-width="160">
        <template #default="{ row }">
          {{ row.chinese_name ? `${row.chinese_name}(${row.username})` : row.username }}
        </template>
      </el-table-column>
      <el-table-column prop="email" label="邮箱" min-width="180" />
      <el-table-column label="已分配角色" min-width="200">
        <template #default="{ row }">
          <el-tag
            v-for="role in row.roles"
            :key="role.id"
            style="margin: 2px 4px 2px 0;"
            closable
            @close="handleRemoveRole(row, role)"
          >
            {{ role.name }}
          </el-tag>
          <span v-if="!row.roles || row.roles.length === 0" style="color: #999;">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="permission_count" label="权限数" width="100" align="center" />
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="showDetail(row)">授权</el-button>
        </template>
      </el-table-column>
    </el-table>

    <MemberDetailDrawer
      v-model="detailDrawerVisible"
      :system-id="systemId"
      :member="selectedMember"
      @success="fetchMembers"
    />

    <!-- Assign Role Dialog -->
    <el-dialog v-model="assignDialogVisible" title="分配角色" width="500px">
      <el-form ref="assignFormRef" :model="assignForm" :rules="assignRules" label-width="100px">
        <el-form-item label="选择用户" prop="user_id">
          <el-select
            v-model="assignForm.user_id"
            filterable
            remote
            :remote-method="searchUsers"
            :loading="searchingUsers"
            placeholder="搜索用户名"
            style="width: 100%"
            :disabled="!!assignForm._fixedUser"
          >
            <el-option
              v-for="user in availableUsers"
              :key="user.id"
              :label="user.chinese_name ? `${user.chinese_name}(${user.username})` : user.username"
              :value="user.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="选择角色" prop="role_ids">
          <el-select
            v-model="assignForm.role_ids"
            multiple
            placeholder="请选择角色"
            style="width: 100%"
          >
            <el-option
              v-for="role in allRoles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assignDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAssign" :loading="assigning">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getMemberUsers } from '@/api/system'
import { getRoles, assignUsers, removeRoleUser } from '@/api/role'
import { getUsers } from '@/api/user'
import MemberDetailDrawer from './MemberDetailDrawer.vue'

interface Role {
  id: number
  name: string
  code: string
}

interface MemberUser {
  user_id: number
  username: string
  chinese_name?: string
  email?: string
  roles: Role[]
  permission_count: number
}

interface User {
  id: number
  username: string
  chinese_name?: string
}

const route = useRoute()
const systemId = computed(() => Number(route.params.id))

const members = ref<MemberUser[]>([])
const loading = ref(false)

const detailDrawerVisible = ref(false)
const selectedMember = ref<{ id: number; username: string; chinese_name?: string; email?: string; role: string } | null>(null)

const assignDialogVisible = ref(false)
const assigning = ref(false)
const allRoles = ref<Role[]>([])
const availableUsers = ref<User[]>([])
const searchingUsers = ref(false)
const assignFormRef = ref<FormInstance>()
const assignForm = ref<{ user_id: number | null; role_ids: number[]; _fixedUser?: boolean }>({
  user_id: null,
  role_ids: [],
  _fixedUser: false
})

const assignRules: FormRules = {
  user_id: [{ required: true, message: '请选择用户', trigger: 'change' }],
  role_ids: [{ required: true, type: 'array', min: 1, message: '请至少选择一个角色', trigger: 'change' }]
}

async function fetchMembers() {
  loading.value = true
  try {
    const data: any = await getMemberUsers(systemId.value)
    members.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取成员列表失败')
  } finally {
    loading.value = false
  }
}

async function fetchRoles() {
  try {
    const data: any = await getRoles(systemId.value)
    allRoles.value = data.list || []
  } catch {
    // ignore
  }
}

async function searchUsers(query: string) {
  if (!query) { availableUsers.value = []; return }
  searchingUsers.value = true
  try {
    const data: any = await getUsers({ page: 1, page_size: 50 })
    const users = data.list || []
    availableUsers.value = users.filter((u: User) =>
      u.username.toLowerCase().includes(query.toLowerCase()) ||
      (u.chinese_name && u.chinese_name.toLowerCase().includes(query.toLowerCase()))
    )
  } catch {
    // ignore
  } finally {
    searchingUsers.value = false
  }
}

function showDetail(member: MemberUser) {
  selectedMember.value = {
    id: member.user_id,
    username: member.username,
    chinese_name: member.chinese_name,
    email: member.email,
    role: 'member'
  }
  detailDrawerVisible.value = true
}

function showAssignDialog() {
  assignForm.value = { user_id: null, role_ids: [], _fixedUser: false }
  availableUsers.value = []
  assignDialogVisible.value = true
}

function showAssignDialogForUser(member: MemberUser) {
  assignForm.value = {
    user_id: member.user_id,
    role_ids: [],
    _fixedUser: true
  }
  availableUsers.value = [{ id: member.user_id, username: member.username, chinese_name: member.chinese_name }]
  assignDialogVisible.value = true
}

async function handleAssign() {
  if (!assignFormRef.value) return
  try {
    await assignFormRef.value.validate()
  } catch {
    return
  }
  if (!assignForm.value.user_id || assignForm.value.role_ids.length === 0) return

  assigning.value = true
  try {
    await Promise.all(
      assignForm.value.role_ids.map(roleId =>
        assignUsers(systemId.value, roleId, [assignForm.value.user_id!])
      )
    )
    ElMessage.success('角色分配成功')
    assignDialogVisible.value = false
    fetchMembers()
  } catch (error: any) {
    ElMessage.error(error.message || '角色分配失败')
  } finally {
    assigning.value = false
  }
}

async function handleRemoveRole(member: MemberUser, role: Role) {
  try {
    await ElMessageBox.confirm(
      `确定要移除 "${member.chinese_name ? `${member.chinese_name}(${member.username})` : member.username}" 的 "${role.name}" 角色吗？`,
      '确认移除',
      { type: 'warning' }
    )
    await removeRoleUser(systemId.value, role.id, member.user_id)
    ElMessage.success('角色移除成功')
    fetchMembers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '角色移除失败')
    }
  }
}

onMounted(() => {
  fetchMembers()
  fetchRoles()
})
</script>

<style scoped>
.member-view {
  padding: 0;
}
</style>
