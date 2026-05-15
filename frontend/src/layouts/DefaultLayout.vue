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
        background-color="var(--sidebar-bg)"
        text-color="rgba(255,255,255,0.65)"
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

        <el-menu-item v-if="userStore.isSuperAdmin() || isSystemAdmin" index="/systems">
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

          <el-menu-item :index="`/systems/${systemStore.currentSystemId}/api-docs`">
            <el-icon><Document /></el-icon>
            <template #title>接口文档</template>
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
              <span class="username">{{ userStore.userInfo?.chinese_name ? `${userStore.userInfo.chinese_name}(${userStore.userInfo.username})` : (userStore.userInfo?.username ?? '用户') }}</span>
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
      <div class="breadcrumb-bar">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item
            v-for="item in breadcrumbs"
            :key="item.path"
            :to="item.path ? { path: item.path } : undefined"
          >
            {{ item.title }}
          </el-breadcrumb-item>
        </el-breadcrumb>
      </div>
      <el-main class="layout-main">
        <router-view :key="route.fullPath" />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
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
const isSystemAdmin = computed(() =>
  systemStore.systems.some(s => s.current_user_role === 'admin')
)

const breadcrumbs = computed(() => {
  const items: { title: string; path?: string }[] = [{ title: '首页', path: '/dashboard' }]
  const path = route.path

  // System-scoped routes: /systems/:id/xxx
  const sysMatch = path.match(/^\/systems\/(\d+)(?:\/(.+))?$/)
  if (sysMatch) {
    items.push({ title: '系统管理', path: '/systems' })
    const sys = systemStore.systems.find(s => s.id === Number(sysMatch[1]))
    if (sys) {
      items.push({ title: sys.name })
    }
    if (sysMatch[2] && route.meta.title) {
      items.push({ title: route.meta.title as string })
    }
    return items
  }

  // Non-system routes: use route title
  if (path === '/dashboard') return items
  if (route.meta.title) {
    items.push({ title: route.meta.title as string })
  }
  return items
})

onMounted(async () => {
  try {
    await userStore.fetchUserInfo()
  } catch {
    // ignore
  }
  // Fetch systems for all authenticated users (super-admins see all, others see their own)
  try {
    await systemStore.fetchSystems()
  } catch {
    // ignore
  }
  // If URL already has a system ID, sync it to the store
  const match = route.path.match(/^\/systems\/(\d+)\//)
  if (match) {
    const urlSystemId = Number(match[1])
    if (urlSystemId && urlSystemId !== systemStore.currentSystemId) {
      systemStore.setCurrentSystem(urlSystemId)
    }
  }
})

// Keep store in sync when navigating between system-scoped routes
watch(() => route.path, (path) => {
  const match = path.match(/^\/systems\/(\d+)\//)
  if (match) {
    const urlSystemId = Number(match[1])
    if (urlSystemId && urlSystemId !== systemStore.currentSystemId) {
      systemStore.setCurrentSystem(urlSystemId)
    }
  }
})

function handleSystemChange(id: number | null) {
  if (!id) return
  systemStore.setCurrentSystem(id)
  // If currently on a system-scoped route, navigate to the same page under the new system
  const match = route.path.match(/^\/systems\/\d+\/(.+)$/)
  if (match) {
    router.replace(`/systems/${id}/${match[1]}`)
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
  background-color: var(--sidebar-bg);
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
  background-color: rgba(255, 255, 255, 0.1) !important;
}

.logo {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: var(--font-size-xl);
  font-weight: 700;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.aside-menu {
  border-right: none;
}

.menu-divider {
  margin: var(--space-sm) var(--space-md);
  border-color: rgba(255, 255, 255, 0.1);
}

.menu-divider :deep(.el-divider__text) {
  background-color: var(--sidebar-bg);
  color: rgba(255, 255, 255, 0.3);
  font-size: var(--font-size-xs);
  padding: 0 var(--space-sm);
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
  background: var(--header-bg);
  box-shadow: var(--shadow-sm);
  padding: 0 var(--space-md);
  height: var(--header-height);
}

.header-left {
  display: flex;
  align-items: center;
  width: 100px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.system-label {
  font-size: var(--font-size-base);
  color: var(--color-text-regular);
  white-space: nowrap;
}

.collapse-btn {
  font-size: var(--font-size-xl);
  cursor: pointer;
  color: var(--color-text-regular);
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  cursor: pointer;
  color: var(--color-text-regular);
}

.username {
  font-size: var(--font-size-base);
}

.breadcrumb-bar {
  padding: var(--space-sm) var(--space-lg);
  background: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border-light);
}

.layout-main {
  background: var(--color-bg-page);
  overflow: auto;
  padding: var(--space-lg);
}

@media (max-width: 1024px) {
  .header-left {
    width: auto;
  }

  .system-label {
    display: none;
  }

  .breadcrumb-bar {
    padding: var(--space-xs) var(--space-md);
  }
}
</style>
