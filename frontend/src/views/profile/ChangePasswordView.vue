<template>
  <div class="change-password-view">
    <div class="page-header">
      <h2>修改密码</h2>
    </div>

    <el-card class="password-card">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        style="max-width: 500px"
      >
        <el-form-item label="旧密码" prop="old_password">
          <el-input
            v-model="form.old_password"
            type="password"
            show-password
            placeholder="请输入旧密码"
            autofocus
          />
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input
            v-model="form.new_password"
            type="password"
            show-password
            placeholder="请输入新密码"
          />
        </el-form-item>
        <el-form-item label="确认新密码" prop="confirm_password">
          <el-input
            v-model="form.confirm_password"
            type="password"
            show-password
            placeholder="请再次输入新密码"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            修改密码
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { changePassword } from '@/api/auth'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const submitting = ref(false)

const formRef = ref<FormInstance>()
const form = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const validateConfirmPassword = (rule: any, value: string, callback: any) => {
  if (value !== form.value.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  old_password: [
    { required: true, message: '请输入旧密码', trigger: 'blur' }
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码长度不能少于 8 位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

async function handleSubmit() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      await changePassword({
        old_password: form.value.old_password,
        new_password: form.value.new_password
      })
      ElMessage.success('密码修改成功，请重新登录')
      await authStore.logout()
      router.push('/login')
    } catch (error: any) {
      ElMessage.error(error.message || '密码修改失败')
    } finally {
      submitting.value = false
    }
  })
}
</script>

<style scoped>
.change-password-view {
  padding: 0;
}

.password-card {
  max-width: 600px;
}
</style>
