<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2>SQL上线管理系统</h2>
      <el-form :model="form" @submit.prevent="handleLogin">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" size="large" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width: 100%" size="large">
            登录
          </el-button>
        </el-form-item>
      </el-form>
      <div class="footer-link">
        <router-link to="/register">管理员注册</router-link>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login } from '../api/index'

const router = useRouter()
const form = reactive({ username: '', password: '' })
const loading = ref(false)

const handleLogin = async () => {
  if (!form.username || !form.password) {
    ElMessage.error('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    const { data } = await login(form)
    localStorage.setItem('token', data.token)
    localStorage.setItem('username', form.username)
    ElMessage.success('登录成功')
    router.push('/workorders')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  width: 400px;
  padding: 20px;
}
.login-card h2 {
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