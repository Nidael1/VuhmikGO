<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import { prescriptionRepository } from '@/infrastructure/repositories/prescriptionRepository'
import type { Patient } from '@/domain/types/patient'

const router = useRouter()
const patients = ref<Patient[]>([])
const patientId = ref('')
const medicamentoGenerico = ref('')
const dosis = ref('')
const diagnostico = ref('')
const indicaciones = ref('')
const seguimiento = ref('')
const error = ref('')
const saving = ref(false)

onMounted(async () => {
  try {
    patients.value = await patientRepository.list()
  } catch (e: any) {
    error.value = e.message
  }
})

async function save() {
  error.value = ''
  if (!patientId.value) { error.value = 'Selecciona un paciente'; return }
  if (!medicamentoGenerico.value.trim()) { error.value = 'El medicamento es obligatorio'; return }
  if (!dosis.value.trim()) { error.value = 'La dosis es obligatoria'; return }
  saving.value = true
  try {
    const draft = await prescriptionRepository.create(patientId.value, {
      medicamento_generico: medicamentoGenerico.value.trim(),
      dosis: dosis.value.trim(),
      diagnostico: diagnostico.value.trim() || undefined,
      indicaciones: indicaciones.value.trim() || undefined,
      seguimiento: seguimiento.value.trim() || undefined,
    })
    await prescriptionRepository.emit(draft.id)
    router.push(`/patients/${patientId.value}`)
  } catch (e: any) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Nueva receta</h2>
          <p class="page-sub">Campos mínimos NOM-024-SSA3-2012</p>
        </div>
        <RouterLink to="/prescriptions" class="btn-back">← Cancelar</RouterLink>
      </div>
      <div class="card">
        <div class="form-group">
          <label>Paciente *</label>
          <select v-model="patientId">
            <option value="">— Selecciona un paciente —</option>
            <option v-for="p in patients" :key="p.id" :value="p.id">
              {{ p.nombre }} · {{ p.num_expediente }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <label>Medicamento genérico *</label>
          <input v-model="medicamentoGenerico" type="text" placeholder="p. ej. Paracetamol" />
        </div>
        <div class="form-group">
          <label>Dosis *</label>
          <input v-model="dosis" type="text" placeholder="p. ej. 500mg cada 8h por 3 días" />
        </div>
        <div class="form-group">
          <label>Diagnóstico <span class="optional">(opcional)</span></label>
          <input v-model="diagnostico" type="text" placeholder="p. ej. Faringitis aguda" />
        </div>
        <div class="form-group">
          <label>Indicaciones <span class="optional">(opcional)</span></label>
          <input v-model="indicaciones" type="text" placeholder="p. ej. Reposo e hidratación" />
        </div>
        <div class="form-group">
          <label>Seguimiento <span class="optional">(opcional)</span></label>
          <input v-model="seguimiento" type="text" placeholder="p. ej. Control en 7 días" />
        </div>
        <div class="alert-error" v-if="error">{{ error }}</div>
        <div class="form-actions">
          <button class="btn-primary" @click="save" :disabled="saving">
            {{ saving ? 'Emitiendo...' : 'Emitir receta' }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 600px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-6); }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: var(--space-1); }
.btn-back { color: var(--color-clinical-blue); font-size: 14px; text-decoration: none; }
.card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-4); }
.form-group { display: flex; flex-direction: column; gap: var(--space-2); }
label { font-size: 14px; font-weight: 500; color: var(--text-primary); }
input, select { font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-bg); outline: none; }
input:focus, select:focus { border-color: var(--color-turquoise); }
.optional { font-size: 12px; color: var(--text-secondary); font-weight: 400; }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
.form-actions { display: flex; justify-content: flex-end; }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-3) var(--space-6); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
