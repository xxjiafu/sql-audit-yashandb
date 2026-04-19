<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1>SQL审计发布平台</h1>
        <p>安全 · 可控 · 可审计</p>
      </div>
      <el-card shadow="always" class="login-card">
        <el-form :model="form" @submit.prevent="handleLogin">
          <el-form-item>
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              size="large"
              prefix-icon="el-icon-user"
            />
          </el-form-item>
          <el-form-item>
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              size="large"
              prefix-icon="el-icon-lock"
            />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              native-type="submit"
              :loading="loading"
              style="width: 100%"
              size="large"
            >
              登 录
            </el-button>
          </el-form-item>
        </el-form>
        <div class="footer-link">
          <router-link to="/register">没有账号？点击注册</router-link>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login, getCurrentUser } from '../api/index'

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
    // 获取用户角色，根据角色跳转
    try {
      const userRes = await getCurrentUser()
      const role = userRes.data.role
      if (['admin', 'leader', 'dba'].includes(role)) {
        // 管理员/组长/DBA跳转到全部工单页面，查看待审核工单
        router.push('/admin/workorders')
      } else {
        router.push('/workorders')
      }
    } catch (e) {
      router.push('/workorders')
    }
    ElMessage.success('登录成功')
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
  background: linear-gradient(135deg, #8ec5fc 0%, #e0c3fc 50%, #a8edea 100%);
  padding: 20px;
}

.login-box {
  width: 420px;
}

.login-header {
  text-align: center;
  color: #fff;
  margin-bottom: 30px;
}

.login-header h1 {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 10px 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.login-header p {
  font-size: 16px;
  opacity: 0.9;
  margin: 0;
}

.login-card {
  border-radius: 12px;
  padding: 10px 30px 30px;
}

.footer-link {
  text-align: center;
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid #f0f0f0;
}

.footer-link a {
  color: #409eff;
  text-decoration: none;
  font-size: 14px;
}

.footer-link a:hover {
  text-decoration: underline;
}
</style>