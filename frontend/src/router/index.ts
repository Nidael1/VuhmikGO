import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/app/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/presentation/views/LoginView.vue'),
    },
    {
      path: '/',
      redirect: '/evidence',
    },
    {
      path: '/evidence',
      name: 'evidence-list',
      component: () => import('@/presentation/views/EvidenceListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/evidence/new',
      name: 'evidence-new',
      component: () => import('@/presentation/views/EvidenceDraftView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/evidence/:id',
      name: 'evidence-detail',
      component: () => import('@/presentation/views/EvidenceDetailView.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login' }
  }
})

export default router
