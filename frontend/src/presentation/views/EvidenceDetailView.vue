<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { evidenceRepository } from '@/infrastructure/repositories/evidenceRepository'
import type { Evidence } from '@/domain/types/evidence'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const ev = ref<Evidence | null>(null)
const loading = ref(true)
const error = ref('')
const actionError = ref('')
const actionLoading = ref(false)

const showVoidForm = ref(false)
const reasonCode = ref('')
const replacementId = ref('')

const reasonOptions = [
  { code: 'RC-VOID-001', label: 'Error detectado en el contenido' },
  { code: 'RC-VOID-002', label: 'La información requiere actualización' },
  { code: 'RC-VOID-003', label: 'Anulación solicitada formalmente' },
  { code: 'RC-VOID-004', label: 'Decisión administrativa documentada' },
]

onMounted(async () => {
  try { ev.value = await evidenceRepository.get(id) }
  catch (e: any) { error.value = e.message }
  finally { loading.value = false }
})

const canEmit = computed(() => ev.value?.state === 'draft')
const canVoid = computed(() => ev.value?.state === 'issued' || ev.value?.state === 'locked')
const canExport = computed(() => ev.value?.state === 'issued' || ev.value?.state === 'locked')

const stateLabel: Record<string, string> = {
  draft: 'Borrador', issued: 'Emitida', locked: 'Bloqueada', voided: 'Anulada',
}
const stateClass: Record<string, string> = {
  draft: 'state-draft', issued: 'state-issued', locked: 'state-locked', voided: 'state-voided',
}

function formatDate(d: string | null) {
  if (!d) return '—'
  return new Date(d).toLocaleString('es-MX')
}

async function emit() {
  if (!confirm('¿Emitir y bloquear esta nota? La acción es irreversible.')) return
  actionError.value = ''
  actionLoading.value = true
  try {
    ev.value = await evidenceRepository.emit(id)
  } catch (e: any) { actionError.value = e.message }
  finally { actionLoading.value = false }
}

async function voidEvidence() {
  if (!reasonCode.value) { actionError.value = 'Selecciona un motivo'; return }
  actionError.value = ''
  actionLoading.value = true
  try {
    if (replacementId.value.trim()) {
      const result = await evidenceRepository.replace(id, {
        reason_code: reasonCode.value,
        replacement_id: replacementId.value,
      })
      ev.value = result.voided
    } else {
      ev.value = await evidenceRepository.void(id, { reason_code: reasonCode.value })
    }
    showVoidForm.value = false
  } catch (e: any) { actionError.value = e.message }
  finally { actionLoading.value = false }
}

async function exportEvidence() {
  try {
    const blob = await evidenceRepository.export(id)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `export_${id}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e: any) { actionError.value = e.message }
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Detalle de nota clínica</h2>
          <p class="page-sub">{{ id }}</p>
        </div>
        <RouterLink to="/evidence" class="btn-back">← Volver</RouterLink>
      </div>

      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>

      <template v-else-if="ev">
        <div class="card">
          <div class="detail-row">
            <span class="detail-label">Estado</span>
            <span :class="['state-badge', stateClass[ev.state]]">{{ stateLabel[ev.state] }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Tenant</span>
            <span class="detail-value">{{ ev.tenant_id }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Creado</span>
            <span class="detail-value">{{ formatDate(ev.created_at) }}</span>
          </div>
          <div class="detail-row" v-if="ev.issued_at">
            <span class="detail-label">Emitido</span>
            <span class="detail-value">{{ formatDate(ev.issued_at) }}</span>
          </div>
          <div class="detail-row" v-if="ev.voided_at">
            <span class="detail-label">Anulado</span>
            <span class="detail-value">{{ formatDate(ev.voided_at) }}</span>
          </div>
          <div class="detail-row" v-if="ev.replaced_by_id">
            <span class="detail-label">Reemplazado por</span>
            <RouterLink :to="`/evidence/${ev.replaced_by_id}`" class="link">{{ ev.replaced_by_id }}</RouterLink>
          </div>
        </div>

        <div class="alert-error" v-if="actionError">{{ actionError }}</div>

        <div class="actions" v-if="canEmit || canVoid || canExport">
          <button v-if="canEmit" class="btn-emit" @click="emit" :disabled="actionLoading">
            Emitir y bloquear
          </button>
          <button v-if="canVoid" class="btn-void" @click="showVoidForm = !showVoidForm">
            Anular
          </button>
          <button v-if="canExport" class="btn-export" @click="exportEvidence">
            Export legal
          </button>
        </div>

        <div class="void-form card" v-if="showVoidForm">
          <h3>Anular nota</h3>
          <div class="form-group">
            <label>Motivo de anulación</label>
            <select v-model="reasonCode">
              <option value="">— Selecciona un motivo —</option>
              <option v-for="r in reasonOptions" :key="r.code" :value="r.code">
                {{ r.code }} — {{ r.label }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>ID de reemplazo (opcional)</label>
            <input v-model="replacementId" type="text" placeholder="Dejar vacío para solo anular" />
          </div>
          <div class="form-actions">
            <button class="btn-back-link" @click="showVoidForm = false">Cancelar</button>
            <button class="btn-void" @click="voidEvidence" :disabled="actionLoading">
              {{ actionLoading ? 'Procesando...' : 'Confirmar anulación' }}
            </button>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 720px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-6); }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: var(--space-1); font-family: monospace; }
.btn-back { color: var(--color-clinical-blue); font-size: 14px; text-decoration: none; }
.card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-4); margin-bottom: var(--space-4); }
.detail-row { display: flex; align-items: center; gap: var(--space-4); }
.detail-label { width: 100px; font-size: 13px; color: var(--text-secondary); flex-shrink: 0; }
.detail-value { font-size: 14px; color: var(--text-primary); }
.state-badge { font-size: 12px; font-weight: 600; padding: 2px 10px; border-radius: 99px; }
.state-draft { background: #F1F5F9; color: var(--text-secondary); }
.state-issued, .state-locked { background: #E6FAF5; color: var(--color-jade); }
.state-voided { background: #FFF0F3; color: var(--color-error); }
.link { color: var(--color-clinical-blue); font-size: 14px; }
.actions { display: flex; gap: var(--space-3); margin-bottom: var(--space-4); }
.btn-emit { background: var(--color-jade); color: var(--color-obsidian); border: none; padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); font-weight: 600; font-size: 14px; cursor: pointer; }
.btn-void { background: var(--color-error); color: #fff; border: none; padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); font-weight: 600; font-size: 14px; cursor: pointer; }
.btn-export { background: var(--color-turquoise); color: var(--color-obsidian); border: none; padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); font-weight: 600; font-size: 14px; cursor: pointer; }
.void-form { border-color: var(--color-error); }
.form-group { display: flex; flex-direction: column; gap: var(--space-2); }
label { font-size: 14px; font-weight: 500; color: var(--text-primary); }
select, input { font-family: var(--font-body); padding: var(--space-3); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 14px; color: var(--text-primary); background: var(--app-bg); outline: none; }
select:focus, input:focus { border-color: var(--color-error); }
.form-actions { display: flex; align-items: center; justify-content: flex-end; gap: var(--space-3); }
.btn-back-link { background: transparent; border: none; color: var(--text-secondary); font-size: 14px; cursor: pointer; }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); margin-bottom: var(--space-4); }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
</style>
