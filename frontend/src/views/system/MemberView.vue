<template>
  <div class="member-view">
    <div class="page-header">
      <h2>成员管理</h2>
      <div class="header-actions">
        <el-button v-if="selectedIds.length > 0" type="danger" @click="handleBatchRemove">
          批量移除 ({{ selectedIds.length }})
        </el-button>
        <el-button type="primary" @click="showAddDialog">
          <el-icon><Plus /></el-icon>
          添加成员
        </el-button>
      </div>
    </div>

    <el-table
      :data="members"
      v-loading="loading"
      border
      stripe
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="50" />
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" min-width="120" />
      <el-table-column prop="email" label="邮箱" min-width="180" />
      <el-table-column prop="role" label="角色" width="120">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'primary'">
            {{ row.role === 'admin' ? '管理员' : '成员' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="加入时间" min-width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button type="danger" link @click="handleRemove(row)">移除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Add Member Dialog -->
    <el-dialog v-model="dialogVisible" title="添加成员" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="选择用户" prop="user_id">
          <el-select
            v-model="form.user_id"
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
        </el-form-item>
        <el-form-item label="成员角色" prop="role">
          <el-select v-model="form.role" placeholder="请选择角色">
            <el-option label="管理员" value="admin" />
            <el-option label="成员" value="member" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
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
import { getSystemMembers, addSystemMember, removeSystemMember } from '@/api/system'
import { getUsers } from '@/api/user'

interface Member {
  id: number
  username: string
  email?: string
  role: string
  created_at: string
}

interface User {
  id: number
  username: string
}

const route = useRoute()
const systemId = computed(() => Number(route.params.id))

const members = ref<Member[]>([])
const loading = ref(false)
const selectedIds = ref<number[]>([])

const dialogVisible = ref(false)
const submitting = ref(false)
const availableUsers = ref<User[]>([])
const searchingUsers = ref(false)

const formRef = ref<FormInstance>()
const form = ref({ user_id: null as number | null, role: 'member' })

const rules: FormRules = {
  user_id: [{ required: true, message: '请选择用户', trigger: 'change' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

async function fetchMembers() {
  loading.value = true
  try {
    const data: any = await getSystemMembers(systemId.value)
    members.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取成员列表失败')
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(selection: Member[]) {
  selectedIds.value = selection.map(m => m.id)
}

async function searchUsers(query: string) {
  if (!query) { availableUsers.value = []; return }
  searchingUsers.value = true
  try {
    const data: any = await getUsers({ page: 1, page_size: 50 })
    const users = data.list || []
    availableUsers.value = users.filter((u: User) => u.username.toLowerCase().includes(query.toLowerCase()))
  } catch (error: any) {
    ElMessage.error(error.message || '搜索用户失败')
  } finally {
    searchingUsers.value = false
  }
}

function showAddDialog() {
  form.value = { user_id: null, role: 'member' }
  availableUsers.value = []
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  if (!form.value.user_id) return
  submitting.value = true
  try {
    await addSystemMember(systemId.value, { user_id: form.value.user_id, role: form.value.role })
    ElMessage.success('成员添加成功')
    dialogVisible.value = false
    fetchMembers()
  } catch (error: any) {
    ElMessage.error(error.message || '成员添加失败')
  } finally {
    submitting.value = false
  }
}

async function handleRemove(member: Member) {
  try {
    await ElMessageBox.confirm(`确定要移除成员 "${member.username}" 吗？`, '确认移除', { type: 'warning' })
    await removeSystemMember(systemId.value, member.id)
    ElMessage.success('成员移除成功')
    fetchMembers()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '成员移除失败')
  }
}

async function handleBatchRemove() {
  try {
    await ElMessageBox.confirm(`确定要移除选中的 ${selectedIds.value.length} 个成员吗？`, '批量移除', { type: 'warning' })
    await Promise.all(selectedIds.value.map(id => removeSystemMember(systemId.value, id)))
    ElMessage.success('批量移除成功')
    selectedIds.value = []
    fetchMembers()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '批量移除失败')
  }
}

onMounted(() => { fetchMembers() })
</script>

<style scoped>
.member-view {
  padding: 0;
}
</style>
