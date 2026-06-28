<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import { evidenceRepository } from '@/infrastructure/repositories/evidenceRepository'
import type { Patient } from '@/domain/types/patient'
import type { Evidence } from '@/domain/types/evidence'
import type { Allergy } from '@/domain/types/allergy'
import { allergyRepository } from '@/infrastructure/repositories/allergyRepository'
import { prescriptionRepository } from '@/infrastructure/repositories/prescriptionRepository'
import type { Prescription } from '@/domain/types/prescription'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const patient = ref<Patient | null>(null)
const allNotes = ref<Evidence[]>([])
const loading = ref(true)
const error = ref('')

// Edición inline del nombre
const editingName = ref(false)
const nameValue = ref('')

// Alergias
const allergies = ref<Allergy[]>([])

// Recetas
const prescriptions = ref<Prescription[]>([])
const showRxForm = ref(false)
const rxForm = ref({ medicamento_generico: '', dosis: '', diagnostico: '', indicaciones: '', seguimiento: '' })
const rxLoading = ref(false)
const rxError = ref('')
const showAllergyForm = ref(false)
const allergyForm = ref({ agente: '', tipo_reaccion: '', criticidad: '', certeza: '' })
const allergyLoading = ref(false)
const allergyError = ref('')

// Edición inline de alergia
const editingAllergyId = ref<string | null>(null)
const editAllergyForm = ref({ agente: '', tipo_reaccion: '', criticidad: '', certeza: '' })

onMounted(async () => {
  try {
    const [p, notes, algs, rxs] = await Promise.all([
      patientRepository.get(id),
      evidenceRepository.list(),
      allergyRepository.list(id),
      prescriptionRepository.listByPatient(id),
    ])
    patient.value = p
    nameValue.value = p.nombre
    allNotes.value = notes
    allergies.value = algs
    prescriptions.value = rxs
  } catch (e: any) { error.value = e.message }
  finally { loading.value = false }
})

async function saveName() {
  const trimmed = nameValue.value.trim()
  if (!trimmed || trimmed === patient.value?.nombre) {
    nameValue.value = patient.value?.nombre ?? ''
    editingName.value = false
    return
  }
  try {
    const updated = await patientRepository.update(patient.value!.id, {
      nombre: trimmed.toUpperCase(),
      fecha_nacimiento: patient.value!.fecha_nacimiento,
      sexo: patient.value!.sexo,
      curp: patient.value!.curp,
    })
    if (patient.value) patient.value.nombre = updated.nombre
    nameValue.value = updated.nombre
  } catch (e: any) {
    nameValue.value = patient.value?.nombre ?? ''
    error.value = e.message
  } finally {
    editingName.value = false
  }
}

const activeNotes = computed(() =>
  allNotes.value.filter(n => {
    if (n.subject_ref !== id) return false
    if (n.state === 'voided' || n.state === 'draft') return false
    try {
      const blob = JSON.parse(n.content)
      return blob.type === 'note'
    } catch { return false }
  }).sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
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

function parseNoteContent(raw: string): string {
  try {
    const obj = JSON.parse(raw)
    return obj.text || raw
  } catch { return raw }
}

async function createAllergy() {
  if (!allergyForm.value.agente.trim() || !allergyForm.value.tipo_reaccion.trim()) {
    allergyError.value = 'Agente y tipo de reacción son obligatorios'
    return
  }
  allergyLoading.value = true
  allergyError.value = ''
  try {
    const a = await allergyRepository.create(id, allergyForm.value)
    allergies.value.push(a)
    showAllergyForm.value = false
    allergyForm.value = { agente: '', tipo_reaccion: '', criticidad: '', certeza: '' }
  } catch (e: any) {
    allergyError.value = e.message
  } finally {
    allergyLoading.value = false
  }
}

function startEditAllergy(a: Allergy) {
  editingAllergyId.value = a.id
  editAllergyForm.value = {
    agente: a.agente,
    tipo_reaccion: a.tipo_reaccion,
    criticidad: a.criticidad ?? '',
    certeza: a.certeza ?? '',
  }
}

function cancelEditAllergy() {
  editingAllergyId.value = null
}

async function saveEditAllergy(a: Allergy) {
  const form = editAllergyForm.value
  if (!form.agente.trim() || !form.tipo_reaccion.trim()) return
  try {
    await allergyRepository.void(a.id)
    const nueva = await allergyRepository.create(id, {
      agente: form.agente.trim(),
      tipo_reaccion: form.tipo_reaccion.trim(),
      criticidad: form.criticidad,
      certeza: form.certeza,
      notas: a.notas,
    })
    const idx = allergies.value.findIndex(x => x.id === a.id)
    if (idx !== -1) allergies.value.splice(idx, 1, nueva)
    editingAllergyId.value = null
  } catch (e: any) {
    error.value = e.message
  }
}

async function quitarAllergy(a: Allergy) {
  try {
    await allergyRepository.void(a.id)
    allergies.value = allergies.value.filter(x => x.id !== a.id)
  } catch (e: any) {
    error.value = e.message
  }
}

async function createPrescription() {
  if (!rxForm.value.medicamento_generico.trim() || !rxForm.value.dosis.trim()) {
    rxError.value = 'Medicamento y dosis son obligatorios'
    return
  }
  rxLoading.value = true
  rxError.value = ''
  try {
    const draft = await prescriptionRepository.create(id, rxForm.value)
    const emitted = await prescriptionRepository.emit(draft.id)
    if (emitted) {
      const updated = await prescriptionRepository.listByPatient(id)
      prescriptions.value = updated
      showRxForm.value = false
      rxForm.value = { medicamento_generico: '', dosis: '', diagnostico: '', indicaciones: '', seguimiento: '' }
    }
  } catch (e: any) {
    rxError.value = e.message
  } finally {
    rxLoading.value = false
  }
}

async function exportExpediente() {
  try {
    const res = await fetch(`/api/v1/patients/${id}/export`, {
      headers: {
        'Authorization': `Bearer ${(await import('@/app/stores/auth')).useAuthStore().token}`,
      }
    })
    if (!res.ok) throw new Error('Error al exportar expediente')
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `expediente_${patient.value?.num_expediente ?? id}.json`
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
            <div style="display:flex; align-items:center; gap:6px;">
              <input
                v-if="editingName"
                v-model="nameValue"
                autofocus
                @blur="saveName"
                @keydown.enter.prevent="saveName"
                @keydown.esc="() => { nameValue = patient!.nombre; editingName = false }"
                style="font-size:1.25rem; font-weight:700; border:none; border-bottom:2px solid #00DFA2; outline:none; background:transparent; min-width:8ch; max-width:320px; text-transform:uppercase;"
              />
              <h2 v-else style="margin:0;">{{ nameValue }}</h2>
              <button
                @click="editingName = true"
                title="Editar nombre"
                style="background:none; border:none; cursor:pointer; padding:2px; color:#9ca3af; display:flex; align-items:center; flex-shrink:0;"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14"
                  viewBox="0 0 24 24" fill="none" stroke="currentColor"
                  stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                  <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                </svg>
              </button>
            </div>
            <div style="display:flex; align-items:center; gap:8px;">
              <p class="page-sub" style="margin:0;">Expediente {{ patient.num_expediente }}</p>
              <button class="btn-accion" @click="exportExpediente">Descargar</button>
            </div>
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

        <!-- Barra de seguridad: alergias activas -->
        <div v-if="allergies.length > 0" class="safety-bar">
          <span class="safety-label">⚠ Alergias:</span>
          <span v-for="a in allergies" :key="a.id" class="allergy-chip">
            {{ a.agente }}
          </span>
        </div>

        <!-- Sección de alergias -->
        <div class="seccion">
          <div class="seccion-header">
            <h3>Alergias e intolerancias</h3>
            <button class="btn-primary" @click="showAllergyForm = !showAllergyForm">
              {{ showAllergyForm ? 'Cancelar' : '+ Nueva alergia' }}
            </button>
          </div>

          <!-- Formulario nueva alergia -->
          <div v-if="showAllergyForm" class="allergy-form">
            <div class="alert-error" v-if="allergyError">{{ allergyError }}</div>
            <div class="form-row">
              <label>Agente *</label>
              <input v-model="allergyForm.agente" placeholder="p. ej. penicilina" class="input" />
            </div>
            <div class="form-row">
              <label>Tipo de reacción *</label>
              <input v-model="allergyForm.tipo_reaccion" placeholder="p. ej. rash, anafilaxia" class="input" />
            </div>
            <div class="form-row">
              <label>Criticidad</label>
              <select v-model="allergyForm.criticidad" class="input">
                <option value="">— opcional —</option>
                <option value="leve">Leve</option>
                <option value="moderada">Moderada</option>
                <option value="grave">Grave</option>
              </select>
            </div>
            <div class="form-row">
              <label>Certeza</label>
              <select v-model="allergyForm.certeza" class="input">
                <option value="">— opcional —</option>
                <option value="confirmada">Confirmada</option>
                <option value="sospecha">Sospecha</option>
                <option value="descartada">Descartada</option>
              </select>
            </div>
            <button class="btn-primary" @click="createAllergy" :disabled="allergyLoading">
              {{ allergyLoading ? 'Guardando...' : 'Registrar alergia' }}
            </button>
          </div>

          <!-- Lista de alergias -->
          <div v-if="allergies.length === 0 && !showAllergyForm" class="state-empty-sm">
            Sin alergias registradas.
          </div>
          <div v-else class="allergy-list">
            <div v-for="a in allergies" :key="a.id" class="allergy-item">
              <!-- Modo edición inline -->
              <template v-if="editingAllergyId === a.id">
                <div class="allergy-edit-form">
                  <input v-model="editAllergyForm.agente" class="input" placeholder="Agente" autofocus />
                  <input v-model="editAllergyForm.tipo_reaccion" class="input" placeholder="Tipo de reacción" />
                  <select v-model="editAllergyForm.criticidad" class="input">
                    <option value="">— criticidad —</option>
                    <option value="leve">Leve</option>
                    <option value="moderada">Moderada</option>
                    <option value="grave">Grave</option>
                  </select>
                  <select v-model="editAllergyForm.certeza" class="input">
                    <option value="">— certeza —</option>
                    <option value="confirmada">Confirmada</option>
                    <option value="sospecha">Sospecha</option>
                    <option value="descartada">Descartada</option>
                  </select>
                  <div class="allergy-edit-acciones">
                    <button class="btn-primary" @click="saveEditAllergy(a)">Guardar</button>
                    <button class="btn-accion" @click="cancelEditAllergy">Cancelar</button>
                  </div>
                </div>
              </template>
              <!-- Modo lectura -->
              <template v-else>
                <div class="allergy-meta">
                  <div class="allergy-main">
                    <span class="allergy-agente">{{ a.agente }}</span>
                    <span v-if="a.criticidad" class="allergy-badge" :class="a.criticidad">
                      {{ a.criticidad }}
                    </span>
                  </div>
                  <div class="allergy-acciones">
                    <button class="btn-accion" @click="startEditAllergy(a)">Editar</button>
                    <button class="btn-accion" @click="quitarAllergy(a)">Quitar</button>
                  </div>
                </div>
                <div class="allergy-sub">{{ a.tipo_reaccion }}</div>
                <div v-if="a.certeza" class="allergy-certeza">Certeza: {{ a.certeza }}</div>
              </template>
            </div>
          </div>
        </div>

        <!-- Sección de recetas electrónicas -->
        <div class="seccion">
          <div class="seccion-header">
            <h3>Recetas electrónicas</h3>
            <button class="btn-primary" @click="showRxForm = !showRxForm">
              {{ showRxForm ? 'Cancelar' : '+ Nueva receta' }}
            </button>
          </div>

          <!-- Formulario nueva receta -->
          <div v-if="showRxForm" class="allergy-form">
            <div class="alert-error" v-if="rxError">{{ rxError }}</div>
            <div class="form-row">
              <label>Medicamento genérico *</label>
              <input v-model="rxForm.medicamento_generico" class="input" placeholder="p. ej. Paracetamol" />
            </div>
            <div class="form-row">
              <label>Dosis *</label>
              <input v-model="rxForm.dosis" class="input" placeholder="p. ej. 500mg cada 8h por 3 días" />
            </div>
            <div class="form-row">
              <label>Diagnóstico</label>
              <input v-model="rxForm.diagnostico" class="input" placeholder="p. ej. Faringitis aguda" />
            </div>
            <div class="form-row">
              <label>Indicaciones</label>
              <input v-model="rxForm.indicaciones" class="input" placeholder="p. ej. Reposo e hidratación" />
            </div>
            <div class="form-row">
              <label>Seguimiento</label>
              <input v-model="rxForm.seguimiento" class="input" placeholder="p. ej. Control en 7 días" />
            </div>
            <button class="btn-primary" @click="createPrescription" :disabled="rxLoading">
              {{ rxLoading ? 'Emitiendo...' : 'Emitir receta' }}
            </button>
          </div>

          <!-- Lista de recetas -->
          <div v-if="prescriptions.length === 0 && !showRxForm" class="state-empty-sm">
            Sin recetas emitidas.
          </div>
          <div v-else class="allergy-list">
            <div v-for="rx in prescriptions" :key="rx.id" class="allergy-item">
              <div class="allergy-meta">
                <div class="allergy-main">
                  <span class="allergy-agente">{{ rx.medicamento_generico }}</span>
                </div>
              </div>
              <div class="allergy-sub">{{ rx.dosis }}</div>
              <div v-if="rx.diagnostico" class="allergy-certeza">Dx: {{ rx.diagnostico }}</div>
            </div>
          </div>
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
                </div>
              </div>
              <div class="nota-contenido">
                {{ parseNoteContent(note.content) }}
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

.safety-bar {
  display: flex; align-items: center; gap: var(--space-2);
  background: #FFF7ED; border: 1px solid #FED7AA;
  border-radius: var(--radius-md); padding: var(--space-3) var(--space-4);
  margin-bottom: var(--space-4); font-size: 13px;
}
.safety-label { font-weight: 700; color: #C2410C; }
.allergy-chip {
  background: #FEF3C7; border: 1px solid #FDE68A;
  border-radius: 999px; padding: 2px 10px;
  font-size: 12px; font-weight: 600; color: #92400E;
}
.seccion {
  background: var(--app-surface); border: 1px solid #E2E8F0;
  border-radius: var(--radius-lg); overflow: hidden; margin-bottom: var(--space-4);
}
.seccion-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: var(--space-4) var(--space-6); border-bottom: 1px solid #E2E8F0;
  background: #FAFBFC;
}
.allergy-form {
  padding: var(--space-4) var(--space-6); border-bottom: 1px solid #E2E8F0;
  display: flex; flex-direction: column; gap: var(--space-3);
}
.form-row { display: flex; flex-direction: column; gap: 4px; }
.form-row label { font-size: 12px; font-weight: 600; color: var(--text-secondary); }
.input {
  font-family: var(--font-body); font-size: 14px;
  border: 1.5px solid #E2E8F0; border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3); color: var(--text-primary);
  background: var(--app-surface); outline: none;
}
.input:focus { border-color: var(--color-turquoise); }
.allergy-list { padding: var(--space-2) 0; }
.allergy-item {
  padding: var(--space-3) var(--space-6); border-bottom: 1px solid #F1F5F9;
}
.allergy-item:last-child { border-bottom: none; }
.allergy-meta { display: flex; align-items: flex-start; justify-content: space-between; gap: var(--space-2); }
.allergy-acciones { display: flex; gap: var(--space-2); flex-shrink: 0; }
.allergy-main { display: flex; align-items: center; gap: var(--space-2); margin-bottom: 2px; }
.allergy-agente { font-weight: 600; font-size: 14px; color: var(--text-primary); }
.allergy-badge {
  font-size: 11px; font-weight: 600; border-radius: 999px; padding: 1px 8px;
}
.allergy-badge.leve { background: #DCFCE7; color: #166534; }
.allergy-badge.moderada { background: #FEF9C3; color: #854D0E; }
.allergy-badge.grave { background: #FEE2E2; color: #991B1B; }
.allergy-sub { font-size: 13px; color: var(--text-secondary); }
.allergy-certeza { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }
.allergy-edit-form { display: flex; flex-direction: column; gap: var(--space-2); padding: var(--space-2) 0; }
.allergy-edit-acciones { display: flex; gap: var(--space-2); margin-top: var(--space-1); }
</style>
