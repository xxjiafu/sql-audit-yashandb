<template>
  <div class="admin-workorder-list">
    <el-card>
      <template #header>
        <h3>全部工单</h3>
      </template>
      
      <el-form inline @submit.prevent="fetchList" class="search-form">
        <el-form-item label="标题">
          <el-input v-model="search.title" placeholder="工单标题" clearable style="width: 150px" />
        </el-form-item>
        <el-form-item label="提交人">
          <el-input v-model="search.username" placeholder="用户名" clearable style="width: 120px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="search.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="待审核" value="pending" />
            <el-option label="组长通过" value="leader_approved" />
            <el-option label="DBA通过" value="dba_approved" />
            <el-option label="已执行" value="executed" />
            <el-option label="已驳回" value="leader_rejected" />
            <el-option label="自动驳回" value="auto_rejected" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
      
      <el-table :data="workOrders" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="标题" />
        <el-table-column prop="creator_id" label="提交人" width="100" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" width="180" />
        <el-table-column prop="scheduled_time" label="预约时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link @click="viewDetail(row.id)">查看</el-button>
            <el-button v-if="canApprove(row)" link type="primary" @click="approve(row.id)">通过</el-button>
            <el-button v-if="canApprove(row)" link type="danger" @click="reject(row.id)">驳回</el-button>
            <el-button v-if="row.status === 'dba_approved'" link type="success" @click="execute(row.id)">执行</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-pagination
        v-model:current-page="pagination.page"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @current-change="fetchList"
        layout="total, prev, pager, next"
        style="margin-top: 16px; justify-content: center"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAdminWorkOrderList, leaderApprove, dbaApprove, rejectWorkOrder, executeWorkOrder } from '../api/index'

const router = useRouter()
const workOrders = ref([])
const loading = ref(false)
const search = reactive({ title: '', username: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const fetchList = async () => {
  loading.value = true
  try {
    const params = { page: pagination.page, page_size: pagination.pageSize, ...search }
    Object.keys(params).forEach(key => !params[key] && delete params[key])
    const { data } = await getAdminWorkOrderList(params)
    workOrders.value = data.data
    pagination.total = data.total
  } catch (error) {
    ElMessage.error('获取工单列表失败')
  } finally {
    loading.value = false
  }
}

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

const canApprove = (row) => {
  return row.status === 'pending' || row.status === 'leader_approved'
}

const viewDetail = (id) => router.push(`/admin/workorders/${id}`)

const approve = async (id) => {
  try {
    const workOrder = workOrders.value.find(w => w.id === id)
    if (workOrder.status === 'pending') {
      await leaderApprove(id)
      ElMessage.success('组长审核通过')
    } else if (workOrder.status === 'leader_approved') {
      await dbaApprove(id)
      ElMessage.success('DBA审核通过')
    }
    fetchList()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '审核失败')
  }
}

const reject = async (id) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入驳回原因', '驳回工单', {
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    })
    await rejectWorkOrder(id, { reason: value })
    ElMessage.success('已驳回')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '驳回失败')
    }
  }
}

const execute = async (id) => {
  try {
    await ElMessageBox.confirm('确认执行此工单?', '执行确认', { type: 'warning' })
    await executeWorkOrder(id)
    ElMessage.success('执行成功')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '执行失败')
    }
  }
}

const resetSearch = () => {
  search.title = ''
  search.username = ''
  search.status = ''
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.search-form {
  margin-bottom: 16px;
}
</style>