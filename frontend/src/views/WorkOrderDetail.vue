<template>
  <div class="workorder-detail">
    <el-card v-loading="loading">
      <template #header>
        <div class="header">
          <h3>工单详情</h3>
          <el-button @click="$router.back()">返回</el-button>
        </div>
      </template>
      
      <el-descriptions :column="2" border v-if="workOrder">
        <el-descriptions-item label="ID">{{ workOrder.id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(workOrder.status)">{{ getStatusText(workOrder.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ workOrder.title }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ workOrder.created_at }}</el-descriptions-item>
        <el-descriptions-item label="预约时间">{{ workOrder.scheduled_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="SQL内容" :span="2">
          <pre class="sql-content">{{ workOrder.sql_content }}</pre>
        </el-descriptions-item>
        <el-descriptions-item v-if="workOrder.reject_reason" label="驳回原因" :span="2">
          <span style="color: #f56c6c">{{ workOrder.reject_reason }}</span>
        </el-descriptions-item>
        <el-descriptions-item v-if="workOrder.execution_result" label="执行结果" :span="2">
          <pre class="execution-result">{{ workOrder.execution_result }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getWorkOrder } from '../api/index'

const route = useRoute()
const workOrder = ref(null)
const loading = ref(false)

const getStatusType = (status) => {
  const map = {
    pending: 'info',
    auto_rejected: 'danger',
    leader_rejected: 'danger',
    dba_rejected: 'danger',
    leader_approved: 'success',
    dba_approved: 'success',
    executing: 'warning',
    executed: 'success',
    failed: 'danger'
  }
  return map[status] || 'info'
}

const getStatusText = (status) => {
  const map = {
    pending: '待审核',
    auto_rejected: '自动驳回',
    leader_rejected: '组长驳回',
    dba_rejected: 'DBA驳回',
    leader_approved: '组长通过',
    dba_approved: 'DBA通过',
    executing: '执行中',
    executed: '已执行',
    failed: '执行失败'
  }
  return map[status] || status
}

const fetchDetail = async () => {
  loading.value = true
  try {
    const { data } = await getWorkOrder(route.params.id)
    workOrder.value = data
  } catch (error) {
    ElMessage.error('获取工单详情失败')
  } finally {
    loading.value = false
  }
}

onMounted(fetchDetail)
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header h3 {
  margin: 0;
}
.sql-content, .execution-result {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  max-height: 300px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: monospace;
}
</style>
