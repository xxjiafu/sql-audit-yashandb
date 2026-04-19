<template>
  <div class="layout-container">
    <el-header class="header">
      <div class="logo">SQL审核系统</div>
      <div class="header-right">
        <span class="username">{{ currentUser?.username }}</span>
        <el-button type="text" @click="logout">退出登录</el-button>
      </div>
    </el-header>
    <el-container>
      <el-aside width="200px" class="aside">
        <el-menu
          :default-active="$route.path"
          router
          class="aside-menu"
        >
          <el-menu-item index="/workorders">
            <el-icon><Document /></el-icon>
            <span>我的工单</span>
          </el-menu-item>
          <el-menu-item index="/workorders/create">
            <el-icon><Edit /></el-icon>
            <span>创建工单</span>
          </el-menu-item>
          <el-menu-item index="/admin/db-config">
            <el-icon><Connection /></el-icon>
            <span>数据库配置</span>
          </el-menu-item>
          <el-sub-menu index="/admin" v-if="isAdmin">
            <template #title>
              <el-icon><Setting /></el-icon>
              <span>管理中心</span>
              <el-badge :value="pendingCount" :hidden="pendingCount === 0" class="pending-badge" />
            </template>
            <el-menu-item index="/admin/workorders">所有工单</el-menu-item>
            <el-menu-item index="/admin/users">用户管理</el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-aside>
      <el-main class="main">
        <el-breadcrumb class="breadcrumb">
          <el-breadcrumb-item v-for="item in breadcrumbList" :key="item.path" :to="item.path">
            {{ item.title }}
          </el-breadcrumb-item>
        </el-breadcrumb>
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>
<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Document, Edit, Setting, Connection } from '@element-plus/icons-vue'
import { getCurrentUser, getAdminWorkOrderList } from '@/api/index'
const router = useRouter()
const route = useRoute()
const currentUser = ref(null)
const pendingCount = ref(0)
const isAdmin = computed(() => {
  return currentUser.value?.role === 'admin' || currentUser.value?.role === 'dba' || currentUser.value?.role === 'leader'
})
const breadcrumbList = computed(() => {
  const paths = route.path.split('/').filter(p => p)
  const list = []
  list.push({ path: '/workorders', title: '首页' })
  
  if (paths[0] === 'workorders') {
    if (paths.length === 1) {
      list.push({ path: '/workorders', title: '我的工单' })
    } else if (paths[1] === 'create') {
      list.push({ path: '/workorders', title: '我的工单' })
      list.push({ path: '/workorders/create', title: '创建工单' })
    } else {
      list.push({ path: '/workorders', title: '我的工单' })
      list.push({ path: route.path, title: '工单详情' })
    }
  } else if (paths[0] === 'admin') {
    if (paths[1] === 'workorders') {
      list.push({ path: '/admin/workorders', title: '所有工单' })
      if (paths.length > 2) {
        list.push({ path: route.path, title: '工单详情' })
      }
    }
  } else if (paths[0] === 'db-config') {
    list.push({ path: '/db-config', title: '数据库配置' })
  }
  return list
})
const loadCurrentUser = async () => {
  try {
    const { data } = await getCurrentUser()
    currentUser.value = data
    if (isAdmin.value) {
      fetchPendingCount()
    }
  } catch (error) {
    ElMessage.error('登录已过期')
    router.push('/login')
  }
}
const fetchPendingCount = async () => {
  try {
    const { data } = await getAdminWorkOrderList({ status: 'pending', page: 1, page_size: 1 })
    pendingCount.value = data.total || 0
  } catch (e) {}
}
const logout = () => {
  localStorage.removeItem('token')
  ElMessage.success('已退出登录')
  router.push('/login')
}
onMounted(() => {
  loadCurrentUser()
  // 每30秒刷新待审核数量
  setInterval(() => {
    if (isAdmin.value) fetchPendingCount()
  }, 30000)
})
</script>
<style scoped>
.layout-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background: linear-gradient(90deg, #2d52a0 0%, #667eea 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 60px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.logo {
  font-size: 20px;
  font-weight: 600;
  letter-spacing: 1px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.username {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.9);
}

.header-right button {
  color: white;
  background-color: rgba(255, 255, 255, 0.15);
  border: none;
  transition: background-color 0.3s;
}

.header-right button:hover {
  background-color: rgba(255, 255, 255, 0.25);
}

.aside {
  background-color: #f8fafc;
  border-right: 1px solid #e4e7ed;
}

.aside-menu {
  border-right: none;
  height: 100%;
}

.aside-menu :deep(.el-menu-item:hover) {
  background-color: #ecf5ff;
}

.main {
  background-color: #f5f7fa;
  padding: 20px;
  min-height: calc(100vh - 60px);
}

.breadcrumb {
  margin-bottom: 20px;
  padding: 8px 0;
}

.pending-badge {
  margin-left: 8px;
}
</style>