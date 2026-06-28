<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { prescriptionRepository } from '@/infrastructure/repositories/prescriptionRepository'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import type { Prescription } from '@/domain/types/prescription'
import type { Patient } from '@/domain/types/patient'

const router = useRouter()
const prescriptions = ref<Prescription[]>([])
const patients = ref<Record<string, Patient>>({})
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const [rxs, pats] = await Promise.all([
      prescriptionRepository.listAll(),
      patientRepository.list(),
    ])
    prescriptions.value = rxs
    patients.value = Object.fromEntries(pats.map(p => [p.id, p]))
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('es-MX', {
    year: 'numeric', month: 'long', day: 'numeric',
    hour: '2-digit', minute: '2-digit'
  })
}

function goToPatient(patientId: string) {
  router.push(`/patients/${patientId}`)
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <h2>Recetas</h2>
      </div>

      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>
      <div v-else-if="prescriptions.length === 0" class="state-empty">
        No hay recetas emitidas. Crea una desde el perfil de un paciente.
      </div>

      <div v-else class="rx-list">
        <div
          v-for="rx in prescriptions"
          :key="rx.id"
          class="rx-item"
          @click="goToPatient(rx.patient_id)"
        >
          <div class="rx-header">
            <span class="rx-paciente">
              {{ patients[rx.patient_id]?.nombre ?? rx.patient_id }}
            </span>
            <span class="rx-fecha">{{ formatDate(rx.issued_at ?? rx.created_at) }}</span>
          </div>
          <div class="rx-medicamento">{{ rx.medicamento_generico }}</div>
          <div class="rx-dosis">{{ rx.dosis }}</div>
          <div v-if="rx.diagnostico" class="rx-diagnostico">Dx: {{ rx.diagnostico }}</div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 780px; }
.page-header { margin-bottom: var(--space-6); }
.rx-list { display: flex; flex-direction: column; gap: var(--space-3); }
.rx-item {
  background: var(--app-surface);
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-lg);
  padding: var(--space-4) var(--space-6);
  cursor: pointer;
  transition: border-color 0.15s;
}
.rx-item:hover { border-color: var(--color-turquoise); }
.rx-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--space-2); }
.rx-paciente { font-weight: 600; font-size: 14px; color: var(--text-primary); }
.rx-fecha { font-size: 12px; color: var(--text-secondary); }
.rx-medicamento { font-size: 15px; color: var(--text-primary); font-weight: 600; }
.rx-dosis { font-size: 13px; color: var(--text-secondary); }
.rx-diagnostico { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
</style>
