<template>
  <div class="system-view">
    <div class="page-header">
      <h2>系统管理</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        创建系统
      </el-button>
    </div>

    <el-table :data="systems" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="系统名称" min-width="140" />
      <el-table-column prop="code" label="系统编码" min-width="120">
        <template #default="{ row }">
          <el-tag>{{ row.code }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
      <el-table-column prop="created_at" label="创建时间" min-width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
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

    <!-- Create/Edit Dialog -->
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
        <el-form-item label="系统编码" prop="code">
          <el-input
            v-model="form.code"
            :disabled="isEditing"
            placeholder="请输入系统编码"
          />
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getSystems, createSystem, updateSystem, deleteSystem } from '@/api/system'

interface System {
  id: number
  name: string
  code: string
  description?: string
  created_at: string
}

const systems = ref<System[]>([])
const loading = ref(false)

const dialogVisible = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref<number | null>(null)

const formRef = ref<FormInstance>()
const form = ref({
  name: '',
  code: '',
  description: ''
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入系统名称', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入系统编码', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_-]*$/, message: '编码只能包含字母、数字、下划线和连字符，且以字母开头', trigger: 'blur' }
  ]
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

async function fetchSystems() {
  loading.value = true
  try {
    const data: any = await getSystems()
    systems.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取系统列表失败')
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  isEditing.value = false
  editingId.value = null
  form.value = { name: '', code: '', description: '' }
  dialogVisible.value = true
}

function showEditDialog(system: System) {
  isEditing.value = true
  editingId.value = system.id
  form.value = {
    name: system.name,
    code: system.code,
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
        await createSystem({
          name: form.value.name,
          code: form.value.code,
          description: form.value.description
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchSystems()
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
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
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
</style>
