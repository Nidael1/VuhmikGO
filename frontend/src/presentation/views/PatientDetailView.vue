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
    weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
    hour: '2-digit', minute: '2-digit'
  })
}

async function exportNote(noteId: string) {
  try {
    const blob = await evidenceRepository.export(noteId)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `nota_${noteId}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e: any) { error.value = e.message }
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>

      <template v-else-if="patient">
        <!-- Encabezado del paciente -->
        <div class="page-header">
          <div>
            <h2>{{ patient.nombre }}</h2>
            <p class="page-sub">Expediente {{ patient.num_expediente }}</p>
          </div>
          <RouterLink to="/patients" class="btn-back">← Pacientes</RouterLink>
        </div>

        <!-- Datos del paciente en línea -->
        <div class="patient-meta">
          <span>{{ calcEdad(patient.fecha_nacimiento) }} años</span>
          <span class="sep">·</span>
          <span>{{ sexoLabel[patient.sexo] }}</span>
          <span class="sep">·</span>
          <span class="mono">{{ patient.num_expediente }}</span>
        </div>

        <!-- Expediente clínico — hoja continua -->
        <div class="expediente">
          <div class="expediente-header">
            <h3>Expediente clínico</h3>
            <RouterLink :to="`/evidence/new?patient=${id}`" class="btn-primary">
              + Nueva nota
            </RouterLink>
          </div>

          <div v-if="activeNotes.length === 0" class="state-empty-sm">
            Sin notas clínicas registradas para este paciente.
          </div>

          <!-- Hoja continua de notas -->
          <div v-else class="hoja">
            <div
              v-for="(note, index) in activeNotes"
              :key="note.id"
              class="nota-entrada"
              :class="{ 'primera': index === 0 }"
            >
              <div class="nota-meta">
                <span class="nota-fecha">{{ formatDate(note.created_at) }}</span>
                <div class="nota-acciones">
                  <RouterLink :to="`/evidence/${note.id}/editar`" class="btn-accion">
                    Editar
                  </RouterLink>
                  <button class="btn-accion" @click="exportNote(note.id)">
                    Descargar
                  </button>
                </div>
              </div>
              <div class="nota-contenido">
                {{ note.notes || 'Sin contenido.' }}
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
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-2); }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: 2px; }
.btn-back { color: var(--color-clinical-blue); font-size: 14px; text-decoration: none; white-space: nowrap; }

.patient-meta {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: var(--space-6);
  padding-bottom: var(--space-4);
  border-bottom: 1px solid #E2E8F0;
}
.sep { color: #CBD5E1; }
.mono { font-family: monospace; }

.expediente { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-lg); overflow: hidden; }

.expediente-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid #E2E8F0;
  background: #FAFBFC;
}

.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-2) var(--space-4); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; text-decoration: none; }

.hoja { padding: 0; }

.nota-entrada {
  padding: var(--space-6);
  border-bottom: 1px solid #F1F5F9;
}
.nota-entrada:last-child { border-bottom: none; }

.nota-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-3);
}

.nota-fecha {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: capitalize;
}

.nota-acciones {
  display: flex;
  gap: var(--space-2);
}

.btn-accion {
  font-size: 12px;
  color: var(--color-clinical-blue);
  text-decoration: none;
  background: transparent;
  border: 1px solid #E2E8F0;
  padding: 2px 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: border-color 0.15s;
  font-family: var(--font-body);
}
.btn-accion:hover { border-color: var(--color-clinical-blue); }

.nota-contenido {
  font-size: 15px;
  color: var(--text-primary);
  line-height: 1.7;
  white-space: pre-wrap;
}

.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.state-empty-sm { color: var(--text-secondary); font-size: 14px; padding: var(--space-6); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
</style>
