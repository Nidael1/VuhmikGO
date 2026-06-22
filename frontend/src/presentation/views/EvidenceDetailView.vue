<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
const saving = ref(false)

onMounted(async () => {
  try { ev.value = await evidenceRepository.get(id) }
  catch (e: any) { error.value = e.message }
  finally { loading.value = false }
})

function formatDate(d: string | null) {
  if (!d) return '—'
  return new Date(d).toLocaleString('es-MX', {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit'
  })
}

async function exportEvidence() {
  try {
    const blob = await evidenceRepository.export(id)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `nota_${id}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e: any) { error.value = e.message }
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Nota clínica</h2>
          <p class="page-sub">{{ id }}</p>
        </div>
        <RouterLink to="/evidence" class="btn-back">← Volver</RouterLink>
      </div>

      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>

      <template v-else-if="ev">
        <div class="card">
          <div class="detail-row">
            <span class="detail-label">Fecha</span>
            <span class="detail-value">{{ formatDate(ev.created_at) }}</span>
          </div>
          <div class="detail-row" v-if="ev.issued_at">
            <span class="detail-label">Emitida</span>
            <span class="detail-value">{{ formatDate(ev.issued_at) }}</span>
          </div>
        </div>

        <div class="actions">
          <RouterLink :to="`/evidence/${ev.id}/editar`" class="btn-edit">
            ✏️ Editar nota
          </RouterLink>
          <button class="btn-export" @click="exportEvidence">
            ⬇ Descargar
          </button>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 720px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-6); }
.page-sub { color: var(--text-secondary); font-size: 12px; margin-top: var(--space-1); font-family: monospace; }
.btn-back { color: var(--color-clinical-blue); font-size: 14px; text-decoration: none; }
.card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-4); margin-bottom: var(--space-4); }
.detail-row { display: flex; align-items: center; gap: var(--space-4); }
.detail-label { width: 80px; font-size: 13px; color: var(--text-secondary); flex-shrink: 0; }
.detail-value { font-size: 14px; color: var(--text-primary); }
.actions { display: flex; gap: var(--space-3); }
.btn-edit { background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); font-weight: 600; font-size: 14px; cursor: pointer; text-decoration: none; }
.btn-export { background: transparent; border: 1.5px solid #E2E8F0; color: var(--text-secondary); padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); font-size: 14px; cursor: pointer; }
.btn-export:hover { border-color: var(--color-turquoise); color: var(--color-turquoise); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
</style>
