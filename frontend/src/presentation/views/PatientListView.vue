<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import type { Patient } from '@/domain/types/patient'

const patients = ref<Patient[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try { patients.value = await patientRepository.list() }
  catch (e: any) { error.value = e.message }
  finally { loading.value = false }
})

const sexoLabel: Record<string, string> = { M: 'Masculino', F: 'Femenino', I: 'Indeterminado' }

function calcEdad(fechaNac: string): number {
  const hoy = new Date()
  const nac = new Date(fechaNac)
  let edad = hoy.getFullYear() - nac.getFullYear()
  if (hoy.getMonth() < nac.getMonth() ||
    (hoy.getMonth() === nac.getMonth() && hoy.getDate() < nac.getDate())) edad--
  return edad
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Pacientes</h2>
          <p class="page-sub">Registro de pacientes del consultorio</p>
        </div>
        <RouterLink to="/patients/new" class="btn-primary">+ Nuevo paciente</RouterLink>
      </div>

      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>
      <div v-else-if="patients.length === 0" class="state-empty">
        <p>Sin pacientes registrados.</p>
        <RouterLink to="/patients/new" class="btn-primary">Registrar primer paciente</RouterLink>
      </div>

      <div v-else class="patient-list">
        <RouterLink
          v-for="p in patients"
          :key="p.id"
          :to="`/patients/${p.id}`"
          class="patient-card"
        >
          <div class="card-main">
            <span class="patient-name">{{ p.nombre }}</span>
            <span class="patient-exp">{{ p.num_expediente }}</span>
          </div>
          <div class="card-meta">
            <span>{{ calcEdad(p.fecha_nacimiento) }} años</span>
            <span>·</span>
            <span>{{ sexoLabel[p.sexo] }}</span>
          </div>
        </RouterLink>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 800px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-6); }
.page-sub { color: var(--text-secondary); font-size: 14px; margin-top: var(--space-1); }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; text-decoration: none; white-space: nowrap; }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); display: flex; flex-direction: column; align-items: center; gap: var(--space-4); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-md); padding: var(--space-4); color: var(--color-error); font-size: 14px; }
.patient-list { display: flex; flex-direction: column; gap: var(--space-3); }
.patient-card { display: block; background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-4) var(--space-6); text-decoration: none; transition: border-color 0.15s; }
.patient-card:hover { border-color: var(--color-turquoise); }
.card-main { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-1); }
.patient-name { font-family: var(--font-brand); font-weight: 600; font-size: 15px; color: var(--text-primary); }
.patient-exp { font-size: 12px; color: var(--text-secondary); font-family: monospace; }
.card-meta { font-size: 13px; color: var(--text-secondary); display: flex; gap: var(--space-2); }
</style>
