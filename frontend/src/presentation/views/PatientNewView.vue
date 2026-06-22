<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'

const router = useRouter()
const nombre = ref('')
const fechaNacimiento = ref('')
const sexo = ref('M')
const error = ref('')
const saving = ref(false)

async function save() {
  error.value = ''
  if (!nombre.value.trim()) { error.value = 'El nombre es obligatorio'; return }
  if (!fechaNacimiento.value) { error.value = 'La fecha de nacimiento es obligatoria'; return }
  saving.value = true
  try {
    const p = await patientRepository.create({
      nombre: nombre.value,
      fecha_nacimiento: fechaNacimiento.value,
      sexo: sexo.value,
    })
    router.push(`/patients/${p.id}`)
  } catch (e: any) { error.value = e.message }
  finally { saving.value = false }
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Nuevo paciente</h2>
          <p class="page-sub">Campos requeridos por NOM-004-SSA3-2012</p>
        </div>
        <RouterLink to="/patients" class="btn-back">← Cancelar</RouterLink>
      </div>
      <div class="card">
        <div class="form-group">
          <label>Nombre completo *</label>
          <input v-model="nombre" type="text" placeholder="Nombre completo del paciente" />
        </div>
        <div class="form-group">
          <label>Fecha de nacimiento *</label>
          <input v-model="fechaNacimiento" type="date" />
        </div>
        <div class="form-group">
          <label>Sexo *</label>
          <select v-model="sexo">
            <option value="M">Masculino</option>
            <option value="F">Femenino</option>
            <option value="I">Indeterminado</option>
          </select>
        </div>
        <div class="alert-error" v-if="error">{{ error }}</div>
        <div class="form-actions">
          <button class="btn-primary" @click="save" :disabled="saving">
            {{ saving ? 'Guardando...' : 'Registrar paciente' }}
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
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
.form-actions { display: flex; justify-content: flex-end; }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-3) var(--space-6); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
