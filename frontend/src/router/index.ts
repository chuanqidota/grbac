import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/utils/token'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/LoginView.vue'),
      meta: { title: '登录' },
    },
    {
      path: '/',
      component: () => import('@/layouts/DefaultLayout.vue'),
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/DashboardView.vue'),
          meta: { title: '首页', icon: 'HomeFilled' },
        },
        {
          path: 'users',
          name: 'Users',
          component: () => import('@/views/user/UserView.vue'),
          meta: { title: '用户管理', icon: 'User', superAdmin: true },
        },
        {
          path: 'systems',
          name: 'Systems',
          component: () => import('@/views/system/SystemView.vue'),
          meta: { title: '系统管理', icon: 'Monitor', superAdmin: true },
        },
        {
          path: 'systems/:id/roles',
          name: 'SystemRoles',
          component: () => import('@/views/system/RoleView.vue'),
          meta: { title: '角色管理', icon: 'UserFilled' },
        },
        {
          path: 'systems/:id/menus',
          name: 'SystemMenus',
          component: () => import('@/views/system/MenuView.vue'),
          meta: { title: '菜单管理', icon: 'Menu' },
        },
        {
          path: 'systems/:id/permissions',
          name: 'SystemPermissions',
          component: () => import('@/views/system/PermissionView.vue'),
          meta: { title: '接口权限', icon: 'Key' },
        },
        {
          path: 'systems/:id/webhooks',
          name: 'SystemWebhooks',
          component: () => import('@/views/system/WebhookView.vue'),
          meta: { title: 'Webhook', icon: 'Connection' },
        },
        {
          path: 'systems/:id/members',
          name: 'SystemMembers',
          component: () => import('@/views/system/MemberView.vue'),
          meta: { title: '成员管理', icon: 'User' },
        },
        {
          path: 'audit-logs',
          name: 'AuditLogs',
          component: () => import('@/views/audit/AuditLogView.vue'),
          meta: { title: '审计日志', icon: 'Document', superAdmin: true },
        },
        {
          path: 'profile/password',
          name: 'ChangePassword',
          component: () => import('@/views/profile/ChangePasswordView.vue'),
          meta: { title: '修改密码', icon: 'Lock' },
        },
      ],
    },
    {
      path: '/403',
      name: 'Forbidden',
      component: () => import('@/views/error/ForbiddenView.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/error/NotFoundView.vue'),
    },
  ],
})

router.beforeEach((to, from, next) => {
  const token = getToken()

  if (to.path === '/login') {
    next()
    return
  }

  if (!token) {
    next('/login')
    return
  }

  document.title = (to.meta.title as string) ? `${to.meta.title} - GRBAC` : 'GRBAC'
  next()
})

export default router
