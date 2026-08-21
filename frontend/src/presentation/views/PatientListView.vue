<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import type { Patient } from '@/domain/types/patient'

const patients = ref<Patient[]>([])
const loading = ref(true)
const error = ref('')
const search = ref('')
const sortBy = ref<'nombre' | 'expediente' | 'reciente'>('nombre')

onMounted(async () => {
  try { patients.value = await patientRepository.list() }
  catch (e: any) { error.value = e.message }
  finally { loading.value = false }
})

const sexoLabel: Record<string, string> = { M: 'Masculino', F: 'Femenino', I: 'Indeterminado' }

function calcEdad(fechaNac: string): number {
  const hoy = new Date()
  const nac = new Date(fechaNac)
  let edad = hoy.getFullYear() - nac.getFullYear()
  if (hoy.getMonth() < nac.getMonth() ||
    (hoy.getMonth() === nac.getMonth() && hoy.getDate() < nac.getDate())) edad--
  return edad
}

const filtered = computed(() => {
  let list = [...patients.value]
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(p =>
      p.nombre.toLowerCase().includes(q) ||
      p.num_expediente.toLowerCase().includes(q)
    )
  }
  if (sortBy.value === 'nombre') {
    list.sort((a, b) => a.nombre.localeCompare(b.nombre))
  } else if (sortBy.value === 'expediente') {
    list.sort((a, b) => a.num_expediente.localeCompare(b.num_expediente))
  } else {
    list.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  }
  return list
})

function bucketLabel(iso: string): string {
  const d = new Date(iso)
  const hoy = new Date()
  const ayer = new Date(); ayer.setDate(hoy.getDate() - 1)
  const inicioSemana = new Date(); inicioSemana.setDate(hoy.getDate() - hoy.getDay()); inicioSemana.setHours(0,0,0,0)
  if (d.toDateString() === hoy.toDateString()) return 'Hoy'
  if (d.toDateString() === ayer.toDateString()) return 'Ayer'
  if (d >= inicioSemana) return 'Esta semana'
  return d.toLocaleDateString('es-MX', { month: 'long', year: 'numeric' }).replace(/^\w/, c => c.toUpperCase())
}

const grouped = computed(() => {
  if (sortBy.value === 'nombre') {
    const map = new Map<string, typeof filtered.value>()
    for (const p of filtered.value) {
      const letra = p.nombre[0]?.toUpperCase() ?? '#'
      if (!map.has(letra)) map.set(letra, [])
      map.get(letra)!.push(p)
    }
    return [...map.entries()].map(([label, items]) => ({ label, items }))
  }
  if (sortBy.value === 'reciente') {
    const map = new Map<string, typeof filtered.value>()
    for (const p of filtered.value) {
      const label = bucketLabel(p.created_at)
      if (!map.has(label)) map.set(label, [])
      map.get(label)!.push(p)
    }
    return [...map.entries()].map(([label, items]) => ({ label, items }))
  }
  return [{ label: '', items: filtered.value }]
})
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Pacientes</h2>
          <p class="page-sub" v-if="!loading">
            {{ patients.length }} paciente{{ patients.length !== 1 ? 's' : '' }} registrado{{ patients.length !== 1 ? 's' : '' }}
          </p>
        </div>
        <RouterLink to="/patients/new" class="btn-primary">+ Nuevo paciente</RouterLink>
      </div>

      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>

      <template v-else>
        <div class="controls" v-if="patients.length > 0">
          <input
            v-model="search"
            type="text"
            placeholder="Buscar por nombre o expediente..."
            class="search-input"
          />
          <select v-model="sortBy" class="sort-select">
            <option value="nombre">Alfabético</option>
            <option value="expediente">Más antiguo</option>
            <option value="reciente">Más reciente</option>
          </select>
        </div>

        <div v-if="filtered.length === 0 && patients.length === 0" class="state-empty">
          <p>Sin pacientes registrados.</p>
          <RouterLink to="/patients/new" class="btn-primary">Registrar primer paciente</RouterLink>
        </div>

        <div v-else-if="filtered.length === 0" class="state-empty">
          <p>Sin resultados para "{{ search }}"</p>
        </div>

        <div v-else class="grupos">
          <div v-for="grupo in grouped" :key="grupo.label" class="grupo">
            <div v-if="grupo.label" class="grupo-label">{{ grupo.label }}</div>
            <div class="patient-list">
              <RouterLink
                v-for="p in grupo.items"
                :key="p.id"
                :to="`/patients/${p.id}`"
                class="patient-card"
              >
                <div class="card-main">
                  <span class="patient-name">{{ p.nombre }}</span>
                  <span class="patient-exp">{{ p.num_expediente }}</span>
                </div>
                <div class="card-meta">
                  <span>{{ calcEdad(p.fecha_nacimiento) }} años</span>
                  <span>·</span>
                  <span>{{ sexoLabel[p.sexo] }}</span>
                </div>
              </RouterLink>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { width: 100%; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-4); }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: var(--space-1); }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; text-decoration: none; white-space: nowrap; }
.controls { display: flex; gap: var(--space-3); margin-bottom: var(--space-4); align-items: center; }
.search-input { flex: 1; font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-surface); outline: none; }
.search-input:focus { border-color: var(--color-turquoise); }
.sort-select { font-family: var(--font-body); font-size: 13px; color: var(--text-secondary); background: var(--app-surface); border: 1.5px solid #E2E8F0; border-radius: var(--radius-sm); padding: var(--space-2) var(--space-3); outline: none; cursor: pointer; transition: border-color 0.15s; }
.sort-select:focus { border-color: var(--color-turquoise); color: var(--text-primary); }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); display: flex; flex-direction: column; align-items: center; gap: var(--space-4); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-md); padding: var(--space-4); color: var(--color-error); font-size: 14px; }
.grupos { display: flex; flex-direction: column; gap: var(--space-6); }
.grupo { display: flex; flex-direction: column; gap: var(--space-3); }
.grupo-label { font-size: 11px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; color: var(--text-secondary); padding-bottom: var(--space-1); border-bottom: 1px solid #E2E8F0; }
.patient-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: var(--space-3); }
.patient-card {
  background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4); text-decoration: none; transition: border-color 0.15s, box-shadow 0.15s;
  display: flex; flex-direction: column; gap: 3px;
}
.patient-card:hover { border-color: var(--color-turquoise); box-shadow: 0 2px 8px rgba(0,200,212,0.08); }
.card-main { display: flex; align-items: baseline; justify-content: space-between; gap: var(--space-2); }
.patient-name { font-weight: 700; font-size: 13px; color: var(--text-primary); }
.patient-exp { font-size: 11px; color: var(--text-secondary); font-family: monospace; }
.card-meta { font-size: 11px; color: var(--text-secondary); display: flex; gap: var(--space-2); }
</style>
