<template>
  <div class="workorder-list">
    <el-card>
      <template #header>
        <div class="header">
          <h3>我的工单</h3>
          <el-button type="primary" @click="$router.push('/workorders/create')">
            创建工单
          </el-button>
        </div>
      </template>
      
      <el-table :data="workOrders" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="标题" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button link @click="viewDetail(row.id)">查看</el-button>
            <el-button link v-if="row.status === 'pending'" @click="editWorkOrder(row.id)">编辑</el-button>
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
import { ElMessage } from 'element-plus'
import { getWorkOrderList } from '../api/index'

const router = useRouter()
const workOrders = ref([])
const loading = ref(false)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const fetchList = async () => {
  loading.value = true
  try {
    const { data } = await getWorkOrderList(pagination.page, pagination.pageSize)
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

const viewDetail = (id) => router.push(`/workorders/${id}`)
const editWorkOrder = (id) => router.push(`/workorders/${id}/edit`)

onMounted(fetchList)
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
</style>
