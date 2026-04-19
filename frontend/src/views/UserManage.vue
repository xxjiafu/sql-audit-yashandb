<template>
  <div class="user-manage">
    <el-card>
      <template #header>
        <div class="header">
          <h3>用户管理</h3>
          <el-button type="primary" @click="showAddDialog = true">新增用户</el-button>
        </div>
      </template>

      <el-table :data="users" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="role" label="角色">
          <template #default="{ row }">
            <el-tag :type="getRoleType(row.role)">{{ getRoleText(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="280">
          <template #default="{ row }">
            <el-button link type="primary" @click="editRole(row)">修改角色</el-button>
            <el-button link type="primary" @click="changePassword(row)">修改密码</el-button>
            <el-button link type="danger" @click="deleteUser(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showAddDialog" title="新增用户" width="400px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role" placeholder="选择角色">
            <el-option label="开发人员" value="developer" />
            <el-option label="组长" value="leader" />
            <el-option label="DBA" value="dba" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAdd">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showEditDialog" title="修改角色" width="400px">
      <el-form label-width="80px">
        <el-form-item label="用户名">
          <el-input :value="editUser?.username" disabled />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editRoleValue" placeholder="选择角色">
            <el-option label="开发人员" value="developer" />
            <el-option label="组长" value="leader" />
            <el-option label="DBA" value="dba" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleEditRole">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showPasswordDialog" title="修改密码" width="400px">
      <el-form label-width="80px">
        <el-form-item label="用户名">
          <el-input :value="editUser?.username" disabled />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="newPassword" type="password" placeholder="请输入新密码（至少6位）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">取消</el-button>
        <el-button type="primary" @click="handleChangePassword">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserList, createUser, updateUserRole, updateUserPassword, deleteUser as apiDeleteUser } from '@/api/index'

const users = ref([])
const loading = ref(false)
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const showPasswordDialog = ref(false)
const editUser = ref(null)
const editRoleValue = ref('')
const newPassword = ref('')
const form = ref({ username: '', password: '', role: 'developer' })

const fetchList = async () => {
  loading.value = true
  try {
    const { data } = await getUserList()
    users.value = data
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

const getRoleType = (role) => {
  const map = { admin: 'danger', leader: 'warning', dba: 'success', developer: 'info' }
  return map[role] || 'info'
}

const getRoleText = (role) => {
  const map = { admin: '管理员', leader: '组长', dba: 'DBA', developer: '开发人员' }
  return map[role] || role
}

const handleAdd = async () => {
  if (!form.value.username || !form.value.password) {
    ElMessage.error('请填写用户名和密码')
    return
  }
  try {
    await createUser(form.value)
    ElMessage.success('创建成功')
    showAddDialog.value = false
    form.value = { username: '', password: '', role: 'developer' }
    fetchList()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '创建失败')
  }
}

const editRole = (row) => {
  editUser.value = row
  editRoleValue.value = row.role
  showEditDialog.value = true
}

const handleEditRole = async () => {
  try {
    await updateUserRole(editUser.value.id, editRoleValue.value)
    ElMessage.success('修改成功')
    showEditDialog.value = false
    fetchList()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '修改失败')
  }
}

const changePassword = (row) => {
  editUser.value = row
  newPassword.value = ''
  showPasswordDialog.value = true
}

const handleChangePassword = async () => {
  if (!newPassword.value || newPassword.value.length < 6) {
    ElMessage.error('请输入新密码，至少6个字符')
    return
  }
  try {
    await updateUserPassword(editUser.value.id, newPassword.value)
    ElMessage.success('密码修改成功')
    showPasswordDialog.value = false
    fetchList()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '修改失败')
  }
}

const deleteUser = async (id) => {
  try {
    await ElMessageBox.confirm('确认删除此用户?')
    await apiDeleteUser(id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

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