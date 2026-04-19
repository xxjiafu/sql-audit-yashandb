import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import DefaultLayout from '../layouts/DefaultLayout.vue'
import WorkOrderList from '../views/WorkOrderList.vue'
import WorkOrderCreate from '../views/WorkOrderCreate.vue'
import WorkOrderDetail from '../views/WorkOrderDetail.vue'
import AdminWorkOrderList from '../views/AdminWorkOrderList.vue'
import AdminWorkOrderDetail from '../views/AdminWorkOrderDetail.vue'
import DBConfig from '../views/DBConfig.vue'
import UserManage from '../views/UserManage.vue'

const routes = [
  { path: '/', redirect: '/workorders' },
  { path: '/login', component: Login },
  { path: '/register', component: Register },
  {
    path: '/',
    component: DefaultLayout,
    children: [
      { path: '/workorders', component: WorkOrderList },
      { path: '/workorders/create', component: WorkOrderCreate },
      { path: '/workorders/:id', component: WorkOrderDetail },
      { path: '/workorders/:id/edit', component: WorkOrderCreate },
      { path: '/db-config', component: DBConfig },
      { path: '/admin/workorders', component: AdminWorkOrderList },
      { path: '/admin/workorders/:id', component: AdminWorkOrderDetail },
      { path: '/admin/db-config', component: DBConfig },
      { path: '/admin/users', component: UserManage }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const publicPaths = ['/login', '/register']
  const isPublic = publicPaths.includes(to.path)
  const token = localStorage.getItem('token')

  if (isPublic) {
    next()
  } else if (!token) {
    next('/login')
  } else {
    next()
  }
})

export default router
