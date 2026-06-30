<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import { consultationRepository } from '@/infrastructure/repositories/consultationRepository'
import { evidenceRepository } from '@/infrastructure/repositories/evidenceRepository'
import { prescriptionRepository } from '@/infrastructure/repositories/prescriptionRepository'
import { useAuthStore } from '@/app/stores/auth'
import type { Patient } from '@/domain/types/patient'

const router = useRouter()
const route = useRoute()
const patientId = route.query.patient as string | undefined
const auth = useAuthStore()

const patients = ref<Patient[]>([])
const selectedPatientId = ref(patientId ?? '')
const patient = ref<Patient | null>(null)
const loading = ref(false)
const error = ref('')

// Signos vitales — T/A separado en sistólica/diastólica
const vitals = ref({ ta_s: '', ta_d: '', fc: '', fr: '', temp: '', peso: '', talla: '', sao2: '' })

// Formatea temperatura: 365 → 36.5
function fmtTemp(e: Event) {
  const raw = (e.target as HTMLInputElement).value.replace(/\D/g, '').slice(0, 3)
  vitals.value.temp = raw.length >= 3 ? raw.slice(0, 2) + '.' + raw.slice(2) : raw
}

// Formatea talla: 170 → 1.70
function fmtTalla(e: Event) {
  const raw = (e.target as HTMLInputElement).value.replace(/\D/g, '').slice(0, 3)
  vitals.value.talla = raw.length >= 2 ? raw.slice(0, 1) + '.' + raw.slice(1) : raw
}

// Solo números para campos enteros
function onlyNum(field: 'fc' | 'fr' | 'peso' | 'sao2' | 'ta_s' | 'ta_d', e: Event) {
  vitals.value[field] = (e.target as HTMLInputElement).value.replace(/\D/g, '').slice(0, 3)
}

// Nota clínica
const nota = ref('')

// Receta opcional
const showRx = ref(false)
const showConfirmNoRx = ref(false)
const pendingSave = ref(false)
const rx = ref({ medicamento_generico: '', dosis: '', diagnostico: '', indicaciones: '', seguimiento: '' })

onMounted(async () => {
  if (!patientId) {
    patients.value = await patientRepository.list()
  } else {
    patient.value = await patientRepository.get(patientId)
  }
})

function confirmNoRx() {
  showConfirmNoRx.value = false
  pendingSave.value = true
  save()
}

async function save() {
  error.value = ''
  const pid = selectedPatientId.value || patientId
  if (!pid) { error.value = 'Selecciona un paciente'; return }
  if (!nota.value.trim()) { error.value = 'La nota clínica es obligatoria'; return }
  if (showRx.value && (!rx.value.medicamento_generico.trim() || !rx.value.dosis.trim())) {
    error.value = 'Medicamento y dosis son obligatorios en la receta'
    return
  }

  // Confirmar si guarda sin receta
  if (!showRx.value && !pendingSave.value) {
    showConfirmNoRx.value = true
    return
  }
  pendingSave.value = false

  loading.value = true
  try {
    // 1. Construir payload de signos vitales
    const vPayload = {
      ta: vitals.value.ta_s && vitals.value.ta_d ? vitals.value.ta_s + '/' + vitals.value.ta_d : '',
      fc: vitals.value.fc,
      fr: vitals.value.fr,
      temp: vitals.value.temp,
      peso: vitals.value.peso,
      talla: vitals.value.talla,
      sao2: vitals.value.sao2,
    }

    // 2. Crear consulta
    const con = await consultationRepository.create(pid, vPayload)

    // 3. Crear nota vinculada a la consulta
    await evidenceRepository.draft({
      subject_ref: pid,
      content: JSON.stringify({ type: 'note', text: nota.value, consultation_id: con.id }),
      ta: vPayload.ta,
      fc: vPayload.fc,
      fr: vPayload.fr,
      temp: vPayload.temp,
      peso: vPayload.peso,
      talla: vPayload.talla,
      sao2: vPayload.sao2,
    })

    // 4. Crear receta si aplica, vinculada a la consulta
    let rxId = ''
    if (showRx.value && rx.value.medicamento_generico.trim()) {
      const draft = await prescriptionRepository.create(pid, { ...rx.value, consultation_id: con.id })
      await prescriptionRepository.emit(draft.id)
      rxId = draft.id
    }

    // 5. Abrir PDF en nueva pestaña si hay receta
    if (rxId) {
      window.open(`/api/v1/prescriptions/${rxId}/print?token=${auth.token}`, '_blank')
    }

    router.push(`/patients/${pid}`)
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Nueva consulta</h2>
          <p class="page-sub" v-if="patient">
            Paciente: <strong>{{ patient.nombre }}</strong> · Exp. {{ patient.num_expediente }}
          </p>
        </div>
        <RouterLink :to="patientId ? `/patients/${patientId}` : '/consultations'" class="btn-back">← Cancelar</RouterLink>
      </div>

      <!-- Modal de confirmación: guardar sin receta -->
      <div v-if="showConfirmNoRx" class="modal-overlay">
        <div class="modal">
          <h3 class="modal-title">¿Guardar sin receta?</h3>
          <p class="modal-body">Esta consulta no tiene receta electrónica adjunta. ¿Deseas continuar?</p>
          <div class="modal-actions">
            <button class="btn-secondary" @click="showConfirmNoRx = false">Agregar receta</button>
            <button class="btn-primary" @click="confirmNoRx">Guardar sin receta</button>
          </div>
        </div>
      </div>

      <div class="card">
        <!-- Selector de paciente si no viene de un paciente -->
        <div class="form-group" v-if="!patientId">
          <label>Paciente *</label>
          <select v-model="selectedPatientId" class="input-field">
            <option value="">— Selecciona un paciente —</option>
            <option v-for="p in patients" :key="p.id" :value="p.id">
              {{ p.nombre }} · {{ p.num_expediente }}
            </option>
          </select>
        </div>

        <!-- Signos vitales -->
        <div class="section">
          <h3 class="section-title">Signos vitales <span class="optional">(opcional)</span></h3>
          <div class="vitals-grid">
            <div class="vital-row">
              <label>T/A</label>
              <div class="vital-ta">
                <input :value="vitals.ta_s" @input="onlyNum('ta_s', $event)" class="input-field vital-half" placeholder="120" inputmode="numeric" maxlength="3" />
                <span class="vital-sep">/</span>
                <input :value="vitals.ta_d" @input="onlyNum('ta_d', $event)" class="input-field vital-half" placeholder="80" inputmode="numeric" maxlength="3" />
              </div>
            </div>
            <div class="vital-row">
              <label>FC</label>
              <input :value="vitals.fc" @input="onlyNum('fc', $event)" class="input-field" placeholder="72" inputmode="numeric" maxlength="3" />
            </div>
            <div class="vital-row">
              <label>FR</label>
              <input :value="vitals.fr" @input="onlyNum('fr', $event)" class="input-field" placeholder="16" inputmode="numeric" maxlength="3" />
            </div>
            <div class="vital-row">
              <label>Temp</label>
              <input :value="vitals.temp" @input="fmtTemp($event)" class="input-field" placeholder="36.5" inputmode="numeric" maxlength="4" />
            </div>
            <div class="vital-row">
              <label>Peso</label>
              <input :value="vitals.peso" @input="onlyNum('peso', $event)" class="input-field" placeholder="70" inputmode="numeric" maxlength="3" />
            </div>
            <div class="vital-row">
              <label>Talla</label>
              <input :value="vitals.talla" @input="fmtTalla($event)" class="input-field" placeholder="1.70" inputmode="numeric" maxlength="4" />
            </div>
            <div class="vital-row">
              <label>SAO2</label>
              <input :value="vitals.sao2" @input="onlyNum('sao2', $event)" class="input-field" placeholder="98" inputmode="numeric" maxlength="3" />
            </div>
          </div>
        </div>

        <!-- Nota clínica -->
        <div class="section">
          <h3 class="section-title">Nota clínica *</h3>
          <textarea
            v-model="nota"
            rows="8"
            class="input-field"
            placeholder="Ingrese la nota clínica en lenguaje técnico-médico..."
            maxlength="2000"
          />
          <span class="char-count">{{ nota.length }} / 2000</span>
        </div>

        <!-- Receta opcional -->
        <div class="section">
          <div class="section-header">
            <h3 class="section-title">Receta electrónica</h3>
            <button type="button" class="btn-toggle" @click="showRx = !showRx">
              {{ showRx ? 'Sin receta' : '+ Agregar receta' }}
            </button>
          </div>
          <div v-if="showRx" class="rx-form">
            <div class="form-group">
              <label>Medicamento genérico *</label>
              <input v-model="rx.medicamento_generico" class="input-field" placeholder="p. ej. Paracetamol" />
            </div>
            <div class="form-group">
              <label>Dosis *</label>
              <input v-model="rx.dosis" class="input-field" placeholder="p. ej. 500mg cada 8h por 3 días" />
            </div>
            <div class="form-group">
              <label>Diagnóstico <span class="optional">(opcional)</span></label>
              <input v-model="rx.diagnostico" class="input-field" placeholder="p. ej. Faringitis aguda" />
            </div>
            <div class="form-group">
              <label>Indicaciones <span class="optional">(opcional)</span></label>
              <input v-model="rx.indicaciones" class="input-field" placeholder="p. ej. Reposo e hidratación" />
            </div>
            <div class="form-group">
              <label>Seguimiento <span class="optional">(opcional)</span></label>
              <input v-model="rx.seguimiento" class="input-field" placeholder="p. ej. Control en 7 días" />
            </div>
          </div>
        </div>

        <div class="alert-error" v-if="error">{{ error }}</div>

        <div class="form-actions">
          <button class="btn-primary" @click="save" :disabled="loading">
            {{ loading ? 'Guardando...' : 'Guardar consulta' }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 720px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-6); }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: var(--space-1); }
.btn-back { color: var(--color-clinical-blue); font-size: 14px; text-decoration: none; }
.card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-6); }
.section { display: flex; flex-direction: column; gap: var(--space-3); border-top: 1px solid #F1F5F9; padding-top: var(--space-4); }
.section:first-child { border-top: none; padding-top: 0; }
.section-header { display: flex; align-items: center; justify-content: space-between; }
.section-title { font-size: 14px; font-weight: 600; color: var(--text-primary); margin: 0; }
.optional { font-size: 12px; font-weight: 400; color: var(--text-secondary); }
.vitals-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: var(--space-2); }
.vital-row { display: flex; flex-direction: column; gap: 4px; }
.vital-row label { font-size: 11px; font-weight: 600; color: var(--text-secondary); }
.vital-ta { display: flex; align-items: center; gap: 4px; }
.vital-half { flex: 1; min-width: 0; }
.vital-sep { font-size: 14px; font-weight: 600; color: var(--text-secondary); flex-shrink: 0; }
.form-group { display: flex; flex-direction: column; gap: var(--space-2); }
.form-group label { font-size: 14px; font-weight: 500; color: var(--text-primary); }
.input-field { font-family: var(--font-body); padding: var(--space-2) var(--space-3); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 14px; color: var(--text-primary); background: var(--app-bg); outline: none; }
.input-field:focus { border-color: var(--color-turquoise); }
textarea.input-field { resize: vertical; font-size: 15px; padding: var(--space-3) var(--space-4); }
.char-count { font-size: 12px; color: var(--text-secondary); text-align: right; }
.btn-toggle { font-size: 13px; font-weight: 600; color: var(--color-turquoise); background: none; border: none; cursor: pointer; padding: 0; }
.rx-form { display: flex; flex-direction: column; gap: var(--space-3); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
.form-actions { display: flex; justify-content: flex-end; }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-3) var(--space-6); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { font-family: var(--font-brand); background: transparent; color: var(--text-primary); border: 1.5px solid #E2E8F0; padding: var(--space-3) var(--space-6); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; }
.btn-secondary:hover { border-color: var(--color-turquoise); }

/* Modal de confirmación */
.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center;
  z-index: 100;
}
.modal {
  background: var(--app-surface);
  border-radius: var(--radius-lg);
  padding: var(--space-8);
  max-width: 400px; width: 90%;
  display: flex; flex-direction: column; gap: var(--space-4);
  box-shadow: 0 8px 32px rgba(0,0,0,0.18);
}
.modal-title { font-size: 16px; font-weight: 700; color: var(--text-primary); margin: 0; }
.modal-body { font-size: 14px; color: var(--text-secondary); line-height: 1.6; margin: 0; }
.modal-actions { display: flex; gap: var(--space-3); justify-content: flex-end; }
</style>
