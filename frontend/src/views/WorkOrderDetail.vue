<template>
  <div class="workorder-detail">
    <el-card v-loading="loading">
      <template #header>
        <div class="header">
          <h3>工单详情</h3>
          <div>
            <el-button type="primary" @click="goEdit">修改</el-button>
            <el-button @click="$router.back()">返回</el-button>
          </div>
        </div>
      </template>
      
      <el-descriptions :column="2" border v-if="workOrder">
        <el-descriptions-item label="ID">{{ workOrder.id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(workOrder.status)">{{ getStatusText(workOrder.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ workOrder.title }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(workOrder.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="预约时间">{{ workOrder.scheduled_time ? formatDate(workOrder.scheduled_time) : '-' }}</el-descriptions-item>
        <el-descriptions-item v-if="workOrder.execution_user" label="执行用户">{{ workOrder.execution_user || '-' }}</el-descriptions-item>
        <el-descriptions-item v-if="workOrder.executed_at" label="执行时间">{{ formatDate(workOrder.executed_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="workOrder.sql_content" label="SQL内容" :span="2">
          <pre class="sql-content">{{ workOrder.sql_content }}</pre>
        </el-descriptions-item>
        <el-descriptions-item v-if="workOrder.auto_check_result" label="自动审核结果" :span="2">
          <div v-if="autoCheckResult">
            <el-tag v-if="!autoCheckResult.passed" type="danger" style="margin-bottom: 8px">未通过</el-tag>
            <el-tag v-else type="success" style="margin-bottom: 8px">通过</el-tag>
            <div v-if="autoCheckResult.errors?.length">
              <div v-for="(err, i) in autoCheckResult.errors" :key="i" class="error-item">{{ err }}</div>
            </div>
            <div v-if="autoCheckResult.warnings?.length">
              <div v-for="(warn, i) in autoCheckResult.warnings" :key="i" class="warning-item">{{ warn }}</div>
            </div>
          </div>
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
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getWorkOrder } from '../api/index'

const route = useRoute()
const router = useRouter()
const workOrder = ref(null)
const loading = ref(false)
const autoCheckResult = computed(() => {
  if (!workOrder.value?.auto_check_result) return null
  try {
    return JSON.parse(workOrder.value.auto_check_result)
  } catch {
    return null
  }
})

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


const formatDate = (isoStr) => {
  if (!isoStr) return ''
  const d = new Date(isoStr)
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

onMounted(fetchDetail)

const goEdit = () => {
  router.push(`/workorders/${workOrder.value.id}/edit`)
}
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
  max-height: 400px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.file-name {
  margin-left: 10px;
  color: #909399;
  font-size: 14px;
}

.file-area {
  padding: 8px 0;
}
.error-item {
  color: #f56c6c;
  margin: 4px 0;
}
.warning-item {
  color: #e6a23c;
  margin: 4px 0;
}
</style>
