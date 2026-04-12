import axios from 'axios'

const API_BASE = '/api'

const api = axios.create({
  baseURL: API_BASE,
  timeout: 10000
})

api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const login = (data) => api.post('/auth/login', data)
export const register = (data) => api.post('/auth/register', data)
export const getCurrentUser = () => api.get('/auth/me')

export const getWorkOrderList = (page, pageSize = 10) =>
  api.get('/workorders', { params: { page, page_size: pageSize } })

export const getWorkOrder = (id) => api.get(`/workorders/${id}`)

export const createWorkOrder = (data) => api.post('/workorders', data)

export const updateWorkOrder = (id, data) => api.put(`/workorders/${id}`, data)

export const deleteWorkOrder = (id) => api.delete(`/workorders/${id}`)

export const getAdminWorkOrderList = (params) => api.get('/admin/workorders', { params })

export const getAdminWorkOrder = (id) => api.get(`/workorders/${id}`)

export const leaderApprove = (id) => api.put(`/admin/workorders/${id}/leader-approve`)

export const dbaApprove = (id) => api.put(`/admin/workorders/${id}/dba-approve`)

export const rejectWorkOrder = (id, data) => api.put(`/admin/workorders/${id}/reject`, data)

export const executeWorkOrder = (id) => api.put(`/admin/workorders/${id}/execute`)

export const scheduleWorkOrder = (id, data) => api.put(`/admin/workorders/${id}/schedule`, data)

export const getUserList = () => api.get('/admin/users')

export const updateUserRole = (id, role) => api.put(`/admin/users/${id}/role`, { role })

export const handleUserApply = (id, approved, role) => api.put(`/admin/users/${id}/apply`, { approved, role })