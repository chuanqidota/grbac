<template>
  <div class="error-page">
    <h1>403</h1>
    <p>抱歉，您没有权限访问此页面</p>
    <el-button type="primary" @click="router.push('/dashboard')">返回首页</el-button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped>
.error-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  text-align: center;
}

.error-page h1 {
  font-size: 96px;
  color: var(--color-warning);
  margin-bottom: var(--space-md);
}

.error-page p {
  font-size: var(--font-size-xl);
  color: var(--color-text-regular);
  margin-bottom: var(--space-lg);
}
</style>
