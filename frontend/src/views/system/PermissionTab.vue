<template>
  <div class="permission-tab">
    <div class="tab-header">
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        创建权限
      </el-button>
    </div>

    <el-table :data="permissions" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="code" label="权限编码" min-width="150" />
      <el-table-column prop="name" label="权限名称" min-width="150" />
      <el-table-column prop="method" label="请求方法" width="100">
        <template #default="{ row }">
          <el-tag :type="getMethodTagType(row.method)">
            {{ row.method }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="path" label="请求路径" min-width="200" />
      <el-table-column prop="description" label="描述" min-width="150" />
      <el-table-column prop="created_at" label="创建时间" min-width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="showEditDialog(row)">
            编辑
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
      :title="isEditing ? '编辑权限' : '创建权限'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="权限编码" prop="code">
          <el-input
            v-model="form.code"
            :disabled="isEditing"
            placeholder="请输入权限编码"
          />
        </el-form-item>
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入权限名称" />
        </el-form-item>
        <el-form-item label="请求方法" prop="method">
          <el-select v-model="form.method" placeholder="请选择请求方法">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
            <el-option label="PATCH" value="PATCH" />
          </el-select>
        </el-form-item>
        <el-form-item label="请求路径" prop="path">
          <el-input v-model="form.path" placeholder="请输入请求路径" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入权限描述"
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getPermissions, createPermission, updatePermission, deletePermission } from '@/api/permission'

interface Permission {
  id: number
  code: string
  name: string
  method: string
  path: string
  description?: string
  created_at: string
}

const props = defineProps<{
  systemId: number
}>()

const permissions = ref<Permission[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const dialogVisible = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref<number | null>(null)

const formRef = ref<FormInstance>()
const form = ref({
  code: '',
  name: '',
  method: 'GET',
  path: '',
  description: ''
})

const rules: FormRules = {
  code: [
    { required: true, message: '请输入权限编码', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入权限名称', trigger: 'blur' }
  ],
  method: [
    { required: true, message: '请选择请求方法', trigger: 'change' }
  ],
  path: [
    { required: true, message: '请输入请求路径', trigger: 'blur' }
  ]
}

function getMethodTagType(method: string) {
  const types: Record<string, string> = {
    GET: 'success',
    POST: 'primary',
    PUT: 'warning',
    DELETE: 'danger',
    PATCH: 'info'
  }
  return types[method] || ''
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

async function fetchPermissions() {
  loading.value = true
  try {
    const data: any = await getPermissions(props.systemId, {
      page: currentPage.value,
      page_size: pageSize.value
    })
    permissions.value = data.list || []
    total.value = data.total || 0
  } catch (error: any) {
    ElMessage.error(error.message || '获取权限列表失败')
  } finally {
    loading.value = false
  }
}

function handleSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
  fetchPermissions()
}

function handleCurrentChange(page: number) {
  currentPage.value = page
  fetchPermissions()
}

function showCreateDialog() {
  isEditing.value = false
  editingId.value = null
  form.value = { code: '', name: '', method: 'GET', path: '', description: '' }
  dialogVisible.value = true
}

function showEditDialog(permission: Permission) {
  isEditing.value = true
  editingId.value = permission.id
  form.value = {
    code: permission.code,
    name: permission.name,
    method: permission.method,
    path: permission.path,
    description: permission.description || ''
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
        await updatePermission(props.systemId, editingId.value, {
          name: form.value.name,
          method: form.value.method,
          path: form.value.path,
          description: form.value.description
        })
        ElMessage.success('更新成功')
      } else {
        await createPermission(props.systemId, {
          code: form.value.code,
          name: form.value.name,
          method: form.value.method,
          path: form.value.path,
          description: form.value.description
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchPermissions()
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function handleDelete(permission: Permission) {
  try {
    await ElMessageBox.confirm(
      `确定要删除权限 "${permission.name}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )

    await deletePermission(props.systemId, permission.id)
    ElMessage.success('删除成功')
    fetchPermissions()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

watch(() => props.systemId, () => {
  fetchPermissions()
})

onMounted(() => {
  fetchPermissions()
})
</script>

<style scoped>
.permission-tab {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.tab-header {
  margin-bottom: 16px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
