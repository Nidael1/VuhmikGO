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

function bucketLabel(iso: string): string {
  const d = new Date(iso)
  const hoy = new Date()
  const ayer = new Date(); ayer.setDate(hoy.getDate() - 1)
  const inicioSemana = new Date(); inicioSemana.setDate(hoy.getDate() - hoy.getDay())
  inicioSemana.setHours(0, 0, 0, 0)
  if (d.toDateString() === hoy.toDateString()) return 'Hoy'
  if (d.toDateString() === ayer.toDateString()) return 'Ayer'
  if (d >= inicioSemana) return 'Esta semana'
  return d.toLocaleDateString('es-MX', { month: 'long', year: 'numeric' })
    .replace(/^\w/, c => c.toUpperCase())
}

const grouped = computed(() => {
  if (sortBy.value === 'az') return [{ label: '', items: sorted.value }]
  const map = new Map<string, typeof sorted.value>()
  for (const c of sorted.value) {
    const label = bucketLabel(c.issued_at ?? c.created_at)
    if (!map.has(label)) map.set(label, [])
    map.get(label)!.push(c)
  }
  return [...map.entries()].map(([label, items]) => ({ label, items }))
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
          <select v-model="sortBy" class="sort-select">
            <option value="az">Alfabético</option>
            <option value="antiguo">Más antiguo</option>
            <option value="reciente">Más reciente</option>
          </select>
        </div>

        <div v-if="sorted.length === 0" class="state-empty">
          No hay consultas. Crea una desde el perfil de un paciente o desde aquí.
        </div>

        <div v-else class="grupos">
          <div v-for="grupo in grouped" :key="grupo.label" class="grupo">
            <div v-if="grupo.label" class="grupo-label">{{ grupo.label }}</div>
            <div class="consultation-list">
              <RouterLink
                v-for="c in grupo.items"
                :key="c.id"
                :to="`/consultations/${c.id}`"
                class="consultation-card"
              >
                <div class="con-header">
                  <span class="con-paciente">{{ patients[c.patient_id]?.nombre ?? c.patient_id }}</span>
                  <span class="con-fecha">{{ formatDate(c.issued_at ?? c.created_at) }}</span>
                </div>
                <div class="con-vitals">
                  <span class="vital">{{ c.ta ? `T/A: ${c.ta} mmHg` : '' }}</span>
                  <span class="vital">{{ c.fc ? `FC: ${c.fc} lpm` : '' }}</span>
                  <span class="vital">{{ c.fr ? `FR: ${c.fr} rpm` : '' }}</span>
                  <span class="vital">{{ c.temp ? `Temp: ${c.temp}°C` : '' }}</span>
                  <span class="vital">{{ c.peso ? `Peso: ${c.peso} kg` : '' }}</span>
                  <span class="vital">{{ c.talla ? `Talla: ${c.talla} m` : '' }}</span>
                  <span class="vital">{{ c.sao2 ? `SAO₂: ${c.sao2}%` : '' }}</span>
                  <span class="vital"></span>
                </div>
              </RouterLink>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { width: 100%; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-4); }
.page-sub { font-size: 13px; color: var(--text-secondary); margin-top: 2px; }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-2) var(--space-4); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; text-decoration: none; }
.controls { display: flex; gap: var(--space-3); margin-bottom: var(--space-4); align-items: center; }
.search-input { flex: 1; font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-surface); outline: none; }
.search-input:focus { border-color: var(--color-turquoise); }
.sort-select { font-family: var(--font-body); font-size: 13px; color: var(--text-secondary); background: var(--app-surface); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-3) var(--space-4); outline: none; cursor: pointer; transition: border-color 0.15s; }
.sort-select:focus { border-color: var(--color-turquoise); color: var(--text-primary); }
.grupos { display: flex; flex-direction: column; gap: var(--space-6); }
.grupo { display: flex; flex-direction: column; gap: var(--space-3); }
.grupo-label {
  font-size: 11px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase;
  color: var(--text-secondary); padding-bottom: var(--space-1);
  border-bottom: 1px solid #E2E8F0;
}
.consultation-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: var(--space-3); }
.consultation-card {
  background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4); cursor: pointer; transition: border-color 0.15s, box-shadow 0.15s;
  text-decoration: none; display: flex; flex-direction: column; justify-content: space-between;
}
.consultation-card:hover { border-color: var(--color-turquoise); box-shadow: 0 2px 8px rgba(0,200,212,0.08); }
.con-header { display: flex; flex-direction: column; gap: 2px; }
.con-paciente { font-weight: 700; font-size: 13px; color: var(--text-primary); }
.con-fecha { font-size: 11px; color: var(--text-secondary); }
.con-vitals { display: grid; grid-template-columns: repeat(2, 1fr); gap: 2px 8px; font-size: 11px; color: var(--text-secondary); margin-top: 4px; }
.vital { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; min-height: 15px; }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
</style>
