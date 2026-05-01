import { createRouter, createWebHistory } from 'vue-router'

import ConnectView from '../views/ConnectView.vue'
import DashboardView from '../views/DashboardView.vue'
import DocsView from '../views/DocsView.vue'
import { hasMongoSessionConnection } from '../session/mongoConnection'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/connect',
      name: 'connect',
      component: ConnectView,
    },
    {
      path: '/',
      name: 'dashboard',
      component: DashboardView,
      meta: { requiresMongoConnection: true },
    },
    {
      path: '/docs',
      name: 'docs',
      component: DocsView,
    },
  ],
})

router.beforeEach((to) => {
  if (to.meta.requiresMongoConnection && !hasMongoSessionConnection()) {
    return { name: 'connect', query: { redirect: to.fullPath } }
  }

  if (to.name === 'connect' && hasMongoSessionConnection() && typeof to.query.redirect !== 'string') {
    return { name: 'dashboard' }
  }

  return true
})

export default router
