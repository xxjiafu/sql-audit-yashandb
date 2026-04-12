import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import WorkOrderList from '../views/WorkOrderList.vue'
import WorkOrderCreate from '../views/WorkOrderCreate.vue'
import WorkOrderDetail from '../views/WorkOrderDetail.vue'

const routes = [
  { path: '/', redirect: '/login' },
  { path: '/login', component: Login },
  { path: '/register', component: Register },
  { path: '/workorders', component: WorkOrderList },
  { path: '/workorders/create', component: WorkOrderCreate },
  { path: '/workorders/:id', component: WorkOrderDetail },
  { path: '/workorders/:id/edit', component: WorkOrderCreate }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
