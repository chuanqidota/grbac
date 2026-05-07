<template>
  <div class="dashboard">
    <el-card shadow="hover">
      <template #header>
        <span>欢迎</span>
      </template>
      <p>欢迎使用 GRBAC 权限管理系统</p>
      <p v-if="userStore.userInfo">
        当前用户：<strong>{{ userStore.userInfo.username }}</strong>
      </p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
</script>

<style scoped>
.dashboard {
  padding: 20px;
}
</style>
