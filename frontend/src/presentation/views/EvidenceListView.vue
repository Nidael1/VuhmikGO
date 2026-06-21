<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { evidenceRepository } from '@/infrastructure/repositories/evidenceRepository'
import type { Evidence } from '@/domain/types/evidence'

const router = useRouter()
const items = ref<Evidence[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    items.value = await evidenceRepository.list()
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

const stateLabel: Record<string, string> = {
  draft: 'Borrador',
  issued: 'Emitida',
  locked: 'Bloqueada',
  voided: 'Anulada',
}

const stateClass: Record<string, string> = {
  draft: 'state-draft',
  issued: 'state-issued',
  locked: 'state-locked',
  voided: 'state-voided',
}

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('es-MX', {
    year: 'numeric', month: 'short', day: 'numeric',
  })
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Expedientes clínicos</h2>
          <p class="page-sub">Historial de notas clínicas del consultorio</p>
        </div>
        <RouterLink to="/evidence/new" class="btn-primary">+ Nueva nota</RouterLink>
      </div>

      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>
      <div v-else-if="items.length === 0" class="state-empty">
        <p>Sin registros aún.</p>
        <RouterLink to="/evidence/new" class="btn-primary">Crear primera nota</RouterLink>
      </div>

      <div v-else class="evidence-list">
        <RouterLink
          v-for="item in items"
          :key="item.id"
          :to="`/evidence/${item.id}`"
          class="evidence-card"
        >
          <div class="card-main">
            <span class="card-id">{{ item.id }}</span>
            <span :class="['state-badge', stateClass[item.state]]">
              {{ stateLabel[item.state] }}
            </span>
          </div>
          <div class="card-meta">
            <span>Creado: {{ formatDate(item.created_at) }}</span>
            <span v-if="item.issued_at">· Emitido: {{ formatDate(item.issued_at) }}</span>
          </div>
        </RouterLink>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 800px; }

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: var(--space-6);
}

.page-sub {
  color: var(--text-secondary);
  font-size: 14px;
  margin-top: var(--space-1);
}

.btn-primary {
  font-family: var(--font-brand);
  background: var(--action-primary-bg);
  color: var(--action-primary-text);
  border: none;
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  white-space: nowrap;
}

.state-empty {
  color: var(--text-secondary);
  text-align: center;
  padding: var(--space-8);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
}

.alert-error {
  background: #FFF0F3;
  border: 1px solid var(--color-error);
  border-radius: var(--radius-md);
  padding: var(--space-4);
  color: var(--color-error);
  font-size: 14px;
}

.evidence-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.evidence-card {
  display: block;
  background: var(--app-surface);
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-md);
  padding: var(--space-4) var(--space-6);
  text-decoration: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.evidence-card:hover {
  border-color: var(--color-turquoise);
  box-shadow: 0 2px 8px rgba(0,200,212,0.1);
}

.card-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}

.card-id {
  font-family: var(--font-brand);
  font-weight: 600;
  font-size: 15px;
  color: var(--text-primary);
}

.card-meta {
  font-size: 13px;
  color: var(--text-secondary);
  display: flex;
  gap: var(--space-2);
}

.state-badge {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 99px;
}

.state-draft   { background: #F1F5F9; color: var(--text-secondary); }
.state-issued  { background: #E6FAF5; color: var(--color-jade); }
.state-locked  { background: #E6FAF5; color: var(--color-jade); }
.state-voided  { background: #FFF0F3; color: var(--color-error); }
</style>
