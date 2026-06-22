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
const allNotes = ref<Evidence[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const [p, notes] = await Promise.all([
      patientRepository.get(id),
      evidenceRepository.list(),
    ])
    patient.value = p
    allNotes.value = notes
  } catch (e: any) { error.value = e.message }
  finally { loading.value = false }
})

// Solo notas activas de este paciente — las anuladas son invisibles (ADR-0006)
const activeNotes = computed(() =>
  allNotes.value.filter(n =>
    n.subject_id === id &&
    n.state !== 'voided' &&
    n.state !== 'draft'
  ).sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
)

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
    year: 'numeric', month: 'long', day: 'numeric',
    hour: '2-digit', minute: '2-digit'
  })
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

        <div class="card patient-info">
          <div class="detail-row">
            <span class="detail-label">Edad</span>
            <span class="detail-value">{{ calcEdad(patient.fecha_nacimiento) }} años</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Nacimiento</span>
            <span class="detail-value">{{ new Date(patient.fecha_nacimiento).toLocaleDateString('es-MX', {year:'numeric',month:'long',day:'numeric'}) }}</span>
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

        <div class="section-header">
          <h3>Notas clínicas</h3>
          <RouterLink :to="`/evidence/new?patient=${id}`" class="btn-primary">
            + Nueva nota
          </RouterLink>
        </div>

        <div v-if="activeNotes.length === 0" class="state-empty-sm">
          Sin notas clínicas registradas para este paciente.
        </div>

        <div v-else class="notes-list">
          <div v-for="note in activeNotes" :key="note.id" class="note-card">
            <div class="note-header">
              <span class="note-date">{{ formatDate(note.created_at) }}</span>
              <RouterLink :to="`/evidence/${note.id}/editar`" class="btn-edit-note">
                Editar
              </RouterLink>
            </div>
            <p class="note-preview" v-if="note.notes">
              {{ note.notes.slice(0, 200) }}{{ note.notes.length > 200 ? '...' : '' }}
            </p>
            <RouterLink :to="`/evidence/${note.id}`" class="note-detail-link">
              Ver detalle →
            </RouterLink>
          </div>
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
.patient-info {}
.detail-row { display: flex; align-items: center; gap: var(--space-4); }
.detail-label { width: 90px; font-size: 13px; color: var(--text-secondary); flex-shrink: 0; }
.detail-value { font-size: 14px; color: var(--text-primary); }
.mono { font-family: monospace; }
.section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-4); }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-2) var(--space-4); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; text-decoration: none; }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.state-empty-sm { color: var(--text-secondary); font-size: 14px; padding: var(--space-4) 0; }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
.notes-list { display: flex; flex-direction: column; gap: var(--space-4); }
.note-card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-5) var(--space-6); display: flex; flex-direction: column; gap: var(--space-3); }
.note-header { display: flex; align-items: center; justify-content: space-between; }
.note-date { font-size: 13px; font-weight: 600; color: var(--text-secondary); }
.btn-edit-note { font-size: 13px; color: var(--color-clinical-blue); text-decoration: none; border: 1px solid #E2E8F0; padding: 2px 12px; border-radius: var(--radius-sm); transition: border-color 0.15s; }
.btn-edit-note:hover { border-color: var(--color-clinical-blue); }
.note-preview { font-size: 15px; color: var(--text-primary); line-height: 1.6; white-space: pre-wrap; }
.note-detail-link { font-size: 13px; color: var(--text-secondary); text-decoration: none; }
.note-detail-link:hover { color: var(--color-clinical-blue); }
</style>
