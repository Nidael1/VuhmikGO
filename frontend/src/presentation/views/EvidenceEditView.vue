<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { evidenceRepository } from '@/infrastructure/repositories/evidenceRepository'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import { useAuthStore } from '@/app/stores/auth'
import type { Patient } from '@/domain/types/patient'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const id = route.params.id as string

const notes = ref('')
const patient = ref<Patient | null>(null)
const loading = ref(true)
const saving = ref(false)
const error = ref('')

onMounted(async () => {
  try {
    const ev = await evidenceRepository.get(id)
    try { const blob = JSON.parse(ev.content); notes.value = blob.text || ev.content } catch { notes.value = ev.content }
    if (ev.subject_ref) {
      try { patient.value = await patientRepository.get(ev.subject_ref) }
      catch { /* paciente no encontrado, no es error critico */ }
    }
  } catch (e: any) { error.value = e.message }
  finally { loading.value = false }
})

async function save() {
  if (!notes.value.trim()) { error.value = 'La nota no puede estar vacía'; return }
  saving.value = true
  error.value = ''
  try {
    const res = await fetch(`/api/v1/evidence/${id}/edit`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${auth.token || ''}`,
      },
      body: JSON.stringify({ content: JSON.stringify({type:'note',text:notes.value}) }),
    })
    const data = await res.json()
    if (data.error) { error.value = data.error.message; return }
    // Regresar al paciente si existe
    if (patient.value) {
      router.push(`/patients/${patient.value.id}`)
    } else {
      router.push(`/evidence/${data.data.id}`)
    }
  } catch (e: any) { error.value = e.message }
  finally { saving.value = false }
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Editar nota clínica</h2>
          <p class="page-sub" v-if="patient">
            Paciente: <strong>{{ patient.nombre }}</strong>
            · Exp. {{ patient.num_expediente }}
          </p>
          <p class="page-sub" v-else>Los cambios quedan registrados automáticamente</p>
        </div>
        <button class="btn-back" @click="() => patient ? router.push(`/patients/${patient!.id}`) : router.push(`/evidence/${id}`)">
          ← Cancelar
        </button>
      </div>
      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else class="card">
        <div class="form-group">
          <label for="notes">Nota clínica</label>
          <textarea id="notes" v-model="notes" rows="10"
            placeholder="Actualiza la nota clínica..." maxlength="2000" />
          <span class="char-count">{{ notes.length }} / 2000</span>
        </div>
        <div class="alert-error" v-if="error">{{ error }}</div>
        <div class="form-actions">
          <button class="btn-primary" @click="save" :disabled="saving">
            {{ saving ? 'Guardando...' : 'Guardar cambios' }}
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
.btn-back { background: transparent; border: none; color: var(--color-clinical-blue); font-size: 14px; cursor: pointer; padding: 0; }
.card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-4); }
.form-group { display: flex; flex-direction: column; gap: var(--space-2); }
label { font-size: 14px; font-weight: 500; color: var(--text-primary); }
textarea { font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-bg); resize: vertical; outline: none; }
textarea:focus { border-color: var(--color-turquoise); }
.char-count { font-size: 12px; color: var(--text-secondary); text-align: right; }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
.form-actions { display: flex; justify-content: flex-end; }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-3) var(--space-6); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
</style>
