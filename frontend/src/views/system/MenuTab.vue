<template>
  <div class="menu-tab">
    <div class="tab-header">
      <el-button type="primary" @click="showCreateDialog()">
        <el-icon><Plus /></el-icon>
        创建菜单
      </el-button>
    </div>

    <el-table
      :data="menus"
      v-loading="loading"
      border
      row-key="id"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
    >
      <el-table-column prop="name" label="菜单名称" min-width="180" />
      <el-table-column prop="path" label="路径" min-width="150" />
      <el-table-column prop="icon" label="图标" width="100" />
      <el-table-column prop="sort_order" label="排序" width="80" />
      <el-table-column prop="created_at" label="创建时间" min-width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="showCreateDialog(row.id)">
            添加子菜单
          </el-button>
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
      :title="isEditing ? '编辑菜单' : '创建菜单'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="上级菜单" prop="parent_id">
          <el-cascader
            v-model="form.parent_id"
            :options="menuOptions"
            :props="{ checkStrictly: true, value: 'id', label: 'name', children: 'children' }"
            clearable
            placeholder="无（顶级菜单）"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="菜单名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入菜单名称" />
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="form.path" placeholder="请输入路径" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="请输入图标名称" />
        </el-form-item>
        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="form.sort_order" :min="0" :max="9999" />
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
import { ref, onMounted, watch, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getMenus, createMenu, updateMenu, deleteMenu } from '@/api/menu'

interface Menu {
  id: number
  name: string
  path?: string
  icon?: string
  sort_order?: number
  parent_id?: number
  children?: Menu[]
  created_at: string
}

const props = defineProps<{
  systemId: number
}>()

const menus = ref<Menu[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref<number | null>(null)

const formRef = ref<FormInstance>()
const form = ref({
  parent_id: null as number | null,
  name: '',
  path: '',
  icon: '',
  sort_order: 0
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入菜单名称', trigger: 'blur' }
  ]
}

const menuOptions = computed(() => {
  const buildOptions = (items: Menu[]): any[] => {
    return items.map(item => ({
      id: item.id,
      name: item.name,
      children: item.children ? buildOptions(item.children) : undefined
    }))
  }
  return [{ id: 0, name: '顶级菜单', children: buildOptions(menus.value) }]
})

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

async function fetchMenus() {
  loading.value = true
  try {
    const data: any = await getMenus(props.systemId)
    menus.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取菜单列表失败')
  } finally {
    loading.value = false
  }
}

function showCreateDialog(parentId?: number) {
  isEditing.value = false
  editingId.value = null
  form.value = {
    parent_id: parentId || null,
    name: '',
    path: '',
    icon: '',
    sort_order: 0
  }
  dialogVisible.value = true
}

function showEditDialog(menu: Menu) {
  isEditing.value = true
  editingId.value = menu.id
  form.value = {
    parent_id: menu.parent_id || null,
    name: menu.name,
    path: menu.path || '',
    icon: menu.icon || '',
    sort_order: menu.sort_order || 0
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      const data = {
        parent_id: form.value.parent_id || undefined,
        name: form.value.name,
        path: form.value.path || undefined,
        icon: form.value.icon || undefined,
        sort_order: form.value.sort_order
      }

      if (isEditing.value && editingId.value) {
        await updateMenu(props.systemId, editingId.value, data)
        ElMessage.success('更新成功')
      } else {
        await createMenu(props.systemId, data)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchMenus()
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function handleDelete(menu: Menu) {
  try {
    await ElMessageBox.confirm(
      `确定要删除菜单 "${menu.name}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )

    await deleteMenu(props.systemId, menu.id)
    ElMessage.success('删除成功')
    fetchMenus()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

watch(() => props.systemId, () => {
  fetchMenus()
})

onMounted(() => {
  fetchMenus()
})
</script>

<style scoped>
.menu-tab {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.tab-header {
  margin-bottom: 16px;
}
</style>
