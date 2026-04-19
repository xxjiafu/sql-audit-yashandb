<template>
  <div class="admin-workorder-detail">
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
        <el-descriptions-item label="提交人">{{ workOrder.creator_id }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(workOrder.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="预约时间">{{ workOrder.scheduled_time ? formatDate(workOrder.scheduled_time) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="执行用户">{{ workOrder.execution_user || '-' }}</el-descriptions-item>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAdminWorkOrder, leaderApprove, dbaApprove, rejectWorkOrder, executeWorkOrder } from '../api/index'

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

const canApprove = (row) => row.status === 'pending' || row.status === 'leader_approved'

const canExecute = (row) => {
  // 已成功执行、执行中、已自动驳回的工单不允许再执行，执行失败允许重新执行
  return row.status !== 'executed' && row.status !== 'executing' && row.status !== 'auto_rejected'
}

const fetchDetail = async () => {
  loading.value = true
  try {
    const res = await getAdminWorkOrder(route.params.id)
    workOrder.value = res.data.data
  } catch (error) {
    ElMessage.error('获取工单详情失败')
  } finally {
    loading.value = false
  }
}

const handleApprove = async () => {
  try {
    if (workOrder.value.status === 'pending') {
      await leaderApprove(workOrder.value.id)
      ElMessage.success('组长审核通过')
    } else {
      await dbaApprove(workOrder.value.id)
      ElMessage.success('DBA审核通过')
    }
    router.push('/admin/workorders')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '审核失败')
  }
}

const handleReject = async () => {
  try {
    const { value } = await ElMessageBox.prompt('请输入驳回原因', '驳回工单', { type: 'warning' })
    await rejectWorkOrder(workOrder.value.id, { reason: value })
    ElMessage.success('已驳回')
    router.push('/admin/workorders')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '驳回失败')
    }
  }
}

const handleExecute = async () => {
  try {
    await ElMessageBox.confirm('确认执行此工单?', '执行确认', { type: 'warning' })
    await executeWorkOrder(workOrder.value.id)
    ElMessage.success('执行成功')
    fetchDetail()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '执行失败')
      fetchDetail()
    }
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
.error-item {
  color: #f56c6c;
  margin: 4px 0;
}
.warning-item {
  color: #e6a23c;
  margin: 4px 0;
}
.action-buttons {
  margin-top: 20px;
  text-align: center;
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

.error-item {
  color: #f56c6c;
  margin: 4px 0;
}

.warning-item {
  color: #e6a23c;
  margin: 4px 0;
}

.file-name {
  margin-left: 10px;
  color: #909399;
  font-size: 14px;
}

.file-area {
  padding: 8px 0;
}
</style>
