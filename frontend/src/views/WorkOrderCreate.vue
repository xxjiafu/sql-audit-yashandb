<template>
  <div class="workorder-create">
    <div class="page-header">
      <h2>{{ isEdit ? '编辑工单' : '创建SQL工单' }}</h2>
      <p class="subtitle">{{ isEdit ? '修改工单信息' : '提交新的SQL变更工单' }}</p>
    </div>

    <el-card shadow="hover" class="main-card">
      <el-form :model="form" label-width="120px" class="main-form">
        <el-form-item label="工单标题" required>
          <el-input v-model="form.title" placeholder="请输入工单标题" clearable />
        </el-form-item>

        <el-form-item label="SQL内容">
          <div class="sql-editor-container">
            <codemirror
              v-model="form.sql_content"
              :extensions="extensions"
              @update="onSqlChange"
            />
          </div>
          <div class="sql-hint">
            <template v-if="sqlError">
              <el-tag type="danger" class="error-tag">{{ sqlError }}</el-tag>
            </template>
          </div>
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="预约执行时间">
              <el-date-picker
                v-model="form.scheduled_time"
                type="datetime"
                placeholder="选择预约执行时间"
                :disabled-date="disabledDate"
                format="YYYY-MM-DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="目标数据库">
              <el-select v-model="form.target_db_id" placeholder="选择目标数据库" clearable style="width: 100%">
                <el-option v-for="db in dbConfigs" :key="db.id" :label="db.name + (db.is_active ? ' (已激活)' : '')" :value="db.id" />
              </el-select>
              <div class="hint">留空则使用当前激活的数据库配置</div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="执行用户名">
              <el-input v-model="form.execution_user" placeholder="请指定执行SQL的数据库用户名，留空使用数据库配置默认" clearable />
              <div class="hint">需要以不同用户执行时填写，密码使用数据库配置中保存的密码</div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item class="form-buttons">
          <el-button type="primary" @click="handleSubmit" :loading="loading" icon="el-icon-check">
            {{ isEdit ? '保存修改' : '提交工单' }}
          </el-button>
          <el-button @click="$router.back()" icon="el-icon-back">返回</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createWorkOrder, updateWorkOrder, getWorkOrder, getDBConfig } from '../api/index'
import Codemirror from 'vue-codemirror6'
import { sql, PostgreSQL } from '@codemirror/lang-sql'
import { basicSetup } from 'codemirror'

// 基本Oracle SQL关键字用于简单语法检查
const oracleKeywords = new Set([
  'SELECT', 'FROM', 'WHERE', 'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE',
  'CREATE', 'TABLE', 'ALTER', 'DROP', 'INDEX', 'ADD', 'MODIFY', 'CONSTRAINT',
  'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES', 'UNIQUE', 'NOT', 'NULL', 'DEFAULT',
  'GROUP', 'BY', 'ORDER', 'HAVING', 'DISTINCT', 'JOIN', 'INNER', 'LEFT', 'RIGHT',
  'OUTER', 'ON', 'AS', 'AND', 'OR', 'IN', 'BETWEEN', 'LIKE', 'IS', 'NULL',
  'ASC', 'DESC', 'LIMIT', 'OFFSET', 'FETCH', 'FIRST', 'NEXT', 'ROW', 'ROWS',
  'COMMIT', 'ROLLBACK', 'TRUNCATE', 'RENAME', 'GRANT', 'REVOKE', 'SEQUENCE',
  'NEXTVAL', 'CURRVAL', 'SYSDATE', 'TO_DATE', 'TO_CHAR', 'TO_NUMBER', 'COUNT',
  'SUM', 'AVG', 'MAX', 'MIN', 'VARCHAR2', 'NUMBER', 'DATE', 'TIMESTAMP',
  'FLOAT', 'CHAR', 'CLOB', 'BLOB', 'INT', 'INTEGER', 'SMALLINT', 'DECIMAL'
])

// 简单的语法检查 - 检测常见错误
function checkSqlSyntax(sql) {
  if (!sql || sql.trim().length === 0) return null

  const lines = sql.split('\n')
  // 检查缺失分号
  const lastLine = lines[lines.length - 1].trim()
  if (lastLine && !lastLine.endsWith(';') && lines.length > 0 && lastLine.length > 0) {
    return '提示：最后一行SQL建议以分号 ; 结尾'
  }

  // 检查不平衡括号
  const openParen = (sql.match(/\(/g) || []).length
  const closeParen = (sql.match(/\)/g) || []).length
  if (openParen !== closeParen) {
    return `语法警告：括号不匹配，${openParen}个开括号，${closeParen}个闭括号`
  }

  // 检查常见拼写错误
  const lowerSql = sql.toUpperCase()
  const commonMistakes = {
    'SELEC': 'SELECT',
    'WHRE': 'WHERE',
    'FORM': 'FROM',
    'UPADTE': 'UPDATE',
    'INSRET': 'INSERT',
    'CREAT': 'CREATE',
    'DELET': 'DELETE',
    'VALUSE': 'VALUES'
  }

  for (const [mistake, correct] of Object.entries(commonMistakes)) {
    if (lowerSql.includes(mistake)) {
      return `拼写提示：检测到 "${mistake}" → 应为 "${correct}"`
    }
  }

  return null
}

const router = useRouter()
const route = useRoute()

// CodeMirror配置 - IDEA风格浅色主题
const extensions = [
  basicSetup,
  sql({
    dialect: PostgreSQL
  })
]

const sqlError = ref(null)

// 监听SQL变化进行语法检查 - update event gives ViewUpdate, need to get value from editor
function onSqlChange(viewUpdate) {
  const value = viewUpdate.state.doc.toString()
  form.sql_content = value
  sqlError.value = checkSqlSyntax(value)
}

const form = reactive({
  title: '',
  sql_content: '',
  scheduled_time: null,
  target_db_id: null,
  execution_user: ''
})
const loading = ref(false)
const dbConfigs = ref([])

const isEdit = computed(() => !!route.params.id)

const loadDBConfigs = async () => {
  try {
    const { data } = await getDBConfig()
    dbConfigs.value = data
  } catch (error) {
    console.error('获取数据库配置失败', error)
  }
}

const loadWorkOrder = async () => {
  try {
    const { data } = await getWorkOrder(route.params.id)
    form.title = data.title
    form.sql_content = data.sql_content || ''
    form.target_db_id = data.target_db_id || null
    form.scheduled_time = data.scheduled_time ? new Date(data.scheduled_time) : null
    form.execution_user = data.execution_user || ''
  } catch (error) {
    ElMessage.error('获取工单信息失败')
  }
}

const disabledDate = (time) => {
  return time.getTime() < Date.now() - 86400000
}

const handleSubmit = async () => {
  if (!form.title) {
    ElMessage.error('请输入工单标题')
    return
  }
  if (!form.sql_content) {
    ElMessage.error('请输入SQL内容或上传文件')
    return
  }

  const sql = form.sql_content.toUpperCase()
  if (/\bSELECT\s+\*/.test(sql)) {
    ElMessage.error('禁止使用 SELECT *，请指定具体列名')
    return
  }
  if (/password\s*=\s*['"]/.test(sql)) {
    ElMessage.error('禁止使用明文密码')
    return
  }
  
  loading.value = true
  try {
    const data = {
      title: form.title,
      sql_content: form.sql_content,
      scheduled_time: form.scheduled_time ? formatDate(form.scheduled_time) : null,
      target_db_id: form.target_db_id || null,
      execution_user: form.execution_user || ''
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
  loadDBConfigs()
  if (isEdit.value) {
    loadWorkOrder()
  }
})
</script>

<style scoped>
.workorder-create {
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
  max-width: 900px;
  margin: 0 auto;
  border-radius: 8px;
}

.main-form {
  padding: 20px 0;
}

.sql-editor-container {
  border-radius: 4px;
  overflow: hidden;
  width: 100%;
}

.sql-editor-container :deep(.cm-editor) {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
  min-height: 380px;
  width: 100% !important;
  box-sizing: border-box;
  background-color: #ffffff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.sql-editor-container :deep(.cm-focused) {
  outline: none;
  border-color: #c0c4cc;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.sql-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
  display: flex;
  align-items: center;
}

.error-tag {
  margin-right: 10px;
}

.uploaded-file {
  margin-left: 10px;
  color: #67c23a;
  font-size: 14px;
}

.hint {
  font-size: 12px;
  color: #909399;
  margin-top: 6px;
}

.form-buttons {
  margin-top: 20px;
  padding-left: 120px;
}
</style>
