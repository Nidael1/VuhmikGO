<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { consultationRepository } from '@/infrastructure/repositories/consultationRepository'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import { evidenceRepository } from '@/infrastructure/repositories/evidenceRepository'
import { prescriptionRepository } from '@/infrastructure/repositories/prescriptionRepository'
import { useAuthStore } from '@/app/stores/auth'
import type { Consultation } from '@/domain/types/consultation'
import type { Patient } from '@/domain/types/patient'
import type { Evidence } from '@/domain/types/evidence'
import type { Prescription } from '@/domain/types/prescription'

const route = useRoute()
const id = route.params.id as string

const consultation = ref<Consultation | null>(null)
const patient = ref<Patient | null>(null)
const allNotes = ref<Evidence[]>([])
const receta = ref<Prescription | null>(null)
const loading = ref(true)
const error = ref('')
const auth = useAuthStore()

onMounted(async () => {
  try {
    const [con, notes, rxs] = await Promise.all([
      consultationRepository.get(id),
      evidenceRepository.list(),
      prescriptionRepository.listAll(),
    ])
    consultation.value = con
    allNotes.value = notes
    patient.value = await patientRepository.get(con.patient_id)
    receta.value = rxs.find((r: Prescription) => r.consultation_id === id) ?? null
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

const notaClinica = computed(() => {
  const nota = allNotes.value.find(n => {
    try {
      const blob = JSON.parse(n.content)
      return blob.type === 'note' && blob.consultation_id === id
    } catch { return false }
  })
  if (!nota) return null
  try {
    const obj = JSON.parse(nota.content)
    return obj.text || nota.content
  } catch { return nota.content }
})

function reimprimir() {
  if (!receta.value || !auth.token) return
  window.open(`/api/v1/prescriptions/${receta.value.id}/print?token=${auth.token}`, '_blank')
}

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('es-MX', {
    weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
    hour: '2-digit', minute: '2-digit'
  })
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>

      <template v-else-if="consultation">
        <div class="page-header">
          <div>
            <h2>Consulta médica</h2>
            <p class="page-sub">{{ formatDate(consultation.issued_at ?? consultation.created_at) }}</p>
          </div>
          <RouterLink to="/consultations" class="btn-back">← Consultas</RouterLink>
        </div>

        <div class="seccion seccion--consultas">
          <div class="seccion-header">
            <div class="seccion-titulo">
              <svg class="seccion-icono" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M8 2v4"/><path d="M16 2v4"/>
                <rect x="3" y="4" width="18" height="18" rx="2"/>
                <line x1="3" y1="10" x2="21" y2="10"/>
                <path d="M9 16l2 2 4-4"/>
              </svg>
              <h3>Detalle de consulta</h3>
            </div>
          </div>

          <div class="detalle-grid">
            <div class="detalle-item">
              <span class="detalle-label">Paciente</span>
              <RouterLink v-if="patient" :to="`/patients/${patient.id}`" class="detalle-link">
                {{ patient.nombre }}
              </RouterLink>
            </div>

            <div v-if="consultation.ta || consultation.fc || consultation.fr || consultation.temp || consultation.peso || consultation.talla || consultation.sao2" class="detalle-item">
              <span class="detalle-label">Signos vitales</span>
              <div class="vitals-row">
                <span v-if="consultation.ta" class="vital-chip"><strong>T/A</strong> {{ consultation.ta }} mmHg</span>
                <span v-if="consultation.fc" class="vital-chip"><strong>FC</strong> {{ consultation.fc }} lpm</span>
                <span v-if="consultation.fr" class="vital-chip"><strong>FR</strong> {{ consultation.fr }} rpm</span>
                <span v-if="consultation.temp" class="vital-chip"><strong>Temp</strong> {{ consultation.temp }}°C</span>
                <span v-if="consultation.peso" class="vital-chip"><strong>Peso</strong> {{ consultation.peso }} kg</span>
                <span v-if="consultation.talla" class="vital-chip"><strong>Talla</strong> {{ consultation.talla }} m</span>
                <span v-if="consultation.sao2" class="vital-chip"><strong>SAO2</strong> {{ consultation.sao2 }}%</span>
              </div>
            </div>

            <div class="detalle-item">
              <span class="detalle-label">Nota clínica</span>
              <span class="detalle-valor detalle-nota">{{ notaClinica || 'sin nota' }}</span>
            </div>

            <div v-if="receta" class="detalle-item">
              <span class="detalle-label">Receta electrónica</span>
              <div class="receta-row">
                <div class="receta-info">
                  <span class="receta-med">{{ receta.medicamento_generico }}</span>
                  <span class="receta-dosis">{{ receta.dosis }}</span>
                  <span v-if="receta.diagnostico" class="receta-dx">Dx: {{ receta.diagnostico }}</span>
                </div>
                <button class="btn-reimprimir-sm" @click="reimprimir">
                  <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="6 9 6 2 18 2 18 9"/>
                    <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/>
                    <rect x="6" y="14" width="12" height="8"/>
                  </svg>
                  Imprimir
                </button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 780px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-6); }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: 2px; text-transform: capitalize; }
.btn-back { color: var(--color-clinical-blue); font-size: 14px; text-decoration: none; white-space: nowrap; }

.seccion {
  background: var(--app-surface);
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.seccion-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid #E2E8F0;
}

.seccion-titulo { display: flex; align-items: center; gap: var(--space-2); }
.seccion-titulo h3 { margin: 0; font-size: 14px; font-weight: 700; color: var(--text-primary); }
.seccion-icono { display: flex; align-items: center; color: var(--text-secondary); flex-shrink: 0; }

.seccion--consultas .seccion-header {
  background: #F2FDFB;
  border-left: 3px solid var(--color-turquoise, #00DFA2);
}

.detalle-grid {
  display: flex;
  flex-direction: column;
}

.detalle-item {
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid #F1F5F9;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.detalle-item:last-child { border-bottom: none; }

.detalle-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.detalle-valor {
  font-size: 15px;
  color: var(--text-primary);
}

.detalle-nota {
  line-height: 1.7;
  white-space: pre-wrap;
}

.detalle-link {
  font-size: 15px;
  color: var(--color-clinical-blue);
  text-decoration: none;
  font-weight: 600;
  width: fit-content;
}
.detalle-link:hover { text-decoration: underline; }

.vitals-row { display: flex; flex-wrap: wrap; gap: 6px; }
.vital-chip {
  font-size: 12px;
  background: #EEF9F7;
  color: var(--color-turquoise);
  border: 1px solid #C3EDE8;
  border-radius: 20px;
  padding: 2px 10px;
  white-space: nowrap;
}
.vital-chip strong { font-weight: 700; }

.receta-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}
.receta-info { display: flex; flex-direction: column; gap: 2px; }
.receta-med { font-size: 15px; font-weight: 600; color: var(--text-primary); }
.receta-dosis { font-size: 13px; color: var(--text-secondary); }
.receta-dx { font-size: 12px; color: var(--text-secondary); }
.btn-reimprimir-sm {
  display: flex;
  align-items: center;
  gap: 3px;
  font-family: var(--font-body);
  font-size: 11px;
  font-weight: 600;
  color: var(--color-clinical-blue, #3B82F6);
  background: transparent;
  border: 1px solid var(--color-clinical-blue, #3B82F6);
  border-radius: var(--radius-sm);
  padding: 2px 8px;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  flex-shrink: 0;
}
.btn-reimprimir-sm:hover {
  background: var(--color-clinical-blue, #3B82F6);
  color: #fff;
}
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.alert-error {
  background: #FFF0F3; border: 1px solid var(--color-error);
  border-radius: var(--radius-sm); padding: var(--space-3);
  font-size: 14px; color: var(--color-error);
}
</style>
