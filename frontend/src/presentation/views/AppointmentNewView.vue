<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { appointmentRepository } from '@/infrastructure/repositories/appointmentRepository'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import type { Patient } from '@/domain/types/patient'
import { onMounted } from 'vue'

const router = useRouter()
const route = useRoute()

const patients = ref<Patient[]>([])
const patientId = ref((route.query.patient as string) ?? '')
const scheduledAt = ref('')
const durationMin = ref(30)
const reason = ref('')
const notes = ref('')
const loading = ref(false)
const error = ref('')

onMounted(async () => {
  patients.value = await patientRepository.list()
})

async function guardar() {
  if (!patientId.value || !scheduledAt.value) {
    error.value = 'Paciente y fecha/hora son obligatorios.'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await appointmentRepository.create({
      patient_id: patientId.value,
      scheduled_at: new Date(scheduledAt.value).toISOString(),
      duration_min: durationMin.value,
      reason: reason.value,
      notes: notes.value,
    })
    router.push('/appointments')
  } catch (e: any) {
    error.value = e.message ?? 'Error al guardar la cita.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <h2>Nueva cita</h2>
        <RouterLink to="/appointments" class="btn-secondary">← Volver</RouterLink>
      </div>

      <div class="form-card">
        <div v-if="error" class="alert-error">{{ error }}</div>

        <div class="field">
          <label>Paciente</label>
          <select v-model="patientId" class="input">
            <option value="" disabled>Selecciona un paciente...</option>
            <option v-for="p in patients" :key="p.id" :value="p.id">{{ p.nombre }}</option>
          </select>
        </div>

        <div class="field">
          <label>Fecha y hora</label>
          <input v-model="scheduledAt" type="datetime-local" class="input" />
        </div>

        <div class="field">
          <label>Duración</label>
          <select v-model="durationMin" class="input">
            <option :value="15">15 minutos</option>
            <option :value="30">30 minutos</option>
            <option :value="45">45 minutos</option>
            <option :value="60">1 hora</option>
          </select>
        </div>

        <div class="field">
          <label>Motivo <span class="opcional">(opcional)</span></label>
          <input v-model="reason" type="text" class="input" placeholder="Ej. Revisión general, seguimiento..." />
        </div>

        <div class="field">
          <label>Notas internas <span class="opcional">(opcional)</span></label>
          <textarea v-model="notes" class="input textarea" rows="3" placeholder="Notas para el médico..." />
        </div>

        <div class="form-actions">
          <button class="btn-primary" :disabled="loading" @click="guardar">
            {{ loading ? 'Guardando...' : 'Agendar cita' }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 640px; margin: 0 auto; }
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-6); }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-3) var(--space-6); border-radius: var(--radius-md); font-size: 15px; font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { font-size: 14px; color: var(--text-secondary); text-decoration: none; }
.btn-secondary:hover { color: var(--text-primary); }
.form-card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-lg); padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-5); }
.field { display: flex; flex-direction: column; gap: var(--space-2); }
label { font-size: 13px; font-weight: 600; color: var(--text-primary); }
.opcional { font-weight: 400; color: var(--text-secondary); }
.input { font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-surface); outline: none; }
.input:focus { border-color: var(--color-turquoise); }
.textarea { resize: vertical; }
.form-actions { display: flex; justify-content: flex-end; padding-top: var(--space-2); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
</style>
