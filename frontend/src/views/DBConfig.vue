<template>
  <div class="db-config">
    <el-card>
      <template #header>
        <h3>目标数据库配置</h3>
      </template>
      
      <el-form :model="form" label-width="120px" style="max-width: 600px">
        <el-form-item label="配置名称">
          <el-input v-model="form.name" placeholder="例如：生产库" />
        </el-form-item>
        <el-form-item label="数据库地址">
          <el-input v-model="form.host" placeholder="例如：192.168.1.100" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input v-model="form.port" placeholder="1688" type="number" />
        </el-form-item>
        <el-form-item label="实例名">
          <el-input v-model="form.instance" placeholder="默认：YASHANDB" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="数据库用户名" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" placeholder="数据库密码" type="password" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSave">保存配置</el-button>
        </el-form-item>
      </el-form>
      
      <el-divider />
      
       <el-table :data="configs" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="配置名称" />
        <el-table-column prop="host" label="数据库地址" />
        <el-table-column prop="port" label="端口" width="80" />
        <el-table-column prop="instance" label="实例名" width="120" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button link type="danger" @click="remove(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDBConfig, saveDBConfig, deleteDBConfig } from '@/api/index'

const configs = ref([])
const form = reactive({ name: '', host: '', port: '1688', instance: 'YASHANDB', username: '', password: '' })

const fetchList = async () => {
  try {
    const { data } = await getDBConfig()
    configs.value = data
  } catch (error) {
    ElMessage.error('获取配置失败')
  }
}

const handleSave = async () => {
  if (!form.name || !form.host || !form.port || !form.instance || !form.username || !form.password) {
    ElMessage.error('请填写所有字段')
    return
  }
  try {
    await saveDBConfig({
      name: form.name,
      host: form.host,
      port: parseInt(form.port),
      instance: form.instance,
      username: form.username,
      password: form.password,
      is_active: true
    })
    ElMessage.success('保存成功')
    form.name = ''
    form.host = ''
    form.port = '1688'
    form.instance = 'YASHANDB'
    form.username = ''
    form.password = ''
    fetchList()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const remove = async (id) => {
  try {
    await ElMessageBox.confirm('确认删除此配置?')
    await deleteDBConfig(id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(fetchList)
</script>

<style scoped>
h3 {
  margin: 0;
}
</style>