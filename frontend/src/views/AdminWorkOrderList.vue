<template>
  <div class="admin-workorder-list">
    <div class="page-header">
      <h2>全部工单管理</h2>
      <p class="subtitle">审批、执行和管理所有SQL变更工单</p>
    </div>

    <el-card shadow="hover" class="main-card">
      <template #header>
        <div class="card-header">
          <span class="card-title"><i class="el-icon-document"></i> 工单列表</span>
        </div>
      </template>

      <el-form inline @submit.prevent="fetchList" class="search-form">
        <el-form-item label="标题">
          <el-input v-model="search.title" placeholder="搜索工单标题" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item label="提交人">
          <el-input v-model="search.username" placeholder="输入用户名" clearable style="width: 140px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="search.status" placeholder="选择状态" clearable style="width: 150px">
            <el-option label="全部状态" value="" />
            <el-option label="待审核" value="pending" />
            <el-option label="组长通过" value="leader_approved" />
            <el-option label="DBA通过" value="dba_approved" />
            <el-option label="已执行" value="executed" />
            <el-option label="已驳回" value="leader_rejected" />
            <el-option label="自动驳回" value="auto_rejected" />
            <el-option label="执行失败" value="failed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList" icon="el-icon-search">搜索</el-button>
          <el-button @click="resetSearch" icon="el-icon-refresh">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="batch-actions" v-if="selectedIds.length > 0">
        <el-button type="danger" size="small" icon="el-icon-delete" @click="batchDelete">
          批量删除选中 ({{ selectedIds.length }})
        </el-button>
      </div>

      <el-table :data="workOrders" stripe v-loading="loading" class="work-table" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column prop="title" label="工单标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="target_db" label="目标数据库" width="180" />
        <el-table-column prop="creator_id" label="提交人ID" width="90" align="center" />
        <el-table-column prop="status" label="状态" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" width="170" />
        <el-table-column prop="scheduled_time" label="预约执行" width="160" />
        <el-table-column label="操作" width="280" fixed="right" align="center">
          <template #default="{ row }">
            <el-button size="small" link @click="viewDetail(row.id)">查看</el-button>
            <el-button v-if="canApprove(row)" size="small" link type="primary" @click="approve(row.id)">通过</el-button>
            <el-button v-if="canApprove(row)" size="small" link type="danger" @click="reject(row.id)">驳回</el-button>
            <el-button v-if="canExecute(row)" size="small" link type="success" @click="execute(row.id)">执行</el-button>
            <el-button size="small" link type="danger" @click="deleteWorkOrder(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          @current-change="fetchList"
          layout="total, sizes, prev, pager, next, jumper"
          :page-sizes="[10, 20, 50, 100]"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAdminWorkOrderList, leaderApprove, dbaApprove, rejectWorkOrder, executeWorkOrder, deleteAdminWorkOrder } from '../api/index'

const router = useRouter()
const workOrders = ref([])
const loading = ref(false)
const search = reactive({ title: '', username: '', status: 'pending' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const selectedIds = ref([])

const handleSelectionChange = (selection) => {
  selectedIds.value = selection.map(item => item.id)
}

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

const canExecute = (row) => {
  return row.status === 'dba_approved' || row.status === 'failed'
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

const deleteWorkOrder = async (id) => {
  try {
    await ElMessageBox.confirm('确认删除此工单? 删除后无法恢复', '删除确认', { type: 'danger' })
    await deleteAdminWorkOrder(id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

	const batchDelete = async () => {
	  if (selectedIds.value.length === 0) {
	    ElMessage.warning('请先选择要删除的工单')
	    return
	  }
	  try {
	    await ElMessageBox.confirm(
	      `确认删除选中的 ${selectedIds.value.length} 个工单?\n删除后无法恢复！`,
	      '批量删除确认',
	      { type: 'danger' }
	    )
	    // 逐个删除
	    await Promise.all(selectedIds.value.map(id => deleteAdminWorkOrder(id)))
	    ElMessage.success(`成功删除 ${selectedIds.value.length} 个工单`)
	    selectedIds.value = []
	    fetchList()
	  } catch (error) {
	    if (error !== 'cancel') {
	      ElMessage.error('批量删除失败')
	    }
	  }
	}

onMounted(fetchList)
</script>

<style scoped>
.admin-workorder-list {
  padding: 20px;
  min-height: 100vh;
  background-color: #f5f7fa;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  color: #303133;
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 8px 0;
}

.page-header .subtitle {
  color: #909399;
  font-size: 14px;
  margin: 0;
}

.main-card {
  border-radius: 8px;
}

.card-header {
  display: flex;
  align-items: center;
}

.card-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.search-form {
  margin-bottom: 20px;
  padding: 10px 0;
}

.batch-actions {
  margin-bottom: 10px;
  padding: 10px 0;
}

.work-table {
  margin-top: 10px;
}

.pagination-wrapper {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-pagination) {
  justify-content: flex-end;
}
</style>