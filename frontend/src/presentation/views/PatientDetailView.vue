<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import { evidenceRepository } from '@/infrastructure/repositories/evidenceRepository'
import type { Patient } from '@/domain/types/patient'
import type { Evidence } from '@/domain/types/evidence'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const patient = ref<Patient | null>(null)
const notes = ref<Evidence[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const [p, allNotes] = await Promise.all([
      patientRepository.get(id),
      evidenceRepository.list(),
    ])
    patient.value = p
    // Filtrar notas que pertenecen a este paciente por subject_id
    notes.value = allNotes.filter(n => n.subject_id === id)
  } catch (e: any) { error.value = e.message }
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

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('es-MX', {
    year: 'numeric', month: 'short', day: 'numeric'
  })
}

const stateLabel: Record<string, string> = {
  draft: 'Borrador', issued: 'Emitida', locked: 'Bloqueada', voided: 'Anulada'
}
const stateClass: Record<string, string> = {
  draft: 'state-draft', issued: 'state-issued', locked: 'state-locked', voided: 'state-voided'
}

// Nueva nota vinculada a este paciente
function nuevaNota() {
  router.push(`/evidence/new?patient=${id}`)
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>

      <template v-else-if="patient">
        <div class="page-header">
          <div>
            <h2>{{ patient.nombre }}</h2>
            <p class="page-sub">Expediente {{ patient.num_expediente }}</p>
          </div>
          <RouterLink to="/patients" class="btn-back">← Pacientes</RouterLink>
        </div>

        <!-- Datos del paciente -->
        <div class="card">
          <div class="detail-row">
            <span class="detail-label">Edad</span>
            <span class="detail-value">{{ calcEdad(patient.fecha_nacimiento) }} años</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Nacimiento</span>
            <span class="detail-value">{{ formatDate(patient.fecha_nacimiento) }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Sexo</span>
            <span class="detail-value">{{ sexoLabel[patient.sexo] }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Expediente</span>
            <span class="detail-value mono">{{ patient.num_expediente }}</span>
          </div>
        </div>

        <!-- Notas clínicas del paciente -->
        <div class="section-header">
          <h3>Notas clínicas</h3>
          <button class="btn-primary" @click="nuevaNota">+ Nueva nota</button>
        </div>

        <div v-if="notes.length === 0" class="state-empty-sm">
          Sin notas clínicas registradas para este paciente.
        </div>

        <div v-else class="notes-list">
          <RouterLink
            v-for="note in notes"
            :key="note.id"
            :to="`/evidence/${note.id}`"
            class="note-card"
          >
            <div class="note-main">
              <span class="note-date">{{ formatDate(note.created_at) }}</span>
              <span :class="['state-badge', stateClass[note.state]]">
                {{ stateLabel[note.state] }}
              </span>
            </div>
            <p class="note-preview" v-if="note.notes">
              {{ note.notes.slice(0, 120) }}{{ note.notes.length > 120 ? '...' : '' }}
            </p>
          </RouterLink>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 720px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-6); }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: var(--space-1); }
.btn-back { color: var(--color-clinical-blue); font-size: 14px; text-decoration: none; }
.card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-3); margin-bottom: var(--space-6); }
.detail-row { display: flex; align-items: center; gap: var(--space-4); }
.detail-label { width: 90px; font-size: 13px; color: var(--text-secondary); flex-shrink: 0; }
.detail-value { font-size: 14px; color: var(--text-primary); }
.mono { font-family: monospace; }
.section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-4); }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-2) var(--space-4); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; text-decoration: none; }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.state-empty-sm { color: var(--text-secondary); font-size: 14px; padding: var(--space-4) 0; }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
.notes-list { display: flex; flex-direction: column; gap: var(--space-3); }
.note-card { display: block; background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-4) var(--space-6); text-decoration: none; transition: border-color 0.15s; }
.note-card:hover { border-color: var(--color-turquoise); }
.note-main { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-2); }
.note-date { font-size: 14px; font-weight: 500; color: var(--text-primary); }
.note-preview { font-size: 13px; color: var(--text-secondary); line-height: 1.5; }
.state-badge { font-size: 12px; font-weight: 600; padding: 2px 10px; border-radius: 99px; }
.state-draft { background: #F1F5F9; color: var(--text-secondary); }
.state-issued, .state-locked { background: #E6FAF5; color: var(--color-jade); }
.state-voided { background: #FFF0F3; color: var(--color-error); }
</style>
