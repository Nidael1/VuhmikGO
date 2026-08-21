<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { appointmentRepository } from '@/infrastructure/repositories/appointmentRepository'
import type { Appointment } from '@/domain/types/appointment'

const appointments = ref<Appointment[]>([])
const loading = ref(true)
const error = ref('')
const search = ref('')
const filtroEstado = ref<'todas' | 'hoy' | 'proximas' | 'canceladas'>('hoy')

function localDateToday(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
}

const filtradas = computed(() => {
  const hoy = localDateToday()
  return appointments.value
    .filter(a => {
      const nombre = (a.patient_nombre ?? '').toLowerCase()
      const q = search.value.toLowerCase()
      if (q && !nombre.includes(q)) return false
      const fecha = a.scheduled_at.slice(0, 10)
      if (filtroEstado.value === 'hoy') return fecha === hoy
      if (filtroEstado.value === 'proximas') return fecha > hoy && a.state === 'scheduled'
      if (filtroEstado.value === 'canceladas') return a.state === 'cancelled' || a.state === 'no_show'
      return true
    })
    .sort((a, b) => new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime())
})

onMounted(async () => {
  try {
    appointments.value = await appointmentRepository.listAll()
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

async function cancelar(id: string) {
  if (!confirm('¿Cancelar esta cita?')) return
  try {
    await appointmentRepository.cancel(id)
    const a = appointments.value.find(x => x.id === id)
    if (a) a.state = 'cancelled'
  } catch {
    alert('Error al cancelar la cita')
  }
}

async function noShow(id: string) {
  if (!confirm('¿Marcar como no presentado?')) return
  try {
    await appointmentRepository.noShow(id)
    const a = appointments.value.find(x => x.id === id)
    if (a) a.state = 'no_show'
  } catch {
    alert('Error al actualizar la cita')
  }
}

function formatFecha(iso: string): string {
  return new Date(iso).toLocaleDateString('es-MX', {
    weekday: 'short', year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

function duracionLabel(min: number): string {
  return min < 60 ? `${min} min` : `${Math.floor(min/60)}h${min%60 ? ` ${min%60}min` : ''}`
}

const estadoLabel: Record<string, string> = {
  scheduled: 'Agendada',
  completed: 'Atendida',
  cancelled: 'Cancelada',
  no_show:   'No se presentó',
}

function bucketLabel(iso: string): string {
  const d = new Date(iso)
  const hoy = new Date()
  const manana = new Date(); manana.setDate(hoy.getDate() + 1)
  const ayer = new Date(); ayer.setDate(hoy.getDate() - 1)
  const inicioSemana = new Date(); inicioSemana.setDate(hoy.getDate() - hoy.getDay()); inicioSemana.setHours(0,0,0,0)
  const finSemana = new Date(inicioSemana); finSemana.setDate(inicioSemana.getDate() + 6)
  if (d.toDateString() === hoy.toDateString()) return 'Hoy'
  if (d.toDateString() === manana.toDateString()) return 'Mañana'
  if (d.toDateString() === ayer.toDateString()) return 'Ayer'
  if (d >= inicioSemana && d <= finSemana) return 'Esta semana'
  return d.toLocaleDateString('es-MX', { month: 'long', year: 'numeric' }).replace(/^\w/, c => c.toUpperCase())
}

const grouped = computed(() => {
  if (filtroEstado.value === 'hoy') return [{ label: '', items: filtradas.value }]
  const map = new Map<string, typeof filtradas.value>()
  for (const a of filtradas.value) {
    const label = bucketLabel(a.scheduled_at)
    if (!map.has(label)) map.set(label, [])
    map.get(label)!.push(a)
  }
  return [...map.entries()].map(([label, items]) => ({ label, items }))
})
</script>

<template>
  <AppLayout>
    <div class="page">
      <div class="page-header">
        <div>
          <h2>Citas</h2>
          <p class="page-sub">{{ appointments.length }} cita{{ appointments.length !== 1 ? 's' : '' }} registrada{{ appointments.length !== 1 ? 's' : '' }}</p>
        </div>
        <RouterLink to="/appointments/new" class="btn-primary">+ Nueva cita</RouterLink>
      </div>

      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>
      <div v-else>
        <div class="controls">
          <input v-model="search" type="text" placeholder="Buscar por paciente..." class="search-input" />
          <div class="filtros">
            <button :class="['btn-filtro', filtroEstado === 'hoy' && 'active']" @click="filtroEstado = 'hoy'">Hoy</button>
            <button :class="['btn-filtro', filtroEstado === 'proximas' && 'active']" @click="filtroEstado = 'proximas'">Próximas</button>
            <button :class="['btn-filtro', filtroEstado === 'todas' && 'active']" @click="filtroEstado = 'todas'">Todas</button>
            <button :class="['btn-filtro', filtroEstado === 'canceladas' && 'active']" @click="filtroEstado = 'canceladas'">Canceladas</button>
          </div>
        </div>

        <div v-if="filtradas.length === 0" class="state-empty">
          No hay citas en este filtro. Usa "+ Nueva cita" para agendar.
        </div>

        <div v-else class="grupos">
          <div v-for="grupo in grouped" :key="grupo.label" class="grupo">
            <div v-if="grupo.label" class="grupo-label">{{ grupo.label }}</div>
            <div class="cita-list">
          <div
            v-for="a in grupo.items"
            :key="a.id"
            class="cita-card"
            :class="`cita-card--${a.state}`"
          >
            <div class="cita-estado-bar" />
            <div class="cita-body">
              <div class="cita-header">
                <span class="cita-paciente">{{ a.patient_nombre ?? a.patient_id }}</span>
                <span :class="['cita-badge', `badge--${a.state}`]">{{ estadoLabel[a.state] ?? a.state }}</span>
              </div>
              <div class="cita-fecha">{{ formatFecha(a.scheduled_at) }} · {{ duracionLabel(a.duration_min) }}</div>
              <div v-if="a.reason" class="cita-motivo">{{ a.reason }}</div>
              <div v-if="a.state === 'scheduled'" class="cita-acciones">
                <RouterLink
                  :to="`/consultations/new?patient=${a.patient_id}&appointment=${a.id}`"
                  class="btn-accion btn-accion--primary"
                >Iniciar consulta</RouterLink>
                <button class="btn-accion btn-accion--ghost" @click="noShow(a.id)">No se presentó</button>
                <button class="btn-accion btn-accion--danger" @click="cancelar(a.id)">Cancelar</button>
              </div>
            </div>
          </div>
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
.controls { display: flex; gap: var(--space-3); margin-bottom: var(--space-4); align-items: center; flex-wrap: wrap; }
.search-input { flex: 1; min-width: 180px; font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-surface); outline: none; }
.search-input:focus { border-color: var(--color-turquoise); }
.filtros { display: flex; gap: var(--space-1); flex-wrap: wrap; }
.btn-filtro { background: var(--app-surface); border: 1.5px solid #E2E8F0; color: var(--text-secondary); padding: var(--space-2) var(--space-3); border-radius: var(--radius-sm); font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.15s; white-space: nowrap; }
.btn-filtro:hover { border-color: var(--color-turquoise); color: var(--color-turquoise); }
.btn-filtro.active { background: var(--color-obsidian); border-color: var(--color-obsidian); color: #fff; }
.grupos { display: flex; flex-direction: column; gap: var(--space-6); }
.grupo { display: flex; flex-direction: column; gap: var(--space-3); }
.grupo-label { font-size: 11px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; color: var(--text-secondary); padding-bottom: var(--space-1); border-bottom: 1px solid #E2E8F0; }
.cita-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: var(--space-3); }
.cita-card {
  background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md);
  display: flex; overflow: hidden; transition: border-color 0.15s, box-shadow 0.15s;
}
.cita-card:hover { border-color: var(--color-turquoise); box-shadow: 0 2px 8px rgba(0,200,212,0.08); }
.cita-estado-bar { width: 3px; flex-shrink: 0; }
.cita-card--scheduled .cita-estado-bar { background: #3b82f6; }
.cita-card--completed .cita-estado-bar { background: #22c55e; }
.cita-card--cancelled .cita-estado-bar { background: #94a3b8; }
.cita-card--no_show   .cita-estado-bar { background: #eab308; }
.cita-body { flex: 1; padding: var(--space-3) var(--space-4); display: flex; flex-direction: column; gap: 3px; }
.cita-header { display: flex; align-items: center; justify-content: space-between; }
.cita-paciente { font-weight: 700; font-size: 13px; color: var(--text-primary); }
.cita-badge { font-size: 11px; font-weight: 600; padding: 2px 8px; border-radius: 999px; }
.badge--scheduled { background: #dbeafe; color: #1d4ed8; }
.badge--completed { background: #dcfce7; color: #15803d; }
.badge--cancelled { background: #f1f5f9; color: #64748b; }
.badge--no_show   { background: #fef9c3; color: #854d0e; }
.cita-fecha { font-size: 11px; color: var(--text-secondary); }
.cita-motivo { font-size: 11px; color: var(--text-secondary); font-style: italic; }
.cita-acciones { display: flex; gap: var(--space-2); flex-wrap: wrap; align-items: center; margin-top: var(--space-2); }
.btn-accion { cursor: pointer; text-decoration: none; transition: all 0.15s; border: none; background: none; font-family: var(--font-body); }

/* Primaria: peso alto — acción esperada */
.btn-accion--primary {
  font-size: 13px; font-weight: 600;
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-sm);
  background: var(--color-obsidian); color: #fff;
  border: 1.5px solid var(--color-obsidian);
}
.btn-accion--primary:hover { opacity: 0.85; }

/* Ghost: peso medio — acción secundaria */
.btn-accion--ghost {
  font-size: 13px; font-weight: 500;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  background: transparent;
  border: 1.5px solid #E2E8F0;
  color: var(--text-secondary);
}
.btn-accion--ghost:hover { border-color: var(--color-turquoise); color: var(--color-turquoise); }

/* Danger: peso mínimo — acción destructiva discreta */
.btn-accion--danger {
  font-size: 12px; font-weight: 500;
  padding: 0; background: none; border: none;
  color: var(--color-error); opacity: 0.6;
  text-decoration: underline; text-decoration-color: transparent;
}
.btn-accion--danger:hover { opacity: 1; text-decoration-color: var(--color-error); }
.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
</style>
