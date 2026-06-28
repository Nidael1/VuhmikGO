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
      redirect: '/patients',
    },
    {
      path: '/patients',
      name: 'patient-list',
      component: () => import('@/presentation/views/PatientListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/patients/new',
      name: 'patient-new',
      component: () => import('@/presentation/views/PatientNewView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/evidence',
      name: 'evidence-list',
      component: () => import('@/presentation/views/EvidenceListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/patients/:id',
      name: 'patient-detail',
      component: () => import('@/presentation/views/PatientDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/evidence/new',
      name: 'evidence-new',
      component: () => import('@/presentation/views/EvidenceDraftView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/evidence/:id/editar',
      name: 'evidence-edit',
      component: () => import('@/presentation/views/EvidenceEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/prescriptions',
      name: 'prescription-list',
      component: () => import('@/presentation/views/PrescriptionListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/presentation/views/ProfileView.vue'),
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
