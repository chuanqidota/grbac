<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapsed ? '64px' : '210px'" class="layout-aside">
      <div class="logo">
        <span v-if="!isCollapsed">GRBAC</span>
        <span v-else>G</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapsed"
        :router="true"
        background-color="#001529"
        text-color="#ffffffa6"
        active-text-color="#ffffff"
        class="aside-menu"
      >
        <el-menu-item index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <template #title>首页</template>
        </el-menu-item>

        <el-menu-item v-if="userStore.isSuperAdmin()" index="/users">
          <el-icon><User /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>

        <el-menu-item v-if="userStore.isSuperAdmin()" index="/systems">
          <el-icon><Monitor /></el-icon>
          <template #title>系统管理</template>
        </el-menu-item>

        <!-- System sub-menus (shown when a system is selected) -->
        <template v-if="systemStore.currentSystemId">
          <el-divider v-if="!isCollapsed" class="menu-divider">
            <span class="divider-label">当前系统: {{ systemStore.currentSystem()?.name }}</span>
          </el-divider>
          <el-divider v-else class="menu-divider" />

          <el-menu-item :index="`/systems/${systemStore.currentSystemId}/roles`">
            <el-icon><UserFilled /></el-icon>
            <template #title>角色管理</template>
          </el-menu-item>

          <el-menu-item :index="`/systems/${systemStore.currentSystemId}/menus`">
            <el-icon><Menu /></el-icon>
            <template #title>菜单管理</template>
          </el-menu-item>

          <el-menu-item :index="`/systems/${systemStore.currentSystemId}/permissions`">
            <el-icon><Key /></el-icon>
            <template #title>接口权限</template>
          </el-menu-item>

          <el-menu-item :index="`/systems/${systemStore.currentSystemId}/webhooks`">
            <el-icon><Connection /></el-icon>
            <template #title>Webhook</template>
          </el-menu-item>

          <el-menu-item :index="`/systems/${systemStore.currentSystemId}/members`">
            <el-icon><User /></el-icon>
            <template #title>成员管理</template>
          </el-menu-item>
        </template>

        <el-divider class="menu-divider" />

        <el-menu-item v-if="userStore.isSuperAdmin()" index="/audit-logs">
          <el-icon><Document /></el-icon>
          <template #title>审计日志</template>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container class="main-container">
      <el-header class="layout-header">
        <div class="header-left">
          <el-icon
            class="collapse-btn"
            @click="isCollapsed = !isCollapsed"
          >
            <Fold v-if="!isCollapsed" />
            <Expand v-else />
          </el-icon>
        </div>
        <div class="header-right">
          <template v-if="systemStore.systems.length > 0">
            <span class="system-label">当前系统:</span>
            <el-select
              v-model="systemStore.currentSystemId"
              placeholder="选择系统"
              size="default"
              clearable
              style="width: 200px"
              @change="handleSystemChange"
            >
              <el-option
                v-for="sys in systemStore.systems"
                :key="sys.id"
                :label="sys.name"
                :value="sys.id"
              />
            </el-select>
          </template>
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="user-dropdown">
              <el-icon><UserFilled /></el-icon>
              <span class="username">{{ userStore.userInfo?.username ?? '用户' }}</span>
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="password">
                  <el-icon><Lock /></el-icon>修改密码
                </el-dropdown-item>
                <el-dropdown-item command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { HomeFilled, Fold, Expand, UserFilled, ArrowDown, SwitchButton, User, Monitor, Document, Lock, Menu, Key, Connection } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { useSystemStore } from '@/stores/system'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const userStore = useUserStore()
const systemStore = useSystemStore()

const isCollapsed = ref(false)
const activeMenu = computed(() => route.path)

onMounted(async () => {
  try {
    await userStore.fetchUserInfo()
  } catch {
    // ignore
  }
  if (userStore.isSuperAdmin()) {
    try {
      await systemStore.fetchSystems()
    } catch {
      // ignore
    }
  }
})

function handleSystemChange(id: number | null) {
  if (id) {
    systemStore.setCurrentSystem(id)
  }
}

async function handleCommand(command: string) {
  if (command === 'logout') {
    await authStore.logout()
    router.push('/login')
  } else if (command === 'password') {
    router.push('/profile/password')
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.layout-aside {
  background-color: #001529;
  transition: width 0.3s;
  overflow: hidden;
  height: 100vh;
}

.layout-aside :deep(.el-menu) {
  height: calc(100vh - 50px);
  overflow-y: auto;
}

.layout-aside :deep(.el-menu-item) {
  cursor: pointer;
}

.layout-aside :deep(.el-menu-item:hover) {
  background-color: #ffffff1a !important;
}

.logo {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
  font-weight: 700;
  border-bottom: 1px solid #ffffff1a;
}

.aside-menu {
  border-right: none;
}

.menu-divider {
  margin: 8px 16px;
  border-color: #ffffff1a;
}

.menu-divider :deep(.el-divider__text) {
  background-color: #001529;
  color: #ffffff4d;
  font-size: 11px;
  padding: 0 8px;
}

.divider-label {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 140px;
  display: inline-block;
}

.main-container {
  overflow: hidden;
}

.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  padding: 0 16px;
}

.header-left {
  display: flex;
  align-items: center;
  width: 100px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.system-label {
  font-size: 14px;
  color: #606266;
  white-space: nowrap;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.username {
  font-size: 14px;
}

.layout-main {
  background: #f0f2f5;
  overflow: auto;
}
</style>
