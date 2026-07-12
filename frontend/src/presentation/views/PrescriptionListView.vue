<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { prescriptionRepository } from '@/infrastructure/repositories/prescriptionRepository'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import { useAuthStore } from '@/app/stores/auth'
import type { Prescription } from '@/domain/types/prescription'
import type { Patient } from '@/domain/types/patient'

const router = useRouter()
const auth = useAuthStore()
const prescriptions = ref<Prescription[]>([])
const patients = ref<Record<string, Patient>>({})
const loading = ref(true)
const error = ref('')
const search = ref('')
const sortBy = ref<'reciente' | 'antiguo' | 'az'>('reciente')

const sorted = computed(() => {
  const list = prescriptions.value.filter(rx => {
    const pac = patients.value[rx.patient_id]?.nombre?.toLowerCase() ?? ''
    const med = rx.medicamento_generico.toLowerCase()
    const q = search.value.toLowerCase()
    return pac.includes(q) || med.includes(q)
  })
  if (sortBy.value === 'az') {
    return [...list].sort((a, b) => {
      const na = patients.value[a.patient_id]?.nombre ?? ''
      const nb = patients.value[b.patient_id]?.nombre ?? ''
      return na.localeCompare(nb)
    })
  }
  if (sortBy.value === 'antiguo') {
    return [...list].sort((a, b) =>
      new Date(a.issued_at ?? a.created_at).getTime() - new Date(b.issued_at ?? b.created_at).getTime()
    )
  }
  // reciente — más nueva primero (default)
  return [...list].sort((a, b) =>
    new Date(b.issued_at ?? b.created_at).getTime() - new Date(a.issued_at ?? a.created_at).getTime()
  )
})

const filtered = computed(() =>
  prescriptions.value.filter(rx => {
    const pac = patients.value[rx.patient_id]?.nombre?.toLowerCase() ?? ''
    const med = rx.medicamento_generico.toLowerCase()
    const q = search.value.toLowerCase()
    return pac.includes(q) || med.includes(q)
  })
)

onMounted(async () => {
  try {
    const [rxs, pats] = await Promise.all([
      prescriptionRepository.listAll(),
      patientRepository.list(),
    ])
    prescriptions.value = rxs
    patients.value = Object.fromEntries(pats.map(p => [p.id, p]))
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('es-MX', {
    year: 'numeric', month: 'long', day: 'numeric',
    hour: '2-digit', minute: '2-digit'
  })
}

function goToPrescription(prescriptionId: string) {
  router.push(`/prescriptions/${prescriptionId}`)
}

function imprimirRx(rxId: string) {
  if (!auth.token) return
  window.open(`/api/v1/prescriptions/${rxId}/print?token=${auth.token}`, '_blank')
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Recetas</h2>
          <p class="page-sub">{{ prescriptions.length }} receta{{ prescriptions.length !== 1 ? 's' : '' }} expedida{{ prescriptions.length !== 1 ? 's' : '' }}</p>
        </div>
        <RouterLink to="/prescriptions/new" class="btn-primary">+ Nueva receta</RouterLink>
      </div>
      <div class="controls">
        <input
          v-model="search"
          type="text"
          placeholder="Buscar por paciente o medicamento..."
          class="search-input"
        />
        <div class="sort-buttons">
          <button :class="['btn-sort', sortBy === 'az' && 'active']" @click="sortBy = 'az'">Alfabético</button>
          <button :class="['btn-sort', sortBy === 'antiguo' && 'active']" @click="sortBy = 'antiguo'">Antiguo</button>
          <button :class="['btn-sort', sortBy === 'reciente' && 'active']" @click="sortBy = 'reciente'">Reciente</button>
        </div>
      </div>

      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>
      <div v-else-if="prescriptions.length === 0" class="state-empty">
        No hay recetas emitidas. Crea una desde el perfil de un paciente.
      </div>

      <div v-else class="rx-list">
        <div
          v-for="rx in sorted"
          :key="rx.id"
          class="rx-item"
          @click="goToPrescription(rx.id)"
        >
          <div class="rx-header">
            <span class="rx-paciente">
              {{ patients[rx.patient_id]?.nombre ?? rx.patient_id }}
            </span>
            <span class="rx-fecha">{{ formatDate(rx.issued_at ?? rx.created_at) }}</span>
          </div>
          <div class="rx-med-row">
            <div>
              <div class="rx-medicamento">{{ rx.medicamento_generico }}</div>
              <div class="rx-dosis">{{ rx.dosis }}</div>
              <div v-if="rx.diagnostico" class="rx-diagnostico">Dx: {{ rx.diagnostico }}</div>
            </div>
            <button class="btn-reimprimir-sm" @click.stop="imprimirRx(rx.id)">
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="6 9 6 2 18 2 18 9"/>
                <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/>
                <rect x="6" y="14" width="12" height="8"/>
              </svg>
              Imprimir
            </button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { width: 100%; max-width: 100%; }
.rx-list { display: flex; flex-direction: column; gap: var(--space-3); }
.rx-item {
  background: var(--app-surface);
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-lg);
  padding: var(--space-4) var(--space-6);
  cursor: pointer;
  transition: border-color 0.15s;
}
.rx-item:hover { border-color: var(--color-turquoise); }
.rx-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--space-2); }
.rx-paciente { font-weight: 600; font-size: 14px; color: var(--text-primary); }
.rx-fecha { font-size: 12px; color: var(--text-secondary); }
.rx-med-row { display: flex; align-items: flex-start; justify-content: space-between; gap: var(--space-3); margin-top: var(--space-1); }
.rx-medicamento { font-size: 15px; color: var(--text-primary); font-weight: 600; }
.rx-dosis { font-size: 13px; color: var(--text-secondary); }
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
  flex-shrink: 0;
}
.btn-reimprimir-sm:hover {
  background: var(--color-clinical-blue, #3B82F6);
  color: #fff;
}
.rx-diagnostico { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }
.controls { display: flex; gap: var(--space-3); margin-bottom: var(--space-4); align-items: center; }
.search-input { flex: 1; font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-surface); outline: none; }
.search-input:focus { border-color: var(--color-turquoise); }
.sort-buttons { display: flex; gap: var(--space-1); }
.btn-sort { background: var(--app-surface); border: 1.5px solid #E2E8F0; color: var(--text-secondary); padding: var(--space-2) var(--space-3); border-radius: var(--radius-sm); font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.15s; white-space: nowrap; }
.btn-sort:hover { border-color: var(--color-turquoise); color: var(--color-turquoise); }
.btn-sort.active { background: var(--color-obsidian); border-color: var(--color-obsidian); color: #fff; }
.page-sub { font-size: 13px; color: var(--text-secondary); margin-top: 2px; }
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-4); }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-2) var(--space-4); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; text-decoration: none; }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
</style>
