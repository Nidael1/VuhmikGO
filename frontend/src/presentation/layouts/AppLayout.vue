<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAuthStore } from '@/app/stores/auth'
import { useRouter, useRoute } from 'vue-router'
import TermsModal from '@/presentation/components/TermsModal.vue'
import { consultationRepository } from '@/infrastructure/repositories/consultationRepository'
import { appointmentRepository } from '@/infrastructure/repositories/appointmentRepository'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import { prescriptionRepository } from '@/infrastructure/repositories/prescriptionRepository'
import type { Consultation } from '@/domain/types/consultation'
import type { Appointment } from '@/domain/types/appointment'
import type { Patient } from '@/domain/types/patient'
import type { Prescription } from '@/domain/types/prescription'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const menuAbierto = ref(false)
const showTerms = ref(false)

function cerrarMenu() { menuAbierto.value = false }

async function logout() {
  if (auth.refreshToken) {
    try {
      await fetch('/api/v1/auth/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: auth.refreshToken }),
      })
    } catch { }
  }
  auth.clearSession()
  router.push('/login')
}

// --- Agenda de hoy ---
const consultasHoy = ref<Consultation[]>([])
const citasHoy = ref<Appointment[]>([])

function localDateToday(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
}

async function cargarAgenda() {
  if (!auth.token) return
  try {
    const [consultas, citas] = await Promise.all([
      consultationRepository.listToday(localDateToday()),
      appointmentRepository.listToday(localDateToday()),
    ])
    consultasHoy.value = consultas
    citasHoy.value = citas
  } catch { /* silencioso — el widget no bloquea la app */ }
}

// Citas scheduled sin consulta asociada (no completadas aún)
const citasPendientes = computed(() =>
  citasHoy.value.filter(c => c.state === 'scheduled')
)
const totalHoy = computed(() => consultasHoy.value.length + citasPendientes.value.length)
const completadasHoy = computed(() => consultasHoy.value.filter(c => c.state === 'issued').length)
const hayEntradas = computed(() => consultasHoy.value.length > 0 || citasPendientes.value.length > 0)

function horaCorta(iso: string): string {
  const d = new Date(iso)
  return `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
}

function nombreCorto(nombre?: string): string {
  if (!nombre) return '—'
  const partes = nombre.trim().split(' ')
  return partes.length >= 2 ? `${partes[0]} ${partes[1][0]}.` : partes[0]
}

// --- Buscador maestro ---
const busqueda = ref('')
const todosLosPacientes = ref<Patient[]>([])
const todasLasConsultas = ref<Consultation[]>([])
const todasLasRecetas = ref<Prescription[]>([])

async function cargarDatosBuscador() {
  if (!auth.token) return
  try {
    const [pats, cons, rxs] = await Promise.all([
      patientRepository.list(),
      consultationRepository.listAll(),
      prescriptionRepository.listAll(),
    ])
    todosLosPacientes.value = pats
    todasLasConsultas.value = cons
    todasLasRecetas.value = rxs
  } catch { /* silencioso */ }
}

function fechaTexto(iso: string): string {
  return new Date(iso).toLocaleDateString('es-MX', { day: 'numeric', month: 'short', year: 'numeric' })
}

const resultados = computed(() => {
  const q = busqueda.value.trim().toLowerCase()
  if (q.length < 2) return null
  const pacienteMap = Object.fromEntries(todosLosPacientes.value.map(p => [p.id, p]))

  const pacientes = todosLosPacientes.value
    .filter(p => {
      const fields = [p.nombre, p.num_expediente, p.fecha_nacimiento ?? ''].join(' ').toLowerCase()
      return fields.includes(q)
    })
    .slice(0, 8)
    .map(p => ({ label: p.nombre, sub: p.num_expediente, ruta: `/patients/${p.id}` }))

  const consultas = todasLasConsultas.value
    .filter(c => {
      const nombre = pacienteMap[c.patient_id]?.nombre?.toLowerCase() ?? ''
      const fecha = fechaTexto(c.issued_at ?? c.created_at).toLowerCase()
      const vitales = [
        c.ta ?? '', c.fc ?? '', c.fr ?? '', c.temp ?? '',
        c.peso ?? '', c.talla ?? '', c.sao2 ?? '',
      ].join(' ').toLowerCase()
      return nombre.includes(q) || fecha.includes(q) || vitales.includes(q)
    })
    .slice(0, 8)
    .map(c => ({
      label: pacienteMap[c.patient_id]?.nombre ?? c.patient_id,
      sub: fechaTexto(c.issued_at ?? c.created_at) + (c.ta ? ` · T/A ${c.ta}` : '') + (c.fc ? ` · FC ${c.fc}` : ''),
      ruta: `/consultations/${c.id}`,
    }))

  const recetas = todasLasRecetas.value
    .filter(rx => {
      const nombre = pacienteMap[rx.patient_id]?.nombre?.toLowerCase() ?? ''
      const fields = [rx.medicamento_generico, rx.dosis ?? '', rx.diagnostico ?? ''].join(' ').toLowerCase()
      const fecha = fechaTexto(rx.issued_at ?? rx.created_at).toLowerCase()
      return nombre.includes(q) || fields.includes(q) || fecha.includes(q)
    })
    .slice(0, 8)
    .map(rx => ({
      label: rx.medicamento_generico,
      sub: (pacienteMap[rx.patient_id]?.nombre ?? '') + (rx.dosis ? ` · ${rx.dosis}` : ''),
      ruta: `/prescriptions/${rx.id}`,
    }))

  return { pacientes, consultas, recetas }
})

const hayResultados = computed(() =>
  resultados.value && (
    resultados.value.pacientes.length > 0 ||
    resultados.value.consultas.length > 0 ||
    resultados.value.recetas.length > 0
  )
)

onMounted(async () => {
  await cargarAgenda()
  await cargarDatosBuscador()
})
// Refresca cada vez que cambia la ruta (el médico registró una consulta y navega)
watch(() => route.fullPath, () => { cargarAgenda(); cargarDatosBuscador() })
</script>

<template>
  <div class="app-shell">
    <!-- Topbar movil -->
    <header class="topbar">
      <button class="hamburger" @click="menuAbierto = !menuAbierto" aria-label="Menu">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line v-if="menuAbierto" x1="18" y1="6" x2="6" y2="18"/><line v-if="menuAbierto" x1="6" y1="6" x2="18" y2="18"/>
          <line v-if="!menuAbierto" x1="3" y1="6" x2="21" y2="6"/><line v-if="!menuAbierto" x1="3" y1="12" x2="21" y2="12"/><line v-if="!menuAbierto" x1="3" y1="18" x2="21" y2="18"/>
        </svg>
      </button>
      <span class="topbar-brand">vuhmik</span>
    </header>

    <!-- Overlay movil -->
    <div v-if="menuAbierto" class="overlay" @click="cerrarMenu" />

    <!-- Sidebar -->
    <aside class="sidebar" :class="{ 'sidebar-open': menuAbierto }">
      <div class="sidebar-brand">
        <span class="brand-icon">V</span>
        <span class="brand-name">vuhmik</span>
      </div>

      <!-- Buscador maestro -->
      <div class="search-maestro">
        <div class="search-input-wrap">
          <svg class="search-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input
            v-model="busqueda"
            class="search-input"
            placeholder="Buscar..."
            @keydown.escape="busqueda = ''"
          />
          <button v-if="busqueda" class="search-clear" @click="busqueda = ''">✕</button>
        </div>
      </div>

      <nav class="sidebar-nav">
        <RouterLink to="/patients" class="nav-item" @click="cerrarMenu">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
          <span>Pacientes</span>
        </RouterLink>
        <RouterLink to="/consultations" class="nav-item" @click="cerrarMenu">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
          <span>Consultas</span>
        </RouterLink>
        <RouterLink to="/appointments" class="nav-item" @click="cerrarMenu">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
          <span>Citas</span>
        </RouterLink>
        <RouterLink to="/prescriptions" class="nav-item" @click="cerrarMenu">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
          <span>Recetas</span>
        </RouterLink>
      </nav>
      <!-- Agenda de hoy -->
      <div class="agenda-hoy">
        <div class="agenda-header">
          <span class="agenda-titulo">Hoy</span>
          <span class="agenda-conteo">{{ completadasHoy }}/{{ totalHoy }}</span>
        </div>
        <div v-if="!hayEntradas" class="agenda-vacio">Sin actividad registrada</div>
        <!-- Citas agendadas pendientes -->
        <RouterLink
          v-for="a in citasPendientes"
          :key="a.id"
          to="/appointments"
          class="agenda-item agenda-item--cita"
          @click="cerrarMenu"
        >
          <span class="agenda-dot dot--blue" />
          <span class="agenda-hora">{{ horaCorta(a.scheduled_at) }}</span>
          <span class="agenda-paciente">{{ nombreCorto(a.patient_nombre) }}</span>
        </RouterLink>
        <!-- Consultas del día -->
        <RouterLink
          v-for="c in consultasHoy"
          :key="c.id"
          :to="`/consultations/${c.id}`"
          class="agenda-item"
          :class="{
            'agenda-item--issued': c.state === 'issued',
            'agenda-item--voided': c.state === 'voided',
            'agenda-item--draft':  c.state === 'draft',
          }"
          @click="cerrarMenu"
        >
          <span class="agenda-dot" :class="{
            'dot--green':  c.state === 'issued',
            'dot--yellow': c.state === 'voided',
            'dot--red':    c.state === 'draft',
          }" />
          <span class="agenda-hora">{{ horaCorta(c.created_at) }}</span>
          <span class="agenda-paciente">{{ nombreCorto(c.patient_nombre) }}</span>
        </RouterLink>
      </div>

      <div class="sidebar-footer">
        <RouterLink to="/profile" class="nav-item-profile" @click="cerrarMenu">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
          <span>Mi perfil</span>
        </RouterLink>
        <div class="user-info" v-if="auth.profile">
          <span class="user-actor">{{ auth.profile.actor_id }}</span>
        </div>
        <button class="btn-terms-sidebar" @click="showTerms = true">Términos y condiciones</button>
        <button class="btn-logout" @click="logout">Cerrar sesión</button>
      </div>
    </aside>

    <TermsModal :open="showTerms" @close="showTerms = false" />

    <main class="main-content">
      <!-- Resultados de búsqueda maestro -->
      <div v-if="busqueda.length >= 2" class="search-resultados">
        <div class="search-resultados-header">
          <h2>Resultados para "{{ busqueda }}"</h2>
          <button class="search-cerrar" @click="busqueda = ''">✕ Cerrar búsqueda</button>
        </div>

        <div v-if="!hayResultados" class="search-vacio">Sin resultados para "{{ busqueda }}"</div>

        <template v-else>
          <!-- Pacientes -->
          <div v-if="resultados!.pacientes.length > 0" class="sr-grupo">
            <div class="sr-label">Pacientes</div>
            <div class="sr-grid sr-grid--pacientes">
              <RouterLink
                v-for="r in resultados!.pacientes" :key="r.ruta"
                :to="r.ruta" class="sr-card" @click="busqueda = ''"
              >
                <span class="sr-card-title">{{ r.label }}</span>
                <span class="sr-card-sub">{{ r.sub }}</span>
              </RouterLink>
            </div>
          </div>

          <!-- Consultas -->
          <div v-if="resultados!.consultas.length > 0" class="sr-grupo">
            <div class="sr-label">Consultas</div>
            <div class="sr-grid sr-grid--consultas">
              <RouterLink
                v-for="r in resultados!.consultas" :key="r.ruta"
                :to="r.ruta" class="sr-card" @click="busqueda = ''"
              >
                <span class="sr-card-title">{{ r.label }}</span>
                <span class="sr-card-sub">{{ r.sub }}</span>
              </RouterLink>
            </div>
          </div>

          <!-- Recetas -->
          <div v-if="resultados!.recetas.length > 0" class="sr-grupo">
            <div class="sr-label">Recetas</div>
            <div class="sr-grid sr-grid--recetas">
              <RouterLink
                v-for="r in resultados!.recetas" :key="r.ruta"
                :to="r.ruta" class="sr-card" @click="busqueda = ''"
              >
                <span class="sr-card-title">{{ r.label }}</span>
                <span class="sr-card-sub">{{ r.sub }}</span>
              </RouterLink>
            </div>
          </div>
        </template>
      </div>

      <!-- Vista normal -->
      <slot v-else />
    </main>
  </div>
</template>

<style scoped>
.app-shell { display: flex; min-height: 100vh; background: var(--app-bg); }

/* Topbar — solo visible en movil */
.topbar {
  display: none;
  position: fixed; top: 0; left: 0; right: 0; z-index: 100;
  height: 52px; background: var(--app-sidebar-bg);
  align-items: center; padding: 0 16px; gap: 12px;
  box-shadow: 0 1px 0 rgba(255,255,255,0.06);
}
.topbar-brand { font-family: var(--font-brand); font-weight: 700; font-size: 17px; color: var(--text-on-dark); letter-spacing: -0.02em; }
.hamburger { background: none; border: none; color: var(--text-on-dark); cursor: pointer; padding: 4px; display: flex; align-items: center; justify-content: center; border-radius: 6px; }
.hamburger:hover { background: rgba(255,255,255,0.08); }

/* Overlay movil */
.overlay { display: none; position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 150; }

/* Sidebar */
.sidebar {
  width: 240px; min-height: 100vh;
  background: var(--app-sidebar-bg);
  display: flex; flex-direction: column;
  padding: var(--space-4); gap: var(--space-6);
  position: fixed; top: 0; left: 0; z-index: 200;
}
.sidebar-brand { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-4) 0; }
.brand-icon { width: 32px; height: 32px; background: var(--color-jade); color: var(--color-obsidian); border-radius: var(--radius-sm); display: flex; align-items: center; justify-content: center; font-family: var(--font-brand); font-weight: 700; font-size: 16px; }
.brand-name { font-family: var(--font-brand); font-weight: 700; font-size: 18px; color: var(--text-on-dark); letter-spacing: -0.02em; }
.sidebar-nav { display: flex; flex-direction: column; gap: var(--space-2); flex: 1; }
.nav-item { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); color: var(--text-on-dark); opacity: 0.7; text-decoration: none; font-size: 14px; transition: all 0.15s; }
.nav-item:hover, .nav-item.router-link-active { background: rgba(255,255,255,0.08); opacity: 1; color: var(--color-turquoise); text-decoration: none; }
.sidebar-footer { display: flex; flex-direction: column; gap: var(--space-2); border-top: 1px solid rgba(255,255,255,0.08); padding-top: var(--space-4); }
.user-actor { font-size: 12px; color: var(--text-on-dark); opacity: 0.5; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.btn-logout { background: transparent; border: 1px solid rgba(255,255,255,0.15); color: var(--text-on-dark); padding: var(--space-2) var(--space-4); border-radius: var(--radius-sm); font-size: 13px; cursor: pointer; text-align: left; transition: all 0.15s; }
.btn-logout:hover { border-color: var(--color-error); color: var(--color-error); }
.btn-terms-sidebar { background: transparent; border: none; color: var(--text-on-dark); opacity: 0.4; font-size: 11px; cursor: pointer; text-align: left; padding: var(--space-1) 0; transition: opacity 0.15s; }
.btn-terms-sidebar:hover { opacity: 0.7; text-decoration: underline; }
.nav-item-profile { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-2) var(--space-3); border-radius: var(--radius-md); color: var(--text-on-dark); opacity: 0.6; text-decoration: none; font-size: 13px; transition: all 0.15s; }
.nav-item-profile:hover, .nav-item-profile.router-link-active { opacity: 1; color: var(--color-turquoise); }

/* Buscador maestro — sidebar input */
.search-maestro { padding: 0 var(--space-2); }
.search-input-wrap { display: flex; align-items: center; gap: 6px; background: rgba(255,255,255,0.07); border: 1px solid rgba(255,255,255,0.1); border-radius: var(--radius-md); padding: 6px 10px; transition: border-color 0.15s; }
.search-input-wrap:focus-within { border-color: var(--color-turquoise); }
.search-icon { color: var(--text-on-dark); opacity: 0.4; flex-shrink: 0; }
.search-input { background: none; border: none; outline: none; color: var(--text-on-dark); font-family: var(--font-body); font-size: 13px; width: 100%; }
.search-input::placeholder { color: rgba(240,246,252,0.35); }
.search-clear { background: none; border: none; color: rgba(240,246,252,0.4); cursor: pointer; font-size: 12px; line-height: 1; padding: 0; flex-shrink: 0; }
.search-clear:hover { color: var(--text-on-dark); }

/* Panel de resultados en main-content */
.search-resultados { width: 100%; }
.search-resultados-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-6); }
.search-resultados-header h2 { margin: 0; }
.search-cerrar { background: none; border: 1px solid #E2E8F0; border-radius: var(--radius-sm); padding: var(--space-1) var(--space-3); font-size: 13px; color: var(--text-secondary); cursor: pointer; font-family: var(--font-body); transition: all 0.15s; }
.search-cerrar:hover { border-color: var(--color-turquoise); color: var(--color-turquoise); }
.search-vacio { color: var(--text-secondary); padding: var(--space-8); text-align: center; }
.sr-grupo { margin-bottom: var(--space-8); }
.sr-label { font-size: 11px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; color: var(--text-secondary); padding-bottom: var(--space-1); border-bottom: 1px solid #E2E8F0; margin-bottom: var(--space-3); }
.sr-grid { display: grid; gap: var(--space-3); }
.sr-grid--pacientes { grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); }
.sr-grid--consultas { grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); }
.sr-grid--recetas   { grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); }
.sr-card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-3) var(--space-4); text-decoration: none; display: flex; flex-direction: column; gap: 3px; transition: border-color 0.15s, box-shadow 0.15s; }
.sr-card:hover { border-color: var(--color-turquoise); box-shadow: 0 2px 8px rgba(0,200,212,0.08); }
.sr-card-title { font-weight: 700; font-size: 13px; color: var(--text-primary); }
.sr-card-sub { font-size: 11px; color: var(--text-secondary); }

/* Agenda de hoy */
.agenda-hoy { display: flex; flex-direction: column; gap: 2px; border-top: 1px solid rgba(255,255,255,0.08); padding-top: var(--space-3); }
.agenda-header { display: flex; align-items: center; justify-content: space-between; padding: 0 var(--space-2) var(--space-2); }
.agenda-titulo { font-size: 11px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--text-on-dark); opacity: 0.45; }
.agenda-conteo { font-size: 11px; color: var(--color-jade); opacity: 0.8; font-variant-numeric: tabular-nums; }
.agenda-vacio { font-size: 12px; color: var(--text-on-dark); opacity: 0.3; padding: var(--space-2) var(--space-2); font-style: italic; }
.agenda-item { display: flex; align-items: center; gap: var(--space-2); padding: 5px var(--space-2); border-radius: var(--radius-sm); text-decoration: none; transition: background 0.12s; border-left: 2px solid transparent; }
.agenda-item:hover { background: rgba(255,255,255,0.06); }
.agenda-item--issued { border-left-color: #22c55e; }
.agenda-item--voided { border-left-color: #eab308; }
.agenda-item--draft  { border-left-color: #ef4444; }
.agenda-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.dot--green  { background: #22c55e; }
.dot--yellow { background: #eab308; }
.dot--red    { background: #ef4444; }
.dot--blue   { background: #3b82f6; }
.agenda-item--cita { border-left-color: #3b82f6; }
.agenda-hora { font-size: 11px; font-variant-numeric: tabular-nums; color: var(--text-on-dark); opacity: 0.55; min-width: 34px; }
.agenda-paciente { font-size: 12px; color: var(--text-on-dark); opacity: 0.8; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.agenda-item--issued .agenda-paciente { opacity: 0.45; }
.agenda-item--issued .agenda-hora { opacity: 0.35; }
.agenda-item--voided .agenda-paciente { opacity: 0.4; text-decoration: line-through; }
.agenda-item--voided .agenda-hora { opacity: 0.35; }

/* Contenido principal */
.main-content { flex: 1; margin-left: 240px; padding: var(--space-8); min-height: 100vh; }

/* Responsive movil */
@media (max-width: 768px) {
  .topbar { display: flex; }
  .overlay { display: block; }
  .sidebar {
    top: 0; left: -240px;
    transition: transform 0.25s ease;
    transform: translateX(0);
  }
  .sidebar-open { transform: translateX(240px); }
  .sidebar .sidebar-brand { padding-top: 60px; }
  .main-content { margin-left: 0; padding: 72px 16px 24px; }
}
</style>
