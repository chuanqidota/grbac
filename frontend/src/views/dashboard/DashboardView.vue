<template>
  <div class="dashboard">
    <div class="welcome-card">
      <div class="welcome-content">
        <h2>欢迎使用 GRBAC 权限管理系统</h2>
        <p v-if="userStore.userInfo">
          当前用户：<strong>{{ userStore.userInfo.chinese_name ? `${userStore.userInfo.chinese_name}(${userStore.userInfo.username})` : userStore.userInfo.username }}</strong>
          <el-tag v-if="userStore.isSuperAdmin()" type="danger" size="small" class="role-tag">
            超级管理员
          </el-tag>
        </p>
      </div>
    </div>

    <div class="stats-section" v-if="userStore.isSuperAdmin()">
      <el-row :gutter="20">
        <el-col :span="8">
          <div class="stat-card">
            <div class="stat-card-icon" style="background: var(--color-primary-light); color: var(--color-primary)">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.systemCount }}</div>
              <div class="stat-label">系统总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-card">
            <div class="stat-card-icon" style="background: var(--color-success-light); color: var(--color-success)">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.userCount }}</div>
              <div class="stat-label">用户总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-card">
            <div class="stat-card-icon" style="background: var(--color-warning-light); color: var(--color-warning)">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.auditCount }}</div>
              <div class="stat-label">今日操作</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="quick-actions">
      <h3>快捷操作</h3>
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="action-card" @click="router.push('/systems')">
            <el-icon class="action-icon" style="color: var(--color-primary)"><Monitor /></el-icon>
            <div class="action-text">系统管理</div>
          </div>
        </el-col>
        <el-col :span="6" v-if="userStore.isSuperAdmin()">
          <div class="action-card" @click="router.push('/users')">
            <el-icon class="action-icon" style="color: var(--color-success)"><User /></el-icon>
            <div class="action-text">用户管理</div>
          </div>
        </el-col>
        <el-col :span="6" v-if="userStore.isSuperAdmin()">
          <div class="action-card" @click="router.push('/audit-logs')">
            <el-icon class="action-icon" style="color: var(--color-warning)"><Document /></el-icon>
            <div class="action-text">审计日志</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="action-card" @click="router.push('/profile/password')">
            <el-icon class="action-icon" style="color: var(--color-danger)"><Lock /></el-icon>
            <div class="action-text">修改密码</div>
          </div>
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

/* ========== Welcome Card — Soft UI ========== */
.welcome-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  padding: var(--space-lg) var(--space-xl);
  box-shadow: var(--shadow-md);
  border: 1px solid rgba(228, 236, 252, 0.8);
  margin-bottom: var(--space-xl);
}

.welcome-content h2 {
  margin: 0 0 var(--space-sm) 0;
  font-size: var(--font-size-2xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.02em;
}

.welcome-content p {
  margin: 0;
  font-size: var(--font-size-lg);
  color: var(--color-text-regular);
}

.role-tag {
  margin-left: var(--space-sm);
  vertical-align: middle;
}

/* ========== Stats Section ========== */
.stats-section {
  margin-bottom: var(--space-xl);
}

.stat-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  box-shadow: var(--shadow-md);
  border: 1px solid rgba(228, 236, 252, 0.8);
  display: flex;
  align-items: center;
  gap: var(--space-md);
  transition: box-shadow var(--transition-base), transform var(--transition-base);
  cursor: pointer;
}

.stat-card:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

.stat-card-icon {
  width: 52px;
  height: 52px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
  box-shadow: var(--shadow-sm);
  transition: box-shadow var(--transition-base);
}

.stat-card:hover .stat-card-icon {
  box-shadow: var(--shadow-md);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1;
  letter-spacing: -0.02em;
}

.stat-label {
  font-size: var(--font-size-base);
  color: var(--color-text-secondary);
  margin-top: var(--space-xs);
  font-weight: 500;
}

/* ========== Quick Actions ========== */
.quick-actions h3 {
  margin: 0 0 var(--space-md) 0;
  font-size: var(--font-size-lg);
  font-weight: 700;
  color: var(--color-text-primary);
}

.action-card {
  cursor: pointer;
  text-align: center;
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  padding: var(--space-xl) var(--space-lg);
  box-shadow: var(--shadow-md);
  border: 1px solid rgba(228, 236, 252, 0.8);
  transition: box-shadow var(--transition-base), transform var(--transition-base);
}

.action-card:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

.action-icon {
  font-size: 36px;
  margin-bottom: var(--space-sm);
}

.action-text {
  font-size: var(--font-size-base);
  color: var(--color-text-regular);
  font-weight: 500;
}
</style>
