<template>
  <div class="permission-view">
    <div class="page-header">
      <h2>接口权限</h2>
      <div class="header-actions">
        <el-button v-if="selectedIds.length > 0" type="danger" @click="handleBatchDelete">
          批量删除 ({{ selectedIds.length }})
        </el-button>
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          添加权限
        </el-button>
      </div>
    </div>

    <div class="filter-bar">
      <el-input
        v-model="urlState.keyword"
        placeholder="搜索权限编码或名称"
        clearable
        style="width: 240px"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-radio-group v-model="urlState.method" @change="handleMethodChange">
        <el-radio-button label="">全部</el-radio-button>
        <el-radio-button label="GET">GET</el-radio-button>
        <el-radio-button label="POST">POST</el-radio-button>
        <el-radio-button label="PUT">PUT</el-radio-button>
        <el-radio-button label="DELETE">DELETE</el-radio-button>
        <el-radio-button label="PATCH">PATCH</el-radio-button>
      </el-radio-group>
    </div>

    <el-skeleton :loading="skeleton" animated :count="5">
      <template #template>
        <el-skeleton-item variant="text" style="width: 40%; height: 32px; margin-bottom: 16px;" />
        <div v-for="i in 5" :key="i" style="display: flex; gap: 16px; margin-bottom: 12px;">
          <el-skeleton-item variant="text" style="width: 3%;" />
          <el-skeleton-item variant="text" style="width: 5%;" />
          <el-skeleton-item variant="text" style="width: 15%;" />
          <el-skeleton-item variant="text" style="width: 15%;" />
          <el-skeleton-item variant="text" style="width: 10%;" />
          <el-skeleton-item variant="text" style="width: 25%;" />
          <el-skeleton-item variant="text" style="width: 12%;" />
        </div>
      </template>
      <template #default>
    <el-table
      :data="permissions"
      v-loading="loading"
      border
      stripe
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="50" />
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="code" label="权限编码" min-width="150" />
      <el-table-column prop="name" label="权限名称" min-width="150" />
      <el-table-column prop="method" label="请求方法" width="100">
        <template #default="{ row }">
          <el-tag :type="getMethodTagType(row.method)">{{ row.method }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="path" label="请求路径" min-width="200" />
      <el-table-column prop="created_at" label="创建时间" min-width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="showEditDialog(row)">编辑</el-button>
          <el-button type="danger" link @click="confirmDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
      </template>
    </el-skeleton>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑权限' : '创建权限'"
      width="500px"
      :before-close="handleDialogClose"
    >
      <el-form ref="formRef" :model="formData" :rules="rules" label-width="100px">
        <el-form-item label="权限编码" prop="code">
          <el-input v-model="formData.code" :disabled="isEditing" placeholder="请输入权限编码" :autofocus="!isEditing" />
        </el-form-item>
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入权限名称" />
        </el-form-item>
        <el-form-item label="请求方法" prop="method">
          <el-select v-model="formData.method" placeholder="请选择请求方法">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
            <el-option label="PATCH" value="PATCH" />
          </el-select>
        </el-form-item>
        <el-form-item label="请求路径" prop="path">
          <el-input v-model="formData.path" placeholder="请输入请求路径" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="formData.description" type="textarea" :rows="3" placeholder="请输入权限描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeDialog">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { getPermissions, createPermission, updatePermission, deletePermission } from '@/api/permission'
import { formatDate } from '@/utils/format'
import { useUrlState, useTable, useFormDialog, useConfirmDelete } from '@/composables'

defineOptions({ name: 'PermissionView' })

interface Permission {
  id: number
  code: string
  name: string
  method: string
  path: string
  description?: string
  created_at: string
}

interface PermissionForm {
  code: string
  name: string
  method: string
  path: string
  description: string
}

const route = useRoute()
const systemId = computed(() => Number(route.params.id))

// --- URL state ---
const { state: urlState } = useUrlState({
  keyword: '',
  page: 1,
  pageSize: 20,
  method: ''
})

// --- Table ---
const table = useTable<Permission>(
  async (params) => {
    const data: any = await getPermissions(systemId.value, {
      ...params,
      keyword: urlState.keyword || undefined,
      method: urlState.method || undefined
    })
    return { list: data.list || [], total: data.total || 0 }
  },
  { defaultPageSize: urlState.pageSize }
)

const { data: permissions, total, loading, skeleton, page, pageSize, refresh, handleSizeChange, handleCurrentChange } = table

// --- Sync URL state to table ---
watch(() => urlState.page, (val) => {
  if (page.value !== val) {
    page.value = val
    refresh()
  }
})

watch(() => urlState.pageSize, (val) => {
  if (pageSize.value !== val) {
    pageSize.value = val
    refresh()
  }
})

// --- Form dialog ---
const formRef = ref<FormInstance>()
const rules: FormRules = {
  code: [{ required: true, message: '请输入权限编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
  method: [{ required: true, message: '请选择请求方法', trigger: 'change' }],
  path: [{ required: true, message: '请输入请求路径', trigger: 'blur' }]
}

async function submitPermission(data: PermissionForm & { id?: number }) {
  if (data.id) {
    await updatePermission(systemId.value, data.id, {
      name: data.name,
      method: data.method,
      path: data.path,
      description: data.description
    })
    ElMessage.success('更新成功')
  } else {
    await createPermission(systemId.value, {
      code: data.code,
      name: data.name,
      method: data.method,
      path: data.path,
      description: data.description
    })
    ElMessage.success('创建成功')
  }
}

const {
  visible: dialogVisible,
  isEditing,
  formData,
  submitting,
  open: openDialog,
  close: closeDialog,
  submit: submitForm
} = useFormDialog<PermissionForm>(submitPermission, { onSuccess: refresh })

// --- Delete ---
const { confirmDelete, batchDelete } = useConfirmDelete(
  (id) => deletePermission(systemId.value, id),
  { onSuccess: refresh, entityName: '权限' }
)

// --- Selected IDs ---
const selectedIds = ref<number[]>([])

function handleSelectionChange(selection: Permission[]) {
  selectedIds.value = selection.map(p => p.id)
}

function handleBatchDelete() {
  batchDelete(selectedIds.value, {
    onComplete: () => { selectedIds.value = [] }
  })
}

// --- Helpers ---
function getMethodTagType(method: string) {
  const types: Record<string, string> = { GET: 'success', POST: 'primary', PUT: 'warning', DELETE: 'danger', PATCH: 'info' }
  return types[method] || ''
}

function handleMethodChange() {
  page.value = 1
  refresh()
}

function showCreateDialog() {
  openDialog({ code: '', name: '', method: 'GET', path: '', description: '' })
}

function showEditDialog(p: Permission) {
  openDialog({
    id: p.id,
    code: p.code,
    name: p.name,
    method: p.method,
    path: p.path,
    description: p.description || ''
  })
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  try {
    await submitForm()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  }
}

// --- Keyword debounce ---
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(() => urlState.keyword, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    refresh()
  }, 300)
})

function handleDialogClose(done: () => void) {
  closeDialog()
}

onUnmounted(() => {
  if (searchTimer) clearTimeout(searchTimer)
})
</script>

<style scoped>
.permission-view {
  padding: 0;
}

.filter-bar {
  display: flex;
  gap: var(--space-md);
  margin-bottom: var(--space-md);
  align-items: center;
}

.pagination {
  margin-top: var(--space-md);
  display: flex;
  justify-content: flex-end;
}
</style>
