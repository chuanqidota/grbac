<template>
  <router-view />
</template>

<script setup lang="ts">
import { onErrorCaptured } from 'vue'
import { ElNotification } from 'element-plus'

onErrorCaptured((err, _instance, info) => {
  console.error('组件错误:', err, info)
  ElNotification.error({
    title: '页面错误',
    message: err.message || '发生了未知错误',
  })
  return false
})
</script>
