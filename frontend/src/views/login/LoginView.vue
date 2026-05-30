<template>
  <div class="login-container">
    <div class="login-card-wrapper">
      <el-card class="login-card" shadow="never">
        <div class="login-header">
          <div class="login-logo">G</div>
          <h2 class="login-title">GRBAC</h2>
          <p class="login-subtitle">权限管理系统</p>
        </div>
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-width="0"
          @submit.prevent="handleLogin"
        >
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              :prefix-icon="User"
              size="large"
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              :prefix-icon="Lock"
              size="large"
              show-password
            />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              size="large"
              :loading="loading"
              class="login-btn"
              native-type="submit"
            >
              登 录
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message ?? '登录失败，请检查用户名和密码')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: var(--color-bg-page);
  position: relative;
  overflow: hidden;
}

/* Soft decorative blobs */
.login-container::before,
.login-container::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
  z-index: 0;
}

.login-container::before {
  width: 400px;
  height: 400px;
  background: var(--color-primary-light);
  top: -100px;
  right: -100px;
}

.login-container::after {
  width: 300px;
  height: 300px;
  background: #dbeafe;
  bottom: -80px;
  left: -80px;
}

.login-card-wrapper {
  position: relative;
  z-index: 1;
}

.login-card {
  width: 420px;
  padding: var(--space-xl) var(--space-lg);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
  border: 1px solid rgba(228, 236, 252, 0.8);
  background: var(--color-bg-card);
}

.login-header {
  text-align: center;
  margin-bottom: var(--space-xl);
}

.login-logo {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg);
  background: var(--color-primary);
  color: #fff;
  font-size: 28px;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto var(--space-md);
  box-shadow: var(--shadow-primary);
}

.login-title {
  margin: 0 0 var(--space-xs) 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-2xl);
  font-weight: 700;
  letter-spacing: -0.02em;
}

.login-subtitle {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-base);
}

.login-btn {
  width: 100%;
  height: 48px;
  font-size: var(--font-size-lg);
  font-weight: 600;
  border-radius: var(--radius-md);
}
</style>
