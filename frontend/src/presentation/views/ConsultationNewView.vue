<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import { consultationRepository } from '@/infrastructure/repositories/consultationRepository'
import { evidenceRepository } from '@/infrastructure/repositories/evidenceRepository'
import { prescriptionRepository } from '@/infrastructure/repositories/prescriptionRepository'
import type { Patient } from '@/domain/types/patient'

const router = useRouter()
const route = useRoute()
const patientId = route.query.patient as string | undefined

const patients = ref<Patient[]>([])
const selectedPatientId = ref(patientId ?? '')
const patient = ref<Patient | null>(null)
const loading = ref(false)
const error = ref('')

// Signos vitales
const vitals = ref({ ta: '', fc: '', fr: '', temp: '', peso: '', talla: '', sao2: '' })

// Nota clínica
const nota = ref('')

// Receta opcional
const showRx = ref(false)
const rx = ref({ medicamento_generico: '', dosis: '', diagnostico: '', indicaciones: '', seguimiento: '' })

onMounted(async () => {
  if (!patientId) {
    patients.value = await patientRepository.list()
  } else {
    patient.value = await patientRepository.get(patientId)
  }
})

async function save() {
  error.value = ''
  const pid = selectedPatientId.value || patientId
  if (!pid) { error.value = 'Selecciona un paciente'; return }
  if (!nota.value.trim()) { error.value = 'La nota clínica es obligatoria'; return }
  if (showRx.value && (!rx.value.medicamento_generico.trim() || !rx.value.dosis.trim())) {
    error.value = 'Medicamento y dosis son obligatorios en la receta'
    return
  }

  loading.value = true
  try {
    // 1. Crear consulta (signos vitales)
    const con = await consultationRepository.create(pid, vitals.value)

    // 2. Crear nota vinculada a la consulta
    await evidenceRepository.draft({
      subject_ref: pid,
      content: JSON.stringify({ type: 'note', text: nota.value, consultation_id: con.id }),
      ta: vitals.value.ta,
      fc: vitals.value.fc,
      fr: vitals.value.fr,
      temp: vitals.value.temp,
      peso: vitals.value.peso,
      talla: vitals.value.talla,
      sao2: vitals.value.sao2,
    })

    // 3. Crear receta si aplica
    if (showRx.value && rx.value.medicamento_generico.trim()) {
      const draft = await prescriptionRepository.create(pid, rx.value)
      await prescriptionRepository.emit(draft.id)
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
            <div class="vital-row"><label>T/A</label><input v-model="vitals.ta" class="input-field" placeholder="120/80" /></div>
            <div class="vital-row"><label>FC</label><input v-model="vitals.fc" class="input-field" placeholder="72 lpm" /></div>
            <div class="vital-row"><label>FR</label><input v-model="vitals.fr" class="input-field" placeholder="16 rpm" /></div>
            <div class="vital-row"><label>Temp</label><input v-model="vitals.temp" class="input-field" placeholder="36.5°C" /></div>
            <div class="vital-row"><label>Peso</label><input v-model="vitals.peso" class="input-field" placeholder="70 kg" /></div>
            <div class="vital-row"><label>Talla</label><input v-model="vitals.talla" class="input-field" placeholder="1.70 m" /></div>
            <div class="vital-row"><label>SAO2</label><input v-model="vitals.sao2" class="input-field" placeholder="98%" /></div>
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
</style>
