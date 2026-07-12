<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { http } from '@/infrastructure/api/httpClient'

const loading = ref(true)
const error = ref('')

const profile = ref({
  nombre_completo: '',
  cedula_profesional: '',
  especialidad: '',
})

onMounted(async () => {
  try {
    const res = await http.get<any>('/profile')
    if (res.data) {
      profile.value.nombre_completo = res.data.NombreCompleto ?? ''
      profile.value.cedula_profesional = res.data.CedulaProfesional ?? ''
      profile.value.especialidad = res.data.Especialidad ?? ''
    }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <h2>Mi perfil</h2>
      </div>

      <div v-if="loading" class="state-empty">Cargando...</div>

      <div v-else-if="error" class="alert-error">{{ error }}</div>

      <div v-else class="profile-card">
        <div class="profile-field">
          <span class="field-label">Nombre completo</span>
          <span class="field-value">{{ profile.nombre_completo || '—' }}</span>
        </div>
        <div class="profile-field">
          <span class="field-label">Cédula profesional</span>
          <span class="field-value">{{ profile.cedula_profesional || '—' }}</span>
        </div>
        <div class="profile-field">
          <span class="field-label">Especialidad</span>
          <span class="field-value">{{ profile.especialidad || '—' }}</span>
        </div>
        <p class="profile-note">
          Para actualizar tus datos profesionales contacta al administrador.
        </p>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { width: 100%; max-width: 100%; }
.page-header { margin-bottom: var(--space-6); }
.profile-card {
  background: var(--app-surface);
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
.profile-field { display: flex; flex-direction: column; gap: 4px; }
.field-label { font-size: 12px; font-weight: 600; color: var(--text-secondary); }
.field-value { font-size: 15px; color: var(--text-primary); }
.profile-note {
  font-size: 12px;
  color: var(--text-secondary);
  border-top: 1px solid #E2E8F0;
  padding-top: var(--space-3);
  margin: 0;
}
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
.state-empty { color: var(--text-secondary); padding: var(--space-8); }
</style>
