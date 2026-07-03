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
import { consultationRepository } from '@/infrastructure/repositories/consultationRepository'
import type { Consultation } from '@/domain/types/consultation'
import { useAuthStore } from '@/app/stores/auth'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const patient = ref<Patient | null>(null)
const allNotes = ref<Evidence[]>([])
const loading = ref(true)
const error = ref('')

const editingName = ref(false)
const nameValue = ref('')

const allergies = ref<Allergy[]>([])
const prescriptions = ref<Prescription[]>([])
const consultations = ref<Consultation[]>([])
const showRxForm = ref(false)
const rxForm = ref({ medicamento_generico: '', dosis: '', diagnostico: '', indicaciones: '', seguimiento: '' })
const rxLoading = ref(false)
const rxError = ref('')
const showAllergyForm = ref(false)
const allergyForm = ref({ agente: '', tipo_reaccion: '', criticidad: '', certeza: '' })
const allergyLoading = ref(false)
const allergyError = ref('')

const editingAllergyId = ref<string | null>(null)
const editAllergyForm = ref({ agente: '', tipo_reaccion: '', criticidad: '', certeza: '' })
const auth = useAuthStore()

// Secciones colapsables — todas cerradas por defecto
const seccionesAbiertas = ref<Record<string, boolean>>({
  alergias: false,
  recetas: false,
  consultas: false,
})

// Panel lateral de notas generales — visible por defecto, toggle con doble clic
const notasPanelAbierto = ref(true)

function toggleNotasPanel() {
  notasPanelAbierto.value = !notasPanelAbierto.value
}

function toggleSeccion(nombre: string) {
  seccionesAbiertas.value[nombre] = !seccionesAbiertas.value[nombre]
}

// Notas clínicas del expediente
const showNotaForm = ref(false)
const notaForm = ref('')
const notaLoading = ref(false)
const notaError = ref('')
const editingNotaId = ref<string | null>(null)
const editNotaForm = ref('')

onMounted(async () => {
  try {
    const [p, notes, algs, rxs, cons] = await Promise.all([
      patientRepository.get(id),
      evidenceRepository.list(),
      allergyRepository.list(id),
      prescriptionRepository.listByPatient(id),
      consultationRepository.listByPatient(id),
    ])
    patient.value = p
    nameValue.value = p.nombre
    allNotes.value = notes
    allergies.value = algs
    prescriptions.value = rxs
    consultations.value = cons
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

const notasExpediente = computed(() =>
  allNotes.value.filter(n => {
    if (n.subject_ref !== id) return false
    if (n.state === 'voided' || n.state === 'draft') return false
    try {
      const blob = JSON.parse(n.content)
      return blob.type === 'note' && !blob.consultation_id
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

function reimprimirRx(rxId: string) {
  if (!auth.token) return
  window.open(`/api/v1/prescriptions/${rxId}/print?token=${auth.token}`, '_blank')
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

async function crearNota() {
  if (!notaForm.value.trim()) {
    notaError.value = 'La nota no puede estar vacía'
    return
  }
  notaLoading.value = true
  notaError.value = ''
  try {
    await evidenceRepository.draft({
      subject_ref: id,
      content: JSON.stringify({ type: 'note', text: notaForm.value.trim() }),
    })
    const notes = await evidenceRepository.list()
    allNotes.value = notes
    showNotaForm.value = false
    notaForm.value = ''
  } catch (e: any) {
    notaError.value = e.message
  } finally {
    notaLoading.value = false
  }
}

function startEditNota(n: Evidence) {
  editingNotaId.value = n.id
  try {
    editNotaForm.value = JSON.parse(n.content)?.text ?? n.content
  } catch {
    editNotaForm.value = n.content
  }
}

function cancelEditNota() {
  editingNotaId.value = null
  editNotaForm.value = ''
}

async function saveEditNota(n: Evidence) {
  if (!editNotaForm.value.trim()) return
  notaLoading.value = true
  try {
    await evidenceRepository.void(n.id, { reason_code: 'RC_VOID_CORRECTION' })
    await evidenceRepository.draft({
      subject_ref: id,
      content: JSON.stringify({ type: 'note', text: editNotaForm.value.trim() }),
    })
    const notes = await evidenceRepository.list()
    allNotes.value = notes
    editingNotaId.value = null
    editNotaForm.value = ''
  } catch (e: any) {
    notaError.value = e.message
  } finally {
    notaLoading.value = false
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
          <svg class="safety-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
            <line x1="12" y1="9" x2="12" y2="13"/>
            <line x1="12" y1="17" x2="12.01" y2="17"/>
          </svg>
          <span class="safety-label">Alergias:</span>
          <span v-for="a in allergies" :key="a.id" class="allergy-chip">
            {{ a.agente }}
          </span>
        </div>

        <!-- LAYOUT DOS COLUMNAS: expediente (izq) + notas generales (der) -->
        <div class="expediente-layout" :class="{ 'expediente-layout--collapsed': !notasPanelAbierto }">
          <div class="expediente-main">

        <!-- SECCIÓN: Alergias e intolerancias -->
        <div class="seccion seccion--alergias">
          <div class="seccion-header">
            <div class="seccion-titulo" @click="toggleSeccion('alergias')" style="cursor:pointer;flex:1;">
              <svg class="seccion-icono" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                <line x1="12" y1="9" x2="12" y2="13"/>
                <line x1="12" y1="17" x2="12.01" y2="17"/>
              </svg>
              <h3>Alergias e intolerancias</h3>
              <span class="seccion-count">{{ allergies.length }}</span>
            </div>
            <button class="btn-primary" @click.stop="showAllergyForm = !showAllergyForm">
              {{ showAllergyForm ? 'Cancelar' : '+ Nueva alergia' }}
            </button>
          </div>

          <div v-show="seccionesAbiertas['alergias']">
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

          <div v-if="allergies.length === 0 && !showAllergyForm" class="state-empty-sm">
            Sin alergias registradas.
          </div>
          <div v-else class="allergy-list">
            <div v-for="a in allergies" :key="a.id" class="allergy-item">
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
        </div>

        <!-- SECCIÓN: Recetas electrónicas -->
        <div class="seccion seccion--recetas">
          <div class="seccion-header">
            <div class="seccion-titulo" @click="toggleSeccion('recetas')" style="cursor:pointer;flex:1;">
              <svg class="seccion-icono" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M8 2v4"/><path d="M16 2v4"/>
                <rect x="3" y="6" width="18" height="16" rx="2"/>
                <line x1="9" y1="13" x2="15" y2="13"/>
                <line x1="9" y1="17" x2="15" y2="17"/>
              </svg>
              <h3>Recetas electrónicas</h3>
              <span class="seccion-count">{{ prescriptions.length }}</span>
            </div>
            <button class="btn-primary" @click.stop="showRxForm = !showRxForm">
              {{ showRxForm ? 'Cancelar' : '+ Nueva receta' }}
            </button>
          </div>

          <div v-show="seccionesAbiertas['recetas']">
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

          <div v-if="prescriptions.length === 0 && !showRxForm" class="state-empty-sm">
            Sin recetas emitidas.
          </div>
          <div v-else class="rx-grid">
            <RouterLink v-for="rx in prescriptions" :key="rx.id" :to="`/prescriptions/${rx.id}`" class="rx-card">
              <div class="rx-card-header">
                <span class="rx-medicamento">{{ rx.medicamento_generico }}</span>
                <div class="rx-card-acciones">
                  <span class="rx-estado">emitida</span>
                  <button class="btn-reimprimir-sm" @click.stop="reimprimirRx(rx.id)">
                    <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <polyline points="6 9 6 2 18 2 18 9"/>
                      <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/>
                      <rect x="6" y="14" width="12" height="8"/>
                    </svg>
                    Imprimir
                  </button>
                </div>
              </div>
              <div class="rx-dosis-text">{{ rx.dosis }}</div>
              <div v-if="rx.diagnostico" class="rx-dx">Dx: {{ rx.diagnostico }}</div>
            </RouterLink>
          </div>
          </div>
        </div>

        <!-- SECCIÓN: Consultas — cronología clínica -->
        <div class="seccion seccion--consultas">
          <div class="seccion-header">
            <div class="seccion-titulo" @click="toggleSeccion('consultas')" style="cursor:pointer;flex:1;">
              <svg class="seccion-icono" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M8 2v4"/><path d="M16 2v4"/>
                <rect x="3" y="4" width="18" height="18" rx="2"/>
                <line x1="3" y1="10" x2="21" y2="10"/>
                <path d="M9 16l2 2 4-4"/>
              </svg>
              <h3>Consultas</h3>
              <span class="seccion-count">{{ consultations.length }}</span>
            </div>
            <RouterLink :to="`/consultations/new?patient=${id}`" class="btn-primary" @click.stop>
              + Nueva consulta
            </RouterLink>
          </div>

          <div v-show="seccionesAbiertas['consultas']">
          <div v-if="consultations.length === 0" class="state-empty-sm">
            Sin consultas registradas para este paciente.
          </div>

          <div class="con-lista">
            <RouterLink
              v-for="con in consultations"
              :key="con.id"
              :to="`/consultations/${con.id}`"
              class="con-card"
            >
              <div class="nota-meta">
                <span class="nota-fecha">{{ formatDate(con.issued_at ?? con.created_at) }}</span>
              </div>

              <div v-if="con.ta || con.fc || con.fr || con.temp || con.peso || con.talla || con.sao2" class="vitals-row">
                <span v-if="con.ta" class="vital-chip"><strong>T/A</strong> {{ con.ta }} mmHg</span>
                <span v-if="con.fc" class="vital-chip"><strong>FC</strong> {{ con.fc }} lpm</span>
                <span v-if="con.fr" class="vital-chip"><strong>FR</strong> {{ con.fr }} rpm</span>
                <span v-if="con.temp" class="vital-chip"><strong>Temp</strong> {{ con.temp }}°C</span>
                <span v-if="con.peso" class="vital-chip"><strong>Peso</strong> {{ con.peso }} kg</span>
                <span v-if="con.talla" class="vital-chip"><strong>Talla</strong> {{ con.talla }} m</span>
                <span v-if="con.sao2" class="vital-chip"><strong>SAO2</strong> {{ con.sao2 }}%</span>
              </div>

              <div class="nota-contenido">
                {{ activeNotes.find(n => {
                  try { return JSON.parse(n.content)?.consultation_id === con.id } catch { return false }
                })?.content ? parseNoteContent(activeNotes.find(n => {
                  try { return JSON.parse(n.content)?.consultation_id === con.id } catch { return false }
                })!.content) : 'sin nota' }}
              </div>
            </RouterLink>
          </div>
          </div>
        </div>

          </div><!-- fin .expediente-main -->

          <!-- PANEL LATERAL: Notas generales del paciente (consultation_id = NULL) -->
          <div
            v-if="notasPanelAbierto"
            class="expediente-notas-panel"
            @dblclick="toggleNotasPanel"
            title="Doble clic para contraer"
          >
            <div class="notas-panel-inner">
              <div class="notas-panel-header">
                <div class="seccion-titulo">
                  <svg class="seccion-icono" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                    <polyline points="14 2 14 8 20 8"/>
                    <line x1="16" y1="13" x2="8" y2="13"/>
                    <line x1="16" y1="17" x2="8" y2="17"/>
                    <polyline points="10 9 9 9 8 9"/>
                  </svg>
                  <h3>Notas generales</h3>
                  <span class="seccion-count">{{ notasExpediente.length }}</span>
                </div>
                <button class="btn-primary" @click.stop="showNotaForm = !showNotaForm; notaError = ''">
                  {{ showNotaForm ? 'Cancelar' : '+ Nueva nota' }}
                </button>
              </div>

              <div v-if="showNotaForm" class="allergy-form">
                <div class="alert-error" v-if="notaError">{{ notaError }}</div>
                <div class="form-row">
                  <label>Nota clínica</label>
                  <textarea
                    v-model="notaForm"
                    class="input"
                    rows="4"
                    placeholder="Observación clínica, seguimiento, nota de progreso..."
                    maxlength="2000"
                    style="resize:vertical;"
                  />
                  <span style="font-size:12px;color:var(--text-secondary);text-align:right">{{ notaForm.length }} / 2000</span>
                </div>
                <button class="btn-primary" @click="crearNota" :disabled="notaLoading">
                  {{ notaLoading ? 'Guardando...' : 'Guardar nota' }}
                </button>
              </div>

              <div v-if="notasExpediente.length === 0 && !showNotaForm" class="state-empty-sm">
                Sin notas generales registradas.
              </div>

              <div v-else class="hoja">
                <div v-for="nota in notasExpediente" :key="nota.id" class="nota-entrada">
                  <div class="nota-meta">
                    <span class="nota-fecha">{{ formatDate(nota.issued_at ?? nota.created_at) }}</span>
                    <button v-if="editingNotaId !== nota.id" class="btn-accion" @click.stop="startEditNota(nota)">Editar</button>
                  </div>
                  <template v-if="editingNotaId === nota.id">
                    <textarea
                      v-model="editNotaForm"
                      class="input"
                      rows="4"
                      style="resize:vertical;width:100%;margin-bottom:var(--space-2)"
                      maxlength="2000"
                      @dblclick.stop
                    />
                    <div style="display:flex;gap:var(--space-2)">
                      <button class="btn-primary" @click.stop="saveEditNota(nota)" :disabled="notaLoading">
                        {{ notaLoading ? 'Guardando...' : 'Guardar' }}
                      </button>
                      <button class="btn-accion" @click.stop="cancelEditNota">Cancelar</button>
                    </div>
                  </template>
                  <div v-else class="nota-contenido">{{ parseNoteContent(nota.content) }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Indicador para reabrir panel (cuando está contraído) -->
          <div
            v-else
            class="expediente-notas-collapsed"
            @dblclick="toggleNotasPanel"
            title="Doble clic para expandir notas"
          >
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/>
              <line x1="16" y1="17" x2="8" y2="17"/>
              <polyline points="10 9 9 9 8 9"/>
            </svg>
          </div>

        </div><!-- fin .expediente-layout -->

      </template>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 1200px; }

/* Layout dos columnas: expediente 65% + notas generales 35% (ADR-0026) */
.expediente-layout {
  display: flex;
  gap: var(--space-6);
  align-items: flex-start;
}

.expediente-main {
  flex: 0 0 65%;
  min-width: 0;
}

.expediente-layout--collapsed .expediente-main {
  flex: 1 1 100%;
}

.expediente-notas-panel {
  flex: 0 0 33%;
  position: sticky;
  top: var(--space-4);
  max-height: calc(100vh - var(--space-8));
  overflow-y: auto;
  background: var(--app-surface);
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-lg);
  user-select: none;
}

.expediente-notas-panel .allergy-form,
.expediente-notas-panel .nota-entrada textarea,
.expediente-notas-panel .nota-entrada button {
  user-select: auto;
}

.notas-panel-inner {
  display: flex;
  flex-direction: column;
}

.notas-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid #E2E8F0;
  background: #F8F5FF;
  border-left: 3px solid #8B5CF6;
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
}

.expediente-notas-collapsed {
  flex: 0 0 32px;
  position: sticky;
  top: var(--space-4);
  display: flex;
  align-items: center;
  justify-content: center;
  background: #F8F5FF;
  border: 1px solid #E2E8F0;
  border-left: 3px solid #8B5CF6;
  border-radius: var(--radius-lg);
  padding: var(--space-4) var(--space-2);
  cursor: pointer;
  color: #8B5CF6;
  min-height: 60px;
}

.expediente-notas-collapsed:hover {
  background: #F0ECFF;
}

/* Responsive: móvil apilado */
@media (max-width: 768px) {
  .expediente-layout {
    flex-direction: column;
  }
  .expediente-main {
    flex: 1 1 100%;
  }
  .expediente-notas-panel {
    flex: 1 1 100%;
    position: static;
    max-height: none;
  }
  .expediente-notas-collapsed {
    display: none;
  }
  .expediente-layout--collapsed .expediente-main {
    flex: 1 1 100%;
  }
}
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

/* Sección base */
.seccion {
  background: var(--app-surface);
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-lg);
  overflow: hidden;
  margin-bottom: var(--space-6);
}

.seccion-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid #E2E8F0;
}
.seccion-header--clickable {
  cursor: pointer;
  user-select: none;
}
.seccion-header--clickable:hover {
  background-color: rgba(0,0,0,0.02);
}
.seccion-chevron {
  font-size: 11px;
  color: var(--text-secondary);
  transition: transform 0.2s;
  margin-left: var(--space-2);
}
.seccion-chevron--open {
  transform: rotate(90deg);
}

.seccion-titulo {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.seccion-titulo h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
}

.seccion-icono {
  display: flex;
  align-items: center;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.seccion-count {
  font-size: 12px;
  font-weight: 600;
  background: #F1F5F9;
  color: var(--text-secondary);
  border-radius: 999px;
  padding: 1px 8px;
}

.seccion--alergias .seccion-header {
  background: #FFFBF5;
  border-left: 3px solid #F97316;
}

.seccion--recetas .seccion-header {
  background: #F5F8FF;
  border-left: 3px solid var(--color-clinical-blue, #3B82F6);
}

.seccion--consultas .seccion-header {
  background: #F2FDFB;
  border-left: 3px solid var(--color-turquoise, #00DFA2);
}

.rx-grid {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.con-lista {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-6);
}
.con-card {
  background: var(--app-bg);
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-md);
  padding: var(--space-4) var(--space-5);
  cursor: pointer;
  transition: border-color 0.15s;
}
.con-card:hover {
  border-color: var(--color-turquoise);
}
.rx-card {
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid #F1F5F9;
}

.rx-card:last-child { border-bottom: none; }

.rx-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2px;
}

.rx-medicamento {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
}

.rx-card-acciones {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.rx-estado {
  font-size: 11px;
  font-weight: 600;
  background: #DCFCE7;
  color: #166534;
  border-radius: 999px;
  padding: 1px 8px;
}
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
}
.btn-reimprimir-sm:hover {
  background: var(--color-clinical-blue, #3B82F6);
  color: #fff;
}

.rx-dosis-text {
  font-size: 13px;
  color: var(--text-secondary);
}

.rx-dx {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 2px;
}

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

.nota-contenido {
  font-size: 15px;
  color: var(--text-primary);
  line-height: 1.7;
  white-space: pre-wrap;
}

.vitals-row { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: var(--space-2); }
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

.rx-chip {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: 13px;
  color: var(--text-secondary);
  background: #F8F4FF;
  border: 1px solid #E5D9FF;
  border-radius: var(--radius-sm);
  padding: var(--space-1) var(--space-3);
  margin-top: var(--space-2);
  width: fit-content;
}
.rx-dosis { color: var(--text-secondary); font-size: 12px; }
.rx-chip-label {
  font-size: 11px;
  font-weight: 700;
  color: var(--color-clinical-blue, #3B82F6);
  background: #E8EFFF;
  border-radius: var(--radius-sm);
  padding: 1px 6px;
}

.safety-bar {
  display: flex; align-items: center; gap: var(--space-2);
  background: #FFF7ED; border: 1px solid #FED7AA;
  border-radius: var(--radius-md); padding: var(--space-3) var(--space-4);
  margin-bottom: var(--space-4); font-size: 13px;
}
.safety-label { font-weight: 700; color: #C2410C; }
.safety-icon { color: #C2410C; flex-shrink: 0; }
.allergy-chip {
  background: #FEF3C7; border: 1px solid #FDE68A;
  border-radius: 999px; padding: 2px 10px;
  font-size: 12px; font-weight: 600; color: #92400E;
}

.allergy-form {
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid #E2E8F0;
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
  padding: var(--space-3) var(--space-6);
  border-bottom: 1px solid #F1F5F9;
}
.allergy-item:last-child { border-bottom: none; }
.allergy-meta { display: flex; align-items: flex-start; justify-content: space-between; gap: var(--space-2); }
.allergy-acciones { display: flex; gap: var(--space-2); flex-shrink: 0; }
.allergy-main { display: flex; align-items: center; gap: var(--space-2); margin-bottom: 2px; }
.allergy-agente { font-weight: 600; font-size: 14px; color: var(--text-primary); }
.allergy-badge { font-size: 11px; font-weight: 600; border-radius: 999px; padding: 1px 8px; }
.allergy-badge.leve { background: #DCFCE7; color: #166534; }
.allergy-badge.moderada { background: #FEF9C3; color: #854D0E; }
.allergy-badge.grave { background: #FEE2E2; color: #991B1B; }
.allergy-sub { font-size: 13px; color: var(--text-secondary); }
.allergy-certeza { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }
.allergy-edit-form { display: flex; flex-direction: column; gap: var(--space-2); padding: var(--space-2) 0; }
.allergy-edit-acciones { display: flex; gap: var(--space-2); margin-top: var(--space-1); }

.btn-primary {
  font-family: var(--font-brand);
  background: var(--action-primary-bg);
  color: var(--action-primary-text);
  border: none;
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-md);
  font-size: 14px; font-weight: 600;
  cursor: pointer; text-decoration: none;
}
.btn-accion {
  font-size: 12px; color: var(--color-clinical-blue);
  text-decoration: none; background: transparent;
  border: 1px solid #E2E8F0;
  padding: 2px 10px; border-radius: var(--radius-sm);
  cursor: pointer; transition: border-color 0.15s;
  font-family: var(--font-body);
}
.btn-accion:hover { border-color: var(--color-clinical-blue); }

.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.state-empty-sm { color: var(--text-secondary); font-size: 14px; padding: var(--space-6); }
.alert-error {
  background: #FFF0F3; border: 1px solid var(--color-error);
  border-radius: var(--radius-sm); padding: var(--space-3);
  font-size: 14px; color: var(--color-error);
}
</style>
