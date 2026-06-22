<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { evidenceRepository } from '@/infrastructure/repositories/evidenceRepository'

const router = useRouter()
const subjectId = ref('')
const notes = ref('')
const error = ref('')
const loading = ref(false)

async function save() {
  error.value = ''
  if (!subjectId.value.trim() || !notes.value.trim()) {
    error.value = 'Todos los campos son obligatorios'
    return
  }
  loading.value = true
  try {
    const ev = await evidenceRepository.draft({
      subject_id: subjectId.value,
      notes: notes.value,
    })
    router.push(`/evidence/${ev.id}`)
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
          <h2>Nueva nota clínica</h2>
          <p class="page-sub">Los campos se guardan automáticamente.</p>
        </div>
        <RouterLink to="/evidence" class="btn-back">← Volver</RouterLink>
      </div>

      <div class="card">
        <div class="form-group">
          <label for="subject">ID del paciente / expediente</label>
          <input id="subject" v-model="subjectId" type="text" placeholder="pac-001" />
        </div>
        <div class="form-group">
          <label for="notes">Nota clínica</label>
          <textarea id="notes" v-model="notes" rows="8" placeholder="Ingrese la nota clínica..." maxlength="2000" />
          <span class="char-count">{{ notes.length }} / 2000</span>
        </div>
        <div class="alert-error" v-if="error">{{ error }}</div>
        <div class="form-actions">
          <button class="btn-primary" @click="save" :disabled="loading">
            {{ loading ? 'Guardando...' : 'Guardar nota' }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 720px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-6); }
.page-sub { color: var(--text-secondary); font-size: 14px; margin-top: var(--space-1); }
.btn-back { color: var(--color-clinical-blue); font-size: 14px; text-decoration: none; padding-top: 4px; }
.card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-4); }
.form-group { display: flex; flex-direction: column; gap: var(--space-2); }
label { font-size: 14px; font-weight: 500; color: var(--text-primary); }
input, textarea { font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-bg); resize: vertical; outline: none; }
input:focus, textarea:focus { border-color: var(--color-turquoise); }
.char-count { font-size: 12px; color: var(--text-secondary); text-align: right; }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
.form-actions { display: flex; align-items: center; justify-content: space-between; }
.badge-draft { background: #F1F5F9; color: var(--text-secondary); font-size: 12px; font-weight: 600; padding: 2px 10px; border-radius: 99px; }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-3) var(--space-6); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
