<template>
  <div class="audit-log-view">
    <div class="page-header">
      <h2>审计日志</h2>
    </div>

    <div class="filter-bar">
      <el-select
        v-model="filterSystemId"
        placeholder="筛选系统"
        clearable
        style="width: 200px"
        @change="handleFilter"
      >
        <el-option
          v-for="system in systems"
          :key="system.id"
          :label="system.name"
          :value="system.id"
        />
      </el-select>
    </div>

    <el-table :data="logs" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="system_name" label="系统" min-width="120" />
      <el-table-column prop="user_name" label="操作人" min-width="100" />
      <el-table-column prop="action" label="操作" min-width="120">
        <template #default="{ row }">
          <el-tag :type="getActionTagType(row.action)">
            {{ row.action }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="resource_type" label="资源类型" min-width="100" />
      <el-table-column prop="resource_id" label="资源ID" width="100" />
      <el-table-column prop="ip_address" label="IP地址" min-width="120" />
      <el-table-column prop="created_at" label="操作时间" min-width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="详情" width="80">
        <template #default="{ row }">
          <el-button type="primary" link @click="showDetail(row)">
            查看
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

    <!-- Detail Dialog -->
    <el-dialog
      v-model="detailVisible"
      title="日志详情"
      width="600px"
    >
      <el-descriptions :column="1" border>
        <el-descriptions-item label="ID">{{ detailLog?.id }}</el-descriptions-item>
        <el-descriptions-item label="系统">{{ detailLog?.system_name }}</el-descriptions-item>
        <el-descriptions-item label="操作人">{{ detailLog?.user_name }}</el-descriptions-item>
        <el-descriptions-item label="操作">{{ detailLog?.action }}</el-descriptions-item>
        <el-descriptions-item label="资源类型">{{ detailLog?.resource_type }}</el-descriptions-item>
        <el-descriptions-item label="资源ID">{{ detailLog?.resource_id }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ detailLog?.ip_address }}</el-descriptions-item>
        <el-descriptions-item label="操作时间">{{ formatDate(detailLog?.created_at) }}</el-descriptions-item>
      </el-descriptions>
      <div v-if="detailLog?.detail" class="detail-json">
        <h4>详细信息</h4>
        <pre>{{ formatJson(detailLog.detail) }}</pre>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAuditLogs } from '@/api/audit'
import { getSystems } from '@/api/system'

interface AuditLog {
  id: number
  system_name: string
  user_name: string
  action: string
  resource_type: string
  resource_id: number
  ip_address: string
  detail?: string
  created_at: string
}

interface System {
  id: number
  name: string
}

const logs = ref<AuditLog[]>([])
const systems = ref<System[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filterSystemId = ref<number | null>(null)

const detailVisible = ref(false)
const detailLog = ref<AuditLog | null>(null)

function getActionTagType(action: string) {
  if (action.includes('create')) return 'success'
  if (action.includes('update')) return 'primary'
  if (action.includes('delete')) return 'danger'
  return 'info'
}

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

function formatJson(jsonStr: string) {
  try {
    return JSON.stringify(JSON.parse(jsonStr), null, 2)
  } catch {
    return jsonStr
  }
}

async function fetchSystems() {
  try {
    const data: any = await getSystems()
    systems.value = data.list || []
  } catch (error: any) {
    console.error('获取系统列表失败:', error)
  }
}

async function fetchLogs() {
  loading.value = true
  try {
    const params: any = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (filterSystemId.value) {
      params.system_id = filterSystemId.value
    }
    const data: any = await getAuditLogs(params)
    logs.value = data.list || []
    total.value = data.total || 0
  } catch (error: any) {
    ElMessage.error(error.message || '获取审计日志失败')
  } finally {
    loading.value = false
  }
}

function handleFilter() {
  currentPage.value = 1
  fetchLogs()
}

function handleSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
  fetchLogs()
}

function handleCurrentChange(page: number) {
  currentPage.value = page
  fetchLogs()
}

function showDetail(log: AuditLog) {
  detailLog.value = log
  detailVisible.value = true
}

onMounted(() => {
  fetchSystems()
  fetchLogs()
})
</script>

<style scoped>
.audit-log-view {
  padding: 0;
}

.filter-bar {
  margin-bottom: var(--space-md);
}

.pagination {
  margin-top: var(--space-md);
  display: flex;
  justify-content: flex-end;
}

.detail-json {
  margin-top: var(--space-md);
}

.detail-json h4 {
  margin: 0 0 var(--space-sm) 0;
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--color-text-primary);
}

.detail-json pre {
  background: var(--color-bg-page);
  padding: var(--space-md);
  border-radius: var(--radius-sm);
  overflow-x: auto;
  font-size: var(--font-size-xs);
  line-height: 1.5;
  color: var(--color-text-regular);
}
</style>
