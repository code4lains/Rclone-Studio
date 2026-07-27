import { createRouter, createWebHashHistory } from 'vue-router'
import Layout from '../components/Layout.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      component: Layout,
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('../views/Dashboard.vue')
        },
        {
          path: 'instances',
          name: 'instances',
          component: () => import('../views/Instances.vue')
        },
        {
          path: 'config',
          name: 'config',
          component: () => import('../views/Config.vue')
        },
        {
          path: 'files',
          name: 'files',
          component: () => import('../views/Files.vue')
        },
        {
          path: 'mounts',
          name: 'mounts',
          component: () => import('../views/Mounts.vue')
        },
        {
          path: 'help',
          name: 'help',
          component: () => import('../views/Help.vue')
        }
      ]
    }
  ]
})

// Navigation guards removed to allow free navigation and rely on Layout.vue's disconnected overlay

export default router
