<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { consultationRepository } from '@/infrastructure/repositories/consultationRepository'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import type { Consultation } from '@/domain/types/consultation'
import type { Patient } from '@/domain/types/patient'

const router = useRouter()
const consultations = ref<Consultation[]>([])
const patients = ref<Record<string, Patient>>({})
const loading = ref(true)
const error = ref('')
const search = ref('')
const sortBy = ref<'reciente' | 'antiguo' | 'az'>('reciente')

const sorted = computed(() => {
  const list = consultations.value.filter(c => {
    const pac = patients.value[c.patient_id]?.nombre?.toLowerCase() ?? ''
    const q = search.value.toLowerCase()
    return pac.includes(q)
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
  return [...list].sort((a, b) =>
    new Date(b.issued_at ?? b.created_at).getTime() - new Date(a.issued_at ?? a.created_at).getTime()
  )
})

onMounted(async () => {
  try {
    const [cons, pats] = await Promise.all([
      consultationRepository.listAll(),
      patientRepository.list(),
    ])
    consultations.value = cons
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
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Consultas</h2>
          <p class="page-sub">{{ consultations.length }} consulta{{ consultations.length !== 1 ? 's' : '' }} registrada{{ consultations.length !== 1 ? 's' : '' }}</p>
        </div>
        <RouterLink to="/consultations/new" class="btn-primary">+ Nueva consulta</RouterLink>
      </div>

      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>
      <div v-else>
        <div class="controls">
          <input v-model="search" type="text" placeholder="Buscar por paciente..." class="search-input" />
          <div class="sort-buttons">
            <button :class="['btn-sort', sortBy === 'az' && 'active']" @click="sortBy = 'az'">Alfabético</button>
            <button :class="['btn-sort', sortBy === 'antiguo' && 'active']" @click="sortBy = 'antiguo'">Antiguo</button>
            <button :class="['btn-sort', sortBy === 'reciente' && 'active']" @click="sortBy = 'reciente'">Reciente</button>
          </div>
        </div>

        <div v-if="sorted.length === 0" class="state-empty">
          No hay consultas. Crea una desde el perfil de un paciente o desde aquí.
        </div>

        <div v-else class="consultation-list">
          <RouterLink
            v-for="c in sorted"
            :key="c.id"
            :to="`/patients/${c.patient_id}`"
            class="consultation-card"
          >
            <div class="con-header">
              <span class="con-paciente">{{ patients[c.patient_id]?.nombre ?? c.patient_id }}</span>
              <span class="con-fecha">{{ formatDate(c.issued_at ?? c.created_at) }}</span>
            </div>
            <div v-if="c.ta || c.fc || c.temp" class="con-vitals">
              <span v-if="c.ta">T/A: {{ c.ta }}</span>
              <span v-if="c.fc">FC: {{ c.fc }}</span>
              <span v-if="c.temp">Temp: {{ c.temp }}</span>
              <span v-if="c.peso">Peso: {{ c.peso }}</span>
            </div>
          </RouterLink>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 780px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-4); }
.page-sub { font-size: 13px; color: var(--text-secondary); margin-top: 2px; }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-2) var(--space-4); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; text-decoration: none; }
.controls { display: flex; gap: var(--space-3); margin-bottom: var(--space-4); align-items: center; }
.search-input { flex: 1; font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-surface); outline: none; }
.search-input:focus { border-color: var(--color-turquoise); }
.sort-buttons { display: flex; gap: var(--space-1); }
.btn-sort { background: var(--app-surface); border: 1.5px solid #E2E8F0; color: var(--text-secondary); padding: var(--space-2) var(--space-3); border-radius: var(--radius-sm); font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.15s; white-space: nowrap; }
.btn-sort:hover { border-color: var(--color-turquoise); color: var(--color-turquoise); }
.btn-sort.active { background: var(--color-obsidian); border-color: var(--color-obsidian); color: #fff; }
.consultation-list { display: flex; flex-direction: column; gap: var(--space-3); }
.consultation-card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-lg); padding: var(--space-4) var(--space-6); cursor: pointer; transition: border-color 0.15s; text-decoration: none; display: block; }
.consultation-card:hover { border-color: var(--color-turquoise); }
.con-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--space-2); }
.con-paciente { font-weight: 600; font-size: 14px; color: var(--text-primary); }
.con-fecha { font-size: 12px; color: var(--text-secondary); }
.con-vitals { display: flex; gap: var(--space-3); font-size: 13px; color: var(--text-secondary); flex-wrap: wrap; }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
</style>
