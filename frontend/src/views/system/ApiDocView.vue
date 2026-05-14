<template>
  <div class="api-doc-view">
    <div class="api-doc-container">
      <!-- Left: API List -->
      <div class="api-list-panel">
        <div class="api-list-header">
          对外接口 <span class="api-count">({{ apiList.length }})</span>
        </div>
        <div class="api-list">
          <div
            v-for="api in apiList"
            :key="api.path"
            class="api-item"
            :class="{ active: selectedApi?.path === api.path }"
            @click="selectedApi = api"
          >
            <div class="api-item-header">
              <el-tag type="success" size="small" effect="dark">{{ api.method }}</el-tag>
              <code class="api-path">{{ api.path }}</code>
            </div>
            <div class="api-desc">{{ api.description }}</div>
          </div>
        </div>
      </div>

      <!-- Right: Detail Panel -->
      <div class="api-detail-panel" v-if="selectedApi">
        <!-- Title -->
        <div class="detail-section detail-header">
          <div class="detail-title">
            <el-tag type="success" effect="dark">{{ selectedApi.method }}</el-tag>
            <code class="detail-path">/api/external{{ selectedApi.path }}</code>
          </div>
          <p class="detail-desc">{{ selectedApi.description }}</p>
        </div>

        <!-- Request Headers -->
        <div class="detail-section">
          <h4>请求头</h4>
          <el-table :data="requestHeaders" border size="small">
            <el-table-column prop="name" label="参数名" width="180">
              <template #default="{ row }">
                <code>{{ row.name }}</code>
              </template>
            </el-table-column>
            <el-table-column prop="required" label="必填" width="80" align="center">
              <template #default="{ row }">
                <span :style="{ color: row.required ? '#f56c6c' : '#909399' }">
                  {{ row.required ? '是' : '否' }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="description" label="说明" />
          </el-table>
        </div>

        <!-- Query Parameters -->
        <div class="detail-section">
          <h4>请求参数</h4>
          <div v-if="selectedApi.params && selectedApi.params.length > 0">
            <el-table :data="selectedApi.params" border size="small">
              <el-table-column prop="name" label="参数名" width="150">
                <template #default="{ row }">
                  <code>{{ row.name }}</code>
                </template>
              </el-table-column>
              <el-table-column prop="required" label="必填" width="80" align="center">
                <template #default="{ row }">
                  <span :style="{ color: row.required ? '#f56c6c' : '#909399' }">
                    {{ row.required ? '是' : '否' }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column prop="type" label="类型" width="100" />
              <el-table-column prop="description" label="说明" />
            </el-table>
          </div>
          <div v-else class="no-params">无额外参数</div>
        </div>

        <!-- Response Example -->
        <div class="detail-section">
          <h4>响应示例</h4>
          <pre class="code-block"><code>{{ selectedApi.responseExample }}</code></pre>
        </div>

        <!-- Try It Out -->
        <div class="detail-section try-section">
          <h4>
            在线测试
            <el-tag type="warning" size="small" style="margin-left: 8px;">Try It Out</el-tag>
          </h4>
          <div class="try-form">
            <el-form label-width="140px" size="default">
              <el-form-item label="X-System-Code">
                <el-input v-model="testForm.systemCode" placeholder="系统编码" />
              </el-form-item>
              <el-form-item
                v-for="param in selectedApi.params"
                :key="param.name"
                :label="param.name"
              >
                <el-input v-model="testForm.params[param.name]" :placeholder="param.description" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleTest" :loading="testing">
                  发送请求
                </el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- Response Result -->
          <div v-if="testResult" class="test-result">
            <div class="result-header">
              <span>响应</span>
              <el-tag :type="testResult.status === 200 ? 'success' : 'danger'" size="small">
                {{ testResult.status }} {{ testResult.statusText }}
              </el-tag>
              <span class="result-time">{{ testResult.time }}ms</span>
            </div>
            <pre class="code-block"><code>{{ formatJson(testResult.data) }}</code></pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useSystemStore } from '@/stores/system'
import { getSystemById } from '@/api/system'
import { callExternalApi, type ExternalApiResult } from '@/api/external'

interface ApiParam {
  name: string
  required: boolean
  type: string
  description: string
}

interface ApiDefinition {
  method: string
  path: string
  description: string
  params?: ApiParam[]
  responseExample: string
}

const route = useRoute()
const systemStore = useSystemStore()
const systemId = computed(() => Number(route.params.id))
const systemCode = ref('')

const apiList: ApiDefinition[] = [
  {
    method: 'GET',
    path: '/user-roles',
    description: '获取用户信息及其在系统中的角色列表',
    params: [
      { name: 'username', required: true, type: 'string', description: '用户名（英文名）' },
    ],
    responseExample: `{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 42,
    "username": "alice",
    "is_super_admin": false,
    "roles": ["editor", "viewer"]
  }
}`,
  },
  {
    method: 'GET',
    path: '/menus',
    description: '获取用户在该系统中有权访问的菜单列表',
    params: [
      { name: 'username', required: true, type: 'string', description: '用户名（英文名）' },
    ],
    responseExample: `{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "Dashboard",
      "path": "/dashboard",
      "icon": "dashboard-icon",
      "parent_id": 0
    }
  ]
}`,
  },
  {
    method: 'GET',
    path: '/user-apis',
    description: '获取用户在该系统中有权访问的 API 接口列表',
    params: [
      { name: 'username', required: true, type: 'string', description: '用户名（英文名）' },
    ],
    responseExample: `{
  "code": 0,
  "message": "success",
  "data": ["user:create", "user:read", "order:export"]
}`,
  },
  {
    method: 'GET',
    path: '/check-permission',
    description: '校验用户是否拥有指定接口的访问权限',
    params: [
      { name: 'username', required: true, type: 'string', description: '用户名（英文名）' },
      { name: 'method', required: true, type: 'string', description: 'HTTP 方法，如 GET、POST' },
      { name: 'path', required: true, type: 'string', description: '接口路径，如 /api/orders' },
    ],
    responseExample: `{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true
  }
}`,
  },
]

const requestHeaders = [
  { name: 'X-System-Code', required: true, description: '系统编码（在系统创建时生成）' },
]

const selectedApi = ref<ApiDefinition>(apiList[0]!)
const testing = ref(false)
const testResult = ref<ExternalApiResult | null>(null)

const testForm = ref({
  systemCode: '',
  params: {} as Record<string, string>,
})

watch(selectedApi, () => {
  testResult.value = null
  testForm.value.params = {}
})

async function fetchSystemCode() {
  try {
    const sys = systemStore.currentSystem()
    if (sys) {
      systemCode.value = sys.code
      testForm.value.systemCode = sys.code
      return
    }
    const data: any = await getSystemById(systemId.value)
    systemCode.value = data.code
    testForm.value.systemCode = data.code
  } catch {
    // ignore
  }
}

async function handleTest() {
  if (!testForm.value.systemCode) return
  testing.value = true
  testResult.value = null
  try {
    const headers: Record<string, string> = {
      'X-System-Code': testForm.value.systemCode,
    }
    const params: Record<string, string> = {}
    if (selectedApi.value.params) {
      for (const p of selectedApi.value.params) {
        const val = testForm.value.params[p.name]
        if (val) {
          params[p.name] = val
        }
      }
    }
    testResult.value = await callExternalApi(
      selectedApi.value.path,
      headers,
      Object.keys(params).length > 0 ? params : undefined
    )
  } catch (error: any) {
    if (error.response) {
      testResult.value = {
        status: error.response.status,
        statusText: error.response.statusText,
        data: error.response.data,
        time: 0,
      }
    } else {
      testResult.value = {
        status: 0,
        statusText: 'Error',
        data: { message: error.message || '请求失败' },
        time: 0,
      }
    }
  } finally {
    testing.value = false
  }
}

function formatJson(data: any): string {
  try {
    return JSON.stringify(data, null, 2)
  } catch {
    return String(data)
  }
}

onMounted(() => {
  fetchSystemCode()
})
</script>

<style scoped>
.api-doc-view {
  padding: 0;
  height: calc(100vh - 120px);
}

.api-doc-container {
  display: flex;
  height: 100%;
  border: 1px solid var(--el-border-color-light);
  border-radius: var(--el-border-radius-base);
  background: var(--el-bg-color);
}

.api-list-panel {
  width: 240px;
  flex-shrink: 0;
  border-right: 1px solid var(--el-border-color-light);
  display: flex;
  flex-direction: column;
}

.api-list-header {
  padding: 14px 16px;
  font-weight: 600;
  font-size: 14px;
  border-bottom: 1px solid var(--el-border-color-light);
  background: var(--el-fill-color-lighter);
}

.api-count {
  color: var(--el-text-color-secondary);
  font-weight: normal;
  font-size: 12px;
}

.api-list {
  flex: 1;
  overflow-y: auto;
}

.api-item {
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid var(--el-border-color-lighter);
  transition: background-color 0.2s;
}

.api-item:hover {
  background: var(--el-fill-color-light);
}

.api-item.active {
  background: var(--el-color-primary-light-9);
  border-left: 3px solid var(--el-color-primary);
  padding-left: 13px;
}

.api-item-header {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 4px;
}

.api-path {
  font-size: 12px;
  font-weight: 500;
}

.api-desc {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.api-detail-panel {
  flex: 1;
  overflow-y: auto;
}

.detail-section {
  padding: 20px 24px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.detail-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.detail-header {
  background: var(--el-fill-color-lighter);
}

.detail-title {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-bottom: 8px;
}

.detail-path {
  font-size: 16px;
  font-weight: 600;
}

.detail-desc {
  margin: 0;
  color: var(--el-text-color-regular);
  font-size: 13px;
}

.no-params {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  font-style: italic;
}

.code-block {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.6;
  overflow-x: auto;
  margin: 0;
}

.try-section {
  background: var(--el-fill-color-lighter);
}

.try-form {
  max-width: 600px;
}

.test-result {
  margin-top: 16px;
}

.result-header {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
  font-size: 13px;
}

.result-time {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>
