<template>
  <div class="workorder-create">
    <el-card>
      <h3>{{ isEdit ? '编辑工单' : '创建SQL工单' }}</h3>
      <el-form :model="form" label-width="100px" style="max-width: 800px">
        <el-form-item label="工单标题" required>
          <el-input v-model="form.title" placeholder="请输入工单标题" />
        </el-form-item>
        
        <el-form-item label="SQL内容">
          <el-input
            v-model="form.sql_content"
            type="textarea"
            :rows="10"
            placeholder="请输入SQL语句"
          />
        </el-form-item>
        
        <el-form-item label="或上传文件">
          <el-upload
            :action="uploadUrl"
            :on-success="handleUploadSuccess"
            :before-upload="beforeUpload"
            :show-file-list="false"
          >
            <el-button>选择SQL文件</el-button>
            <span v-if="form.file_url" style="margin-left: 10px">已上传: {{ form.file_url }}</span>
          </el-upload>
        </el-form-item>
        
        <el-form-item label="预约时间">
          <el-date-picker
            v-model="form.scheduled_time"
            type="datetime"
            placeholder="选择预约执行时间"
            :disabled-date="disabledDate"
            style="width: 100%"
          />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="loading">
            {{ isEdit ? '保存' : '提交' }}
          </el-button>
          <el-button @click="$router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createWorkOrder, updateWorkOrder, getWorkOrder } from '../api/index'

const router = useRouter()
const route = useRoute()

const form = reactive({
  title: '',
  sql_content: '',
  file_url: '',
  scheduled_time: null
})
const loading = ref(false)
const uploadUrl = '/api/upload'

const isEdit = computed(() => !!route.params.id)

const loadWorkOrder = async () => {
  try {
    const { data } = await getWorkOrder(route.params.id)
    form.title = data.title
    form.sql_content = data.sql_content || ''
    form.file_url = data.file_url || ''
    form.scheduled_time = data.scheduled_time ? new Date(data.scheduled_time) : null
  } catch (error) {
    ElMessage.error('获取工单信息失败')
  }
}

const beforeUpload = (file) => {
  if (!file.name.endsWith('.sql')) {
    ElMessage.error('只能上传.sql文件')
    return false
  }
}

const handleUploadSuccess = (response) => {
  form.file_url = response.data?.url || response.url
  ElMessage.success('文件上传成功')
}

const disabledDate = (time) => {
  return time.getTime() < Date.now() - 86400000
}

const handleSubmit = async () => {
  if (!form.title) {
    ElMessage.error('请输入工单标题')
    return
  }
  if (!form.sql_content && !form.file_url) {
    ElMessage.error('请输入SQL内容或上传文件')
    return
  }
  
  loading.value = true
  try {
    const data = {
      title: form.title,
      sql_content: form.sql_content,
      file_url: form.file_url,
      scheduled_time: form.scheduled_time ? formatDate(form.scheduled_time) : null
    }
    
    if (isEdit.value) {
      await updateWorkOrder(route.params.id, data)
      ElMessage.success('工单更新成功')
    } else {
      await createWorkOrder(data)
      ElMessage.success('工单创建成功')
    }
    router.push('/workorders')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || (isEdit.value ? '更新失败' : '创建失败'))
  } finally {
    loading.value = false
  }
}

const formatDate = (date) => {
  const d = new Date(date)
  return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' }).replace(/\//g, '-')
}

onMounted(() => {
  if (isEdit.value) {
    loadWorkOrder()
  }
})
</script>

<style scoped>
.workorder-create h3 {
  margin-bottom: 20px;
}
</style>
