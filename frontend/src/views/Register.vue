<template>
  <div class="register-container">
    <el-card class="register-card">
      <h2>管理员注册</h2>
      <el-form :model="form" @submit.prevent="handleRegister">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.confirmPassword" type="password" placeholder="确认密码" size="large" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width: 100%" size="large">
            注册
          </el-button>
        </el-form-item>
      </el-form>
      <div class="footer-link">
        <router-link to="/login">返回登录</router-link>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { register } from '../api/index'

const router = useRouter()
const form = reactive({ username: '', password: '', confirmPassword: '' })
const loading = ref(false)

const handleRegister = async () => {
  if (!form.username || !form.password) {
    ElMessage.error('请输入用户名和密码')
    return
  }
  if (form.password !== form.confirmPassword) {
    ElMessage.error('两次输入的密码不一致')
    return
  }
  if (form.password.length < 6) {
    ElMessage.error('密码长度不能少于6位')
    return
  }
  loading.value = true
  try {
    await register({ username: form.username, password: form.password })
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.register-card {
  width: 400px;
  padding: 20px;
}
.register-card h2 {
  text-align: center;
  margin-bottom: 24px;
  color: #333;
}
.footer-link {
  text-align: center;
  margin-top: 16px;
}
.footer-link a {
  color: #409eff;
  text-decoration: none;
}
.footer-link a:hover {
  text-decoration: underline;
}
</style>