<template>
  <div class="dashboard">
    <div class="welcome-section">
      <h2>欢迎使用 GRBAC 权限管理系统</h2>
      <p v-if="userStore.userInfo">
        当前用户：<strong>{{ userStore.userInfo.username }}</strong>
        <el-tag v-if="userStore.isSuperAdmin()" type="danger" style="margin-left: 8px">
          超级管理员
        </el-tag>
      </p>
    </div>

    <div class="stats-section" v-if="userStore.isSuperAdmin()">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-icon" style="background: #409eff">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.systemCount }}</div>
              <div class="stat-label">系统总数</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-icon" style="background: #67c23a">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.userCount }}</div>
              <div class="stat-label">用户总数</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-icon" style="background: #e6a23c">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.auditCount }}</div>
              <div class="stat-label">今日操作</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <div class="quick-actions">
      <h3>快捷操作</h3>
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card shadow="hover" class="action-card" @click="router.push('/systems')">
            <el-icon class="action-icon"><Monitor /></el-icon>
            <div class="action-text">系统管理</div>
          </el-card>
        </el-col>
        <el-col :span="6" v-if="userStore.isSuperAdmin()">
          <el-card shadow="hover" class="action-card" @click="router.push('/users')">
            <el-icon class="action-icon"><User /></el-icon>
            <div class="action-text">用户管理</div>
          </el-card>
        </el-col>
        <el-col :span="6" v-if="userStore.isSuperAdmin()">
          <el-card shadow="hover" class="action-card" @click="router.push('/audit-logs')">
            <el-icon class="action-icon"><Document /></el-icon>
            <div class="action-text">审计日志</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover" class="action-card" @click="router.push('/profile/password')">
            <el-icon class="action-icon"><Lock /></el-icon>
            <div class="action-text">修改密码</div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Monitor, User, Document, Lock } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { getSystems } from '@/api/system'
import { getUsers } from '@/api/user'
import { getAuditLogs } from '@/api/audit'

const router = useRouter()
const userStore = useUserStore()

const stats = ref({
  systemCount: 0,
  userCount: 0,
  auditCount: 0
})

async function fetchStats() {
  if (!userStore.isSuperAdmin()) return

  try {
    const [systemsData, usersData, auditData] = await Promise.all([
      getSystems(),
      getUsers({ page: 1, page_size: 1 }),
      getAuditLogs({ page: 1, page_size: 1 })
    ])

    const sysData = systemsData as any
    const usrData = usersData as any
    const audData = auditData as any

    stats.value = {
      systemCount: sysData.total || (Array.isArray(sysData) ? sysData.length : 0),
      userCount: usrData.total || 0,
      auditCount: audData.total || 0
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

onMounted(() => {
  fetchStats()
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.welcome-section {
  margin-bottom: var(--space-xl);
}

.welcome-section h2 {
  margin: 0 0 var(--space-sm) 0;
  font-size: var(--font-size-2xl);
  font-weight: 600;
  color: var(--color-text-primary);
}

.welcome-section p {
  margin: 0;
  font-size: var(--font-size-lg);
  color: var(--color-text-regular);
}

.stats-section {
  margin-bottom: var(--space-xl);
}

.stat-card {
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: var(--space-lg);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: var(--font-size-2xl);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1;
}

.stat-label {
  font-size: var(--font-size-base);
  color: var(--color-text-secondary);
  margin-top: var(--space-xs);
}

.quick-actions h3 {
  margin: 0 0 var(--space-md) 0;
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--color-text-primary);
}

.action-card {
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  text-align: center;
}

.action-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.action-card :deep(.el-card__body) {
  padding: var(--space-lg);
}

.action-icon {
  font-size: 36px;
  color: var(--color-primary);
  margin-bottom: var(--space-sm);
}

.action-text {
  font-size: var(--font-size-base);
  color: var(--color-text-regular);
}
</style>
